package layoutcatalog

import (
	"errors"
	"strings"
	"testing"
)

func testExpandCatalog() *Catalog {
	return NewCatalogWithSpec(&LayoutSpec{
		Name: "expand", BodyFormat: BodyFormatFieldsMarkdown,
		Fields: &FieldsSpec{Required: []FieldSpec{{Name: "title"}}, Optional: []FieldSpec{{Name: "summary"}, {Name: "label"}}},
		Body:   &BodySpec{Separator: "---", MinItems: 1},
	})
}

func TestExpandBodyValidationMatrix(t *testing.T) {
	c := testExpandCatalog()
	tests := []struct {
		name, body, field string
		valid             bool
	}{
		{"plain", "title: Details\n---\nBody", "", true},
		{"paragraphs", "title: Details\nsummary: Short\n---\nFirst\n\nSecond", "", true},
		{"fenced delimiters", "title: Details\n---\n```md\n:::hero\n---\n```\nTail", "", true},
		{"fenced close", "title: Details\n---\n~~~md\n:::\n~~~\nTail", "", true},
		{"crlf", "title: Details\r\n---\r\nBody", "", true},
		{"second separator", "title: Details\n---\nBody\n---\nTail", "", true},
		{"unfinished body fence", "title: Details\n---\n```md\nBody", "", false},
		{"missing separator", "title: Details\nBody", "", false},
		{"blank title", "title: \n---\nBody", "title", false},
		{"blank body", "title: Details\n---\n  ", "", false},
		{"nested directive", "title: Details\n---\n:::hero\ntitle: Nested\n:::", "", false},
		{"nested directive punctuation", "title: Details\n---\n:::hero!", "", false},
		{"fenced directive punctuation", "title: Details\n---\n```md\n:::hero!\n```", "", true},
		{"invalid header", "title: Details\nunknown: Bad\n---\nBody", "unknown", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := c.Validate(":::expand\n" + tt.body + "\n:::\n")
			if tt.valid && len(report.Errors) != 0 {
				t.Fatalf("errors = %+v", report.Errors)
			}
			if !tt.valid && len(report.Errors) == 0 {
				t.Fatal("expected validation error")
			}
			if !tt.valid && (report.Errors[0].Module != "expand" || report.Errors[0].Line != 1 || report.Errors[0].Field != tt.field) {
				t.Fatalf("error = %+v", report.Errors[0])
			}
		})
	}
}

func TestExpandRenderBodyRoundTrip(t *testing.T) {
	c := testExpandCatalog()
	body := "title: Details\n---\n```md\n:::hero\n---\n```\nLast line"
	out, err := c.RenderBlock("expand", RenderInput{Body: body})
	if err != nil {
		t.Fatal(err)
	}
	if out != ":::expand\n"+body+"\n:::\n" {
		t.Fatalf("round trip changed body:\n%s", out)
	}
	if report := c.Validate(out); len(report.Errors) != 0 {
		t.Fatalf("rendered block invalid: %+v", report.Errors)
	}
	for _, tt := range []struct {
		body string
		want error
	}{
		{"---\nBody", ErrMissingRequiredField},
		{"title: Details\nBody", ErrInvalidFieldValue},
		{"title: Details\n---\n", ErrInvalidFieldValue},
	} {
		_, err := c.RenderBlock("expand", RenderInput{Body: tt.body})
		if !errors.Is(err, tt.want) {
			t.Errorf("body %q: error %v, want %v", tt.body, err, tt.want)
		}
	}
}

func TestLayoutFenceScannerIgnoresCodeFences(t *testing.T) {
	c := testExpandCatalog()
	markdown := "```md\n:::hero\n:::\n```\n\n:::expand\ntitle: Details\n---\n```md\n:::hero\n:::\n```\nEnd\n:::"
	report := c.Validate(markdown)
	if len(report.Errors) != 0 || len(report.Warnings) != 0 {
		t.Fatalf("code fences parsed as modules: %+v", report)
	}
	if !strings.Contains(markdown, "End\n:::") {
		t.Fatal("test fixture lacks closing fence")
	}
}

func TestFenceDelimiterSupportMatchesRendererByModule(t *testing.T) {
	c := NewCatalog()
	if err := c.Load(); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name, module, markdown string
		valid                  bool
	}{
		{"expand keeps fenced close", "expand", ":::expand\ntitle: Details\n---\n```md\n:::\n```\nBody\n:::", true},
		{"split ordinary body", "split", ":::split\nLeft\n---\nRight\n:::", true},
		{"split cannot hide close in code", "split", ":::split\n```md\n:::\n```\n---\nRight side\n:::", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			report := c.Validate(tt.markdown)
			if (len(report.Errors) == 0) != tt.valid {
				t.Fatalf("validation errors = %+v, want valid=%v", report.Errors, tt.valid)
			}
			if !tt.valid {
				body := "```md\n:::\n```\n---\nRight side"
				if _, err := c.RenderBlock(tt.module, RenderInput{Body: body}); err == nil {
					t.Fatal("render accepted a body the API truncates")
				}
			}
		})
	}
}

func TestExpandWitnessWithFencedDirective(t *testing.T) {
	c := testExpandCatalog()
	example := ":::expand\ntitle: Details\n---\n```md\n:::hero\n:::\n```\nTail\n:::"
	if err := c.ValidateWitness(WitnessContract{Module: "expand", Example: example, AssertContains: "Tail"}); err != nil {
		t.Fatal(err)
	}
}
