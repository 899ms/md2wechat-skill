package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillsWritingReferencesRoundTrip(t *testing.T) {
	oldJSON := skillsReadJSON
	t.Cleanup(func() { skillsReadJSON = oldJSON })
	skillsReadJSON = true
	reader, err := newSkillReader()
	if err != nil {
		t.Fatal(err)
	}
	entries, _, err := reader.ListPath("md2wechat/references/writing")
	if err != nil {
		t.Fatal(err)
	}
	visible := map[string]bool{}
	for _, entry := range entries {
		visible[entry.Path] = !entry.IsDir
	}
	for _, name := range []string{"workflow", "forms", "platforms", "search-strategies", "baidu-baike", "toutiao-baike"} {
		t.Run(name, func(t *testing.T) {
			rel := "references/writing/" + name + ".md"
			if !visible["md2wechat/"+rel] {
				t.Fatalf("not discoverable: %s", rel)
			}
			want, err := os.ReadFile(filepath.Join("..", "..", "skills", "md2wechat", rel))
			if err != nil {
				t.Fatal(err)
			}
			out := captureStdout(t, func() {
				if err := skillsReadCmd.RunE(skillsReadCmd, []string{"md2wechat", rel}); err != nil {
					t.Fatal(err)
				}
			})
			var response struct {
				Success bool   `json:"success"`
				Code    string `json:"code"`
				Data    struct {
					Path    string `json:"path"`
					Content string `json:"content"`
				} `json:"data"`
			}
			if err := json.Unmarshal(out, &response); err != nil {
				t.Fatal(err)
			}
			if !response.Success || response.Code != "SKILL_READ" || response.Data.Path != rel {
				t.Fatalf("unexpected response: %s", out)
			}
			if !bytes.Equal([]byte(response.Data.Content), want) {
				t.Fatalf("embedded content differs from source: %s", rel)
			}
		})
	}
	if _, _, err := reader.ReadReference("md2wechat", "references/writing/not-installed.md"); err == nil {
		t.Fatal("missing writing reference must fail, not fall back silently")
	}
}

func TestSkillsWritingEntrypointsResolve(t *testing.T) {
	reader, err := newSkillReader()
	if err != nil {
		t.Fatal(err)
	}
	route := "md2wechat skills read md2wechat references/writing/workflow.md --json"
	paths := []string{
		filepath.Join("..", "..", "skills", "md2wechat", "SKILL.md"),
		filepath.Join("..", "..", "platforms", "openclaw", "md2wechat", "SKILL.md"),
	}
	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), route) {
			t.Errorf("missing writing route: %s", path)
		}
	}
	workflow, _, err := reader.ReadReference("md2wechat", "references/writing/workflow.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"forms", "platforms", "search-strategies", "baidu-baike", "toutiao-baike"} {
		rel := "references/writing/" + name + ".md"
		if !strings.Contains(string(workflow), "md2wechat skills read md2wechat "+rel+" --json") {
			t.Errorf("workflow cannot reach %s", rel)
		}
		if _, _, err := reader.ReadReference("md2wechat", rel); err != nil {
			t.Fatal(err)
		}
	}
}
