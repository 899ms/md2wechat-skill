package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/geekjourneyx/md2wechat-skill/internal/action"
)

// parseBrandJSON 解析 brand 命令的 JSON 输出
func parseBrandJSON(t *testing.T, output []byte) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("invalid JSON output: %v\nOutput: %s", err, output)
	}
	return result
}

// ============ init group (4 tests) ============

// TestBrandInit_CreatesFile init on empty dir creates brand.yaml, returns BRAND_INITIALIZED
func TestBrandInit_CreatesFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	stdout := captureStdout(t, func() {
		if err := runBrandInit(); err != nil {
			t.Fatalf("runBrandInit() error = %v", err)
		}
	})

	result := parseBrandJSON(t, stdout)

	// 检查 JSON envelope
	if result["success"] != true {
		t.Errorf("expected success=true, got %v", result["success"])
	}
	if result["code"] != "BRAND_INITIALIZED" {
		t.Errorf("expected code=BRAND_INITIALIZED, got %v", result["code"])
	}
	if result["schema_version"] != action.SchemaVersion {
		t.Errorf("expected schema_version=%s, got %v", action.SchemaVersion, result["schema_version"])
	}
	if result["status"] != string(action.StatusCompleted) {
		t.Errorf("expected status=%s, got %v", action.StatusCompleted, result["status"])
	}

	// 检查文件是否创建
	brandPath := filepath.Join(tmpHome, ".config", "md2wechat", "brand.yaml")
	if _, err := os.Stat(brandPath); os.IsNotExist(err) {
		t.Fatalf("brand.yaml not created at %s", brandPath)
	}

	// 检查文件内容不为空
	content, err := os.ReadFile(brandPath)
	if err != nil {
		t.Fatalf("failed to read created brand.yaml: %v", err)
	}
	if len(content) == 0 {
		t.Error("brand.yaml is empty")
	}
}

// TestBrandInit_Idempotent init twice, second call still returns BRAND_INITIALIZED (not error), file not overwritten
func TestBrandInit_Idempotent(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// 第一次 init
	stdout1 := captureStdout(t, func() {
		if err := runBrandInit(); err != nil {
			t.Fatalf("first runBrandInit() error = %v", err)
		}
	})

	result1 := parseBrandJSON(t, stdout1)
	if result1["code"] != "BRAND_INITIALIZED" {
		t.Errorf("first init: expected code=BRAND_INITIALIZED, got %v", result1["code"])
	}

	brandPath := filepath.Join(tmpHome, ".config", "md2wechat", "brand.yaml")

	// 修改文件内容，标记一下
	testContent := "# MODIFIED BY TEST\nschema_version: 1\nname: Test User\n"
	if err := os.WriteFile(brandPath, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to modify brand.yaml: %v", err)
	}

	// 第二次 init
	stdout2 := captureStdout(t, func() {
		if err := runBrandInit(); err != nil {
			t.Fatalf("second runBrandInit() error = %v", err)
		}
	})

	result2 := parseBrandJSON(t, stdout2)
	if result2["success"] != true {
		t.Errorf("second init: expected success=true, got %v", result2["success"])
	}
	if result2["code"] != "BRAND_INITIALIZED" {
		t.Errorf("second init: expected code=BRAND_INITIALIZED, got %v", result2["code"])
	}

	// 检查文件没有被覆盖
	content, err := os.ReadFile(brandPath)
	if err != nil {
		t.Fatalf("failed to read brand.yaml after second init: %v", err)
	}
	if string(content) != testContent {
		t.Error("brand.yaml was overwritten by second init (should be idempotent)")
	}
}

// TestBrandInit_JSONEnvelope output is valid JSON with schema_version:"v1", success:true, status:"completed"
func TestBrandInit_JSONEnvelope(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	stdout := captureStdout(t, func() {
		if err := runBrandInit(); err != nil {
			t.Fatalf("runBrandInit() error = %v", err)
		}
	})

	result := parseBrandJSON(t, stdout)

	// 检查 JSON envelope 契约
	if result["schema_version"] != "v1" {
		t.Errorf("expected schema_version=v1, got %v", result["schema_version"])
	}
	if result["success"] != true {
		t.Errorf("expected success=true, got %v", result["success"])
	}
	if result["status"] != "completed" {
		t.Errorf("expected status=completed, got %v", result["status"])
	}

	// 检查必须的字段存在
	if _, ok := result["code"]; !ok {
		t.Error("missing 'code' field in JSON envelope")
	}
	if _, ok := result["message"]; !ok {
		t.Error("missing 'message' field in JSON envelope")
	}
}

// TestBrandInit_CreatesParentDir init when ~/.config/md2wechat/ doesn't exist, creates dir + file
func TestBrandInit_CreatesParentDir(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// 确保父目录不存在
	configDir := filepath.Join(tmpHome, ".config", "md2wechat")
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Fatalf("config dir already exists (test precondition failed)")
	}

	stdout := captureStdout(t, func() {
		if err := runBrandInit(); err != nil {
			t.Fatalf("runBrandInit() error = %v", err)
		}
	})

	result := parseBrandJSON(t, stdout)
	if result["code"] != "BRAND_INITIALIZED" {
		t.Errorf("expected code=BRAND_INITIALIZED, got %v", result["code"])
	}

	// 检查父目录和文件都被创建
	if info, err := os.Stat(configDir); err != nil {
		t.Fatalf("config dir was not created: %v", err)
	} else if !info.IsDir() {
		t.Fatal("config path exists but is not a directory")
	}

	brandPath := filepath.Join(configDir, "brand.yaml")
	if _, err := os.Stat(brandPath); os.IsNotExist(err) {
		t.Fatalf("brand.yaml was not created at %s", brandPath)
	}
}

// ============ show group (5 tests) ============

// TestBrandShow_NotFound show when no file → BRAND_NOT_FOUND, success:false, status:"action_required"
func TestBrandShow_NotFound(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// 确保文件不存在
	brandPath := filepath.Join(tmpHome, ".config", "md2wechat", "brand.yaml")
	if _, err := os.Stat(brandPath); !os.IsNotExist(err) {
		t.Fatalf("brand.yaml exists (test precondition failed)")
	}

	stdout := captureStdout(t, func() {
		if err := runBrandShow(); err != nil {
			t.Fatalf("runBrandShow() error = %v", err)
		}
	})

	result := parseBrandJSON(t, stdout)

	// 检查返回值
	if result["success"] != false {
		t.Errorf("expected success=false, got %v", result["success"])
	}
	if result["code"] != "BRAND_NOT_FOUND" {
		t.Errorf("expected code=BRAND_NOT_FOUND, got %v", result["code"])
	}
	if result["status"] != "action_required" {
		t.Errorf("expected status=action_required, got %v", result["status"])
	}
}

// TestBrandShow_ValidFile show after init → BRAND_SHOWN, success:true, data.profile present, data.path uses ~/
func TestBrandShow_ValidFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// 先 init
	captureStdout(t, func() {
		if err := runBrandInit(); err != nil {
			t.Fatalf("runBrandInit() error = %v", err)
		}
	})

	// 再 show
	stdout := captureStdout(t, func() {
		if err := runBrandShow(); err != nil {
			t.Fatalf("runBrandShow() error = %v", err)
		}
	})

	result := parseBrandJSON(t, stdout)

	// 检查返回值
	if result["success"] != true {
		t.Errorf("expected success=true, got %v", result["success"])
	}
	if result["code"] != "BRAND_SHOWN" {
		t.Errorf("expected code=BRAND_SHOWN, got %v", result["code"])
	}

	// 检查 data 字段
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data field is not a map: %T", result["data"])
	}

	// 检查 profile 存在
	if _, ok := data["profile"]; !ok {
		t.Error("data.profile is missing")
	}

	// 检查 path 使用 ~/ 格式
	path, ok := data["path"].(string)
	if !ok {
		t.Fatalf("data.path is not a string: %T", data["path"])
	}
	if len(path) == 0 {
		t.Error("data.path is empty")
	}
	// path 应该以 ~/ 开头（normalizeBrandPath 的效果）
	if path[0] != '~' && path[0] != '/' {
		t.Errorf("data.path should be normalized to use ~/ or absolute path, got: %s", path)
	}
}

// TestBrandShow_CorruptYAML show with invalid YAML content → BRAND_READ_FAILED, success:false, status:"failed"
func TestBrandShow_CorruptYAML(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// 创建目录和一个无效的 YAML 文件
	configDir := filepath.Join(tmpHome, ".config", "md2wechat")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	brandPath := filepath.Join(configDir, "brand.yaml")
	corruptContent := "this is not valid YAML:\n  - unclosed bracket: [\n  invalid: {unclosed"
	if err := os.WriteFile(brandPath, []byte(corruptContent), 0644); err != nil {
		t.Fatalf("failed to write corrupt brand.yaml: %v", err)
	}

	stdout := captureStdout(t, func() {
		if err := runBrandShow(); err != nil {
			t.Fatalf("runBrandShow() error = %v", err)
		}
	})

	result := parseBrandJSON(t, stdout)

	// 检查返回值
	if result["success"] != false {
		t.Errorf("expected success=false, got %v", result["success"])
	}
	if result["code"] != "BRAND_READ_FAILED" {
		t.Errorf("expected code=BRAND_READ_FAILED, got %v", result["code"])
	}
	if result["status"] != "failed" {
		t.Errorf("expected status=failed, got %v", result["status"])
	}

	// 检查 error 字段存在
	if _, ok := result["error"]; !ok {
		t.Error("expected 'error' field in response for corrupt YAML")
	}
}

// TestBrandShow_PartialProfile show with partial brand.yaml (only name set) → BRAND_SHOWN with profile, other fields zero-value
func TestBrandShow_PartialProfile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// 创建一个只有部分字段的 brand.yaml
	configDir := filepath.Join(tmpHome, ".config", "md2wechat")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	brandPath := filepath.Join(configDir, "brand.yaml")
	partialContent := `schema_version: 1
name: "Test Author"
`
	if err := os.WriteFile(brandPath, []byte(partialContent), 0644); err != nil {
		t.Fatalf("failed to write partial brand.yaml: %v", err)
	}

	stdout := captureStdout(t, func() {
		if err := runBrandShow(); err != nil {
			t.Fatalf("runBrandShow() error = %v", err)
		}
	})

	result := parseBrandJSON(t, stdout)

	// 检查返回值
	if result["success"] != true {
		t.Errorf("expected success=true, got %v", result["success"])
	}
	if result["code"] != "BRAND_SHOWN" {
		t.Errorf("expected code=BRAND_SHOWN, got %v", result["code"])
	}

	// 检查 profile 存在且包含设置的字段
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data field is not a map: %T", result["data"])
	}

	profile, ok := data["profile"].(map[string]interface{})
	if !ok {
		t.Fatalf("data.profile is not a map: %T", data["profile"])
	}

	// 检查 name 字段正确
	if profile["name"] != "Test Author" {
		t.Errorf("expected profile.name='Test Author', got %v", profile["name"])
	}

	// 检查 schema_version 存在
	if profile["schema_version"] == nil {
		t.Error("profile.schema_version is missing")
	}

	// 其他字段应该是零值或不存在（YAML 解析行为）
	// 这是正常的，不应该报错
}

// TestBrandShow_JSONEnvelope valid JSON, schema_version:"v1", code:"BRAND_SHOWN", data.path non-empty
func TestBrandShow_JSONEnvelope(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// 先 init
	captureStdout(t, func() {
		if err := runBrandInit(); err != nil {
			t.Fatalf("runBrandInit() error = %v", err)
		}
	})

	// 再 show
	stdout := captureStdout(t, func() {
		if err := runBrandShow(); err != nil {
			t.Fatalf("runBrandShow() error = %v", err)
		}
	})

	result := parseBrandJSON(t, stdout)

	// 检查 JSON envelope 契约
	if result["schema_version"] != "v1" {
		t.Errorf("expected schema_version=v1, got %v", result["schema_version"])
	}
	if result["code"] != "BRAND_SHOWN" {
		t.Errorf("expected code=BRAND_SHOWN, got %v", result["code"])
	}

	// 检查 data.path 非空
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data field is not a map: %T", result["data"])
	}

	path, ok := data["path"].(string)
	if !ok {
		t.Fatalf("data.path is not a string: %T", data["path"])
	}
	if len(path) == 0 {
		t.Error("data.path is empty")
	}
}
