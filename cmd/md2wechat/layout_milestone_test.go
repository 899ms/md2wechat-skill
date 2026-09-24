package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/geekjourneyx/md2wechat-skill/internal/layoutcatalog"
)

func TestLayoutMilestoneDiscoveryJSON(t *testing.T) {
	oldJSON, oldFilters := jsonOutput, layoutListFilters
	t.Cleanup(func() {
		jsonOutput, layoutListFilters = oldJSON, oldFilters
		layoutcatalog.ResetDefaultCatalogForTests()
	})
	jsonOutput = true
	layoutcatalog.ResetDefaultCatalogForTests()
	counts := buildLayoutCapabilityData()
	for name, want := range map[string]int{
		"recommended_scenario_count": 83, "recommended_syntax_count": 59,
		"compatibility_module_count": 2, "base_enhancement_count": 4, "render_syntax_count": 65,
	} {
		if got := counts[name]; got != want {
			t.Errorf("%s = %v, want %d", name, got, want)
		}
	}
	for _, tt := range []struct {
		lifecycle       string
		count           int
		present, absent []string
	}{
		{"", 59, []string{"cover-reveal", "expand", "gallery"}, []string{"dialogue", "longimage"}},
		{layoutcatalog.LifecycleCompatibility, 2, []string{"dialogue", "longimage"}, []string{"gallery", "expand"}},
	} {
		layoutListFilters.lifecycle = tt.lifecycle
		out := captureStdout(t, func() {
			if err := layoutListCmd.RunE(layoutListCmd, nil); err != nil {
				t.Fatal(err)
			}
		})
		var response struct {
			Success bool   `json:"success"`
			Code    string `json:"code"`
			Data    struct {
				Count   int `json:"count"`
				Modules []struct {
					Name string `json:"name"`
				} `json:"modules"`
			} `json:"data"`
		}
		if err := json.Unmarshal(out, &response); err != nil {
			t.Fatal(err)
		}
		if !response.Success || response.Code != codeLayoutShown || response.Data.Count != tt.count || len(response.Data.Modules) != tt.count {
			t.Fatalf("list envelope/count: %+v", response)
		}
		names := make([]string, 0, len(response.Data.Modules))
		for _, module := range response.Data.Modules {
			names = append(names, module.Name)
		}
		for _, name := range tt.present {
			if !slices.Contains(names, name) {
				t.Errorf("list missing %s", name)
			}
		}
		for _, name := range tt.absent {
			if slices.Contains(names, name) {
				t.Errorf("list unexpectedly includes %s", name)
			}
		}
	}
	for _, tt := range []struct{ name, format string }{{"cover-reveal", layoutcatalog.BodyFormatFields}, {"expand", layoutcatalog.BodyFormatFieldsMarkdown}, {"gallery", layoutcatalog.BodyFormatMarkdownImages}} {
		out := captureStdout(t, func() {
			if err := layoutShowCmd.RunE(layoutShowCmd, []string{tt.name}); err != nil {
				t.Fatal(err)
			}
		})
		var response struct {
			Success bool   `json:"success"`
			Code    string `json:"code"`
			Data    struct {
				Spec struct {
					Name       string `json:"Name"`
					Lifecycle  string `json:"Lifecycle"`
					BodyFormat string `json:"body_format"`
				} `json:"spec"`
			} `json:"data"`
		}
		if err := json.Unmarshal(out, &response); err != nil {
			t.Fatal(err)
		}
		if !response.Success || response.Code != codeLayoutShown || response.Data.Spec.Name != tt.name || response.Data.Spec.Lifecycle != layoutcatalog.LifecycleRecommended || response.Data.Spec.BodyFormat != tt.format {
			t.Fatalf("show %s: %+v", tt.name, response)
		}
	}
}

func TestLayoutMilestoneRenderValidateMatrix(t *testing.T) {
	oldJSON, oldVars, oldParams, oldCaption, oldBodyFile := jsonOutput, layoutRenderVars, layoutRenderParams, layoutRenderCaption, layoutRenderBodyFile
	oldValidateFile, oldValidateStdin, oldExit := layoutValidateFile, layoutValidateStdin, exitFunc
	t.Cleanup(func() {
		jsonOutput, layoutRenderVars, layoutRenderParams, layoutRenderCaption, layoutRenderBodyFile = oldJSON, oldVars, oldParams, oldCaption, oldBodyFile
		layoutValidateFile, layoutValidateStdin, exitFunc = oldValidateFile, oldValidateStdin, oldExit
		layoutcatalog.ResetDefaultCatalogForTests()
	})
	jsonOutput = true
	layoutcatalog.ResetDefaultCatalogForTests()
	tests := []struct {
		name, module, body string
		params             []string
		caption, wantCode  string
	}{
		{"cover default", "cover-reveal", "title: Invitation", nil, "", ""},
		{"cover candidate", "cover-reveal", "title: Invitation", []string{"svg_fallback=first-layer"}, "", ""},
		{"cover strict", "cover-reveal", "title: Invitation", []string{"svg_fallback=first-layer", "wechat_safe_level=strict"}, "", ""},
		{"expand default", "expand", "title: Details\n---\nBody", nil, "", ""},
		{"expand fenced", "expand", "title: Details\n---\n```md\n:::hero\n---\n```\nTail", nil, "", ""},
		{"expand candidate", "expand", "title: Details\n---\nBody", []string{"svg_fallback=first-layer"}, "", ""},
		{"expand strict", "expand", "title: Details\n---\nBody", []string{"svg_fallback=first-layer", "wechat_safe_level=strict"}, "", ""},
		{"gallery old", "gallery", "![One](https://example.com/one.jpg)", nil, "Old gallery", ""},
		{"gallery two", "gallery", "![One](https://example.com/one.jpg)\n![Two](https://example.com/two.jpg)", nil, "", ""},
		{"missing title", "expand", "---\nBody", nil, "", codeLayoutMissingRequiredField},
		{"missing separator", "expand", "title: Details\nBody", nil, "", codeLayoutInvalidFieldValue},
		{"blank body", "expand", "title: Details\n---\n  ", nil, "", codeLayoutInvalidFieldValue},
		{"nested", "expand", "title: Details\n---\n:::hero\ntitle: Nested\n::: ", nil, "", codeLayoutInvalidFieldValue},
		{"invalid enum", "cover-reveal", "title: Invitation", []string{"svg_fallback=unknown"}, "", codeLayoutInvalidFieldValue},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layoutRenderVars, layoutRenderParams, layoutRenderCaption = nil, tt.params, tt.caption
			path := filepath.Join(t.TempDir(), "body.md")
			if err := os.WriteFile(path, []byte(tt.body), 0o600); err != nil {
				t.Fatal(err)
			}
			layoutRenderBodyFile = path
			var runErr error
			out := captureStdout(t, func() { runErr = layoutRenderCmd.RunE(layoutRenderCmd, []string{tt.module}) })
			if tt.wantCode != "" {
				cliErr, ok := runErr.(*cliError)
				if !ok || cliErr.Code != tt.wantCode {
					t.Fatalf("render error = %v, want %s", runErr, tt.wantCode)
				}
				return
			}
			if runErr != nil {
				t.Fatal(runErr)
			}
			var response struct {
				Success bool   `json:"success"`
				Code    string `json:"code"`
				Data    struct {
					Block string `json:"block"`
				} `json:"data"`
			}
			if err := json.Unmarshal(out, &response); err != nil {
				t.Fatal(err)
			}
			if !response.Success || response.Code != codeLayoutRendered || !strings.Contains(response.Data.Block, tt.body) {
				t.Fatalf("rendered block = %+v", response)
			}
			if err := os.WriteFile(path, []byte(response.Data.Block), 0o600); err != nil {
				t.Fatal(err)
			}
			layoutValidateFile, layoutValidateStdin = path, false
			validated := captureStdout(t, func() {
				if err := layoutValidateCmd.RunE(layoutValidateCmd, nil); err != nil {
					t.Fatal(err)
				}
			})
			var result struct {
				Code string `json:"code"`
				Data struct {
					Errors []any `json:"errors"`
				} `json:"data"`
			}
			if err := json.Unmarshal(validated, &result); err != nil {
				t.Fatal(err)
			}
			if result.Code != codeLayoutValidated || len(result.Data.Errors) != 0 {
				t.Fatalf("render/validate mismatch: %+v", result)
			}
		})
	}
}
