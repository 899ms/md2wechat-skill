package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/geekjourneyx/md2wechat-skill/internal/layoutcatalog"
)

func e2eGate(t *testing.T) {
	t.Helper()
	if os.Getenv("MD2WECHAT_E2E") != "1" {
		t.Skip("set MD2WECHAT_E2E=1 to enable")
	}
	if os.Getenv("MD2WECHAT_BASE_URL") == "" {
		t.Skip("MD2WECHAT_BASE_URL not set")
	}
}

func postConvert(t *testing.T, markdown string) (int, string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"markdown": markdown})
	req, _ := http.NewRequest("POST", os.Getenv("MD2WECHAT_BASE_URL")+"/api/convert", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if k := os.Getenv("MD2WECHAT_API_KEY"); k != "" {
		req.Header.Set("Authorization", "Bearer "+k)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	buf := &bytes.Buffer{}
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp.StatusCode, buf.String()
}

func TestE2EEachModuleAcceptedByAPI(t *testing.T) {
	e2eGate(t)
	c, err := layoutcatalog.DefaultCatalog()
	if err != nil {
		t.Fatal(err)
	}
	for _, mod := range c.ListFiltered(layoutcatalog.ListFilter{}) {
		mod := mod
		t.Run(mod.Name, func(t *testing.T) {
			if mod.Example == "" {
				t.Skip("no example block in spec")
			}
			status, body := postConvert(t, mod.Example)
			if status != 200 {
				t.Errorf("API rejected :::%s example, status=%d body=%s", mod.Name, status, body)
			}
		})
	}
}

func TestE2EOpinionPieceFixture(t *testing.T) {
	e2eGate(t)
	data, err := os.ReadFile("../../internal/layoutcatalog/testdata/integration/opinion-piece.md")
	if err != nil {
		t.Fatal(err)
	}
	status, body := postConvert(t, string(data))
	if status != 200 {
		t.Errorf("API rejected opinion-piece, status=%d body=%s", status, body)
	}
}

func TestE2EValidatorVsAPIConsistency(t *testing.T) {
	e2eGate(t)
	// missing required 'title' — validator should flag it
	bad := ":::hero\neyebrow: only\n:::\n"
	c, err := layoutcatalog.DefaultCatalog()
	if err != nil {
		t.Fatal(err)
	}
	r := c.Validate(bad)
	if len(r.Errors) == 0 {
		t.Fatal("validator should flag missing title")
	}
	status, body := postConvert(t, bad)
	if status == 200 && !strings.Contains(body, "error") {
		t.Logf("WARNING: validator flagged but API accepted — drift detected:\nbody=%s", body)
	}
}
