package maskgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuan-shuo/gologfields/internal/config"
)

func TestNew(t *testing.T) {
	g := New("testpkg")
	if g == nil {
		t.Fatal("New() returned nil")
	}
	if g.packageName != "testpkg" {
		t.Errorf("packageName = %q, want %q", g.packageName, "testpkg")
	}
}

func TestFilterMaskFields(t *testing.T) {
	tests := []struct {
		name     string
		fields   []config.FieldConfig
		expected int
	}{
		{
			name: "all mask fields",
			fields: []config.FieldConfig{
				{Name: "UserId", Type: "int64", Mask: true},
				{Name: "Phone", Type: "string", Mask: true},
			},
			expected: 2,
		},
		{
			name: "mixed fields",
			fields: []config.FieldConfig{
				{Name: "UserId", Type: "int64", Mask: true},
				{Name: "UserName", Type: "string", Mask: false},
				{Name: "Email", Type: "string", Mask: true},
			},
			expected: 2,
		},
		{
			name: "no mask fields",
			fields: []config.FieldConfig{
				{Name: "UserId", Type: "int64", Mask: false},
				{Name: "UserName", Type: "string", Mask: false},
			},
			expected: 0,
		},
		{
			name:     "empty fields",
			fields:   []config.FieldConfig{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterMaskFields(tt.fields)
			if len(result) != tt.expected {
				t.Errorf("filterMaskFields() returned %d fields, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestToLowerFirst(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"UserId", "userId"},
		{"Phone", "phone"},
		{"HTTPRequest", "hTTPRequest"},
		{"A", "a"},
		{"", ""},
		{"alreadyLower", "alreadyLower"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toLowerFirst(tt.input)
			if result != tt.expected {
				t.Errorf("toLowerFirst(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateNewFile(t *testing.T) {
	tempDir := t.TempDir()
	maskFile := filepath.Join(tempDir, "mask.go")

	g := New("testpkg")
	fields := []config.FieldConfig{
		{Name: "UserId", Type: "int64", Mask: true, Comment: "用户ID"},
		{Name: "Phone", Type: "string", Mask: true, Comment: "手机号"},
	}

	err := g.Generate(maskFile, fields)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// 验证文件内容
	content, err := os.ReadFile(maskFile)
	if err != nil {
		t.Fatalf("failed to read mask file: %v", err)
	}

	contentStr := string(content)

	// 检查包声明
	if !strings.Contains(contentStr, "package testpkg") {
		t.Error("mask file should contain package declaration")
	}

	// 检查函数声明
	if !strings.Contains(contentStr, "func maskUserId(userId int64) any") {
		t.Error("mask file should contain maskUserId function")
	}
	if !strings.Contains(contentStr, "func maskPhone(phone string) any") {
		t.Error("mask file should contain maskPhone function")
	}

	// 检查注释
	if !strings.Contains(contentStr, "对用户ID进行脱敏") {
		t.Error("mask file should contain comment for userId")
	}
	if !strings.Contains(contentStr, "对手机号进行脱敏") {
		t.Error("mask file should contain comment for phone")
	}
}

func TestGenerateAppendToExisting(t *testing.T) {
	tempDir := t.TempDir()
	maskFile := filepath.Join(tempDir, "mask.go")

	// 创建已存在的 mask 文件
	existingContent := `package testpkg

// maskUserId 对用户ID进行脱敏
func maskUserId(userId int64) any {
	// 已实现
	return userId
}
`
	if err := os.WriteFile(maskFile, []byte(existingContent), 0644); err != nil {
		t.Fatalf("failed to create existing mask file: %v", err)
	}

	g := New("testpkg")
	fields := []config.FieldConfig{
		{Name: "UserId", Type: "int64", Mask: true, Comment: "用户ID"}, // 已存在
		{Name: "Phone", Type: "string", Mask: true, Comment: "手机号"},  // 新增
	}

	err := g.Generate(maskFile, fields)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// 验证文件内容
	content, err := os.ReadFile(maskFile)
	if err != nil {
		t.Fatalf("failed to read mask file: %v", err)
	}

	contentStr := string(content)

	// 检查原有内容保留
	if !strings.Contains(contentStr, "// 已实现") {
		t.Error("existing implementation should be preserved")
	}

	// 检查新增函数
	if !strings.Contains(contentStr, "func maskPhone(phone string) any") {
		t.Error("new maskPhone function should be added")
	}
}

func TestGenerateNoMaskFields(t *testing.T) {
	tempDir := t.TempDir()
	maskFile := filepath.Join(tempDir, "mask.go")

	g := New("testpkg")
	fields := []config.FieldConfig{
		{Name: "UserId", Type: "int64", Mask: false},
		{Name: "UserName", Type: "string", Mask: false},
	}

	err := g.Generate(maskFile, fields)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// 验证文件未创建
	_, err = os.Stat(maskFile)
	if !os.IsNotExist(err) {
		t.Error("mask file should not be created when no mask fields")
	}
}

func TestGenerateEmptyFields(t *testing.T) {
	tempDir := t.TempDir()
	maskFile := filepath.Join(tempDir, "mask.go")

	g := New("testpkg")
	fields := []config.FieldConfig{}

	err := g.Generate(maskFile, fields)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// 验证文件未创建
	_, err = os.Stat(maskFile)
	if !os.IsNotExist(err) {
		t.Error("mask file should not be created when fields are empty")
	}
}

func TestGenerateWithEmptyComment(t *testing.T) {
	tempDir := t.TempDir()
	maskFile := filepath.Join(tempDir, "mask.go")

	g := New("testpkg")
	fields := []config.FieldConfig{
		{Name: "OrderId", Type: "int64", Mask: true, Comment: ""},
	}

	err := g.Generate(maskFile, fields)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, err := os.ReadFile(maskFile)
	if err != nil {
		t.Fatalf("failed to read mask file: %v", err)
	}

	// 当 Comment 为空时，应该使用 Name 作为描述
	contentStr := string(content)
	if !strings.Contains(contentStr, "对OrderId进行脱敏") {
		t.Error("should use Name as description when Comment is empty")
	}
}

func TestParseExistingMaskFile(t *testing.T) {
	tempDir := t.TempDir()
	maskFile := filepath.Join(tempDir, "mask.go")

	content := `package testpkg

// maskUserId 对用户ID进行脱敏
func maskUserId(userId int64) any {
	return userId
}

// maskPhone 对手机号进行脱敏
func maskPhone(phone string) any {
	return phone
}
`
	if err := os.WriteFile(maskFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create mask file: %v", err)
	}

	g := New("testpkg")
	fileContent, funcs, err := g.parseExistingMaskFile(maskFile)
	if err != nil {
		t.Fatalf("parseExistingMaskFile() error = %v", err)
	}

	if fileContent != content {
		t.Error("file content mismatch")
	}

	if !funcs["maskUserId"] {
		t.Error("maskUserId should be in funcs map")
	}
	if !funcs["maskPhone"] {
		t.Error("maskPhone should be in funcs map")
	}
	if funcs["maskEmail"] {
		t.Error("maskEmail should not be in funcs map")
	}
}

func TestParseExistingMaskFileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	maskFile := filepath.Join(tempDir, "nonexistent.go")

	g := New("testpkg")
	_, _, err := g.parseExistingMaskFile(maskFile)
	if err == nil {
		t.Error("parseExistingMaskFile() should return error for non-existent file")
	}
}

func TestGetMaskFilePath(t *testing.T) {
	tests := []struct {
		outputDir string
		expected  string
	}{
		{"/tmp/output", filepath.Join("/tmp/output", "mask.go")},
		{".", filepath.Join(".", "mask.go")},
		{"logger", filepath.Join("logger", "mask.go")},
	}

	for _, tt := range tests {
		result := GetMaskFilePath(tt.outputDir)
		if result != tt.expected {
			t.Errorf("GetMaskFilePath(%q) = %q, want %q", tt.outputDir, result, tt.expected)
		}
	}
}

func TestGenerateMaskCode(t *testing.T) {
	g := New("testpkg")

	tests := []struct {
		name      string
		fields    []config.FieldConfig
		isNewFile bool
		contains  []string
	}{
		{
			name: "new file with single field",
			fields: []config.FieldConfig{
				{Name: "UserId", Type: "int64", Comment: "用户ID"},
			},
			isNewFile: true,
			contains:  []string{"package testpkg", "func maskUserId(userId int64) any"},
		},
		{
			name: "append to existing file",
			fields: []config.FieldConfig{
				{Name: "Phone", Type: "string", Comment: "手机号"},
			},
			isNewFile: false,
			contains:  []string{"func maskPhone(phone string) any"},
		},
		{
			name: "multiple fields",
			fields: []config.FieldConfig{
				{Name: "UserId", Type: "int64", Comment: "用户ID"},
				{Name: "Email", Type: "string", Comment: "邮箱"},
			},
			isNewFile: true,
			contains:  []string{"maskUserId", "maskEmail"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := g.generateMaskCode(tt.fields, tt.isNewFile)
			for _, s := range tt.contains {
				if !strings.Contains(code, s) {
					t.Errorf("generated code should contain %q", s)
				}
			}
		})
	}
}
