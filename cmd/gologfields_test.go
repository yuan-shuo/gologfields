package cmd

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuan-shuo/gologfields/internal/config"
)

// TestParseFlags 测试命令行参数解析
func TestParseFlags(t *testing.T) {
	// 保存原始 os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		errMsg   string
		wantOpts *Options
	}{
		{
			name:    "missing yaml file",
			args:    []string{"cmd", "-d", "output"},
			wantErr: true,
			errMsg:  "-f is required",
		},
		{
			name:    "missing output dir",
			args:    []string{"cmd", "-f", "test.yaml"},
			wantErr: true,
			errMsg:  "-d is required",
		},
		{
			name: "valid args without mask",
			args: []string{"cmd", "-f", "test.yaml", "-d", "output"},
			wantOpts: &Options{
				YamlFile:  "test.yaml",
				OutputDir: "output",
				GenMask:   false,
			},
		},
		{
			name: "valid args with mask",
			args: []string{"cmd", "-f", "test.yaml", "-d", "output", "-m"},
			wantOpts: &Options{
				YamlFile:  "test.yaml",
				OutputDir: "output",
				GenMask:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置 flag 状态
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			os.Args = tt.args
			opts, err := ParseFlags()

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseFlags() expected error, got nil")
				} else if err.Error() != tt.errMsg {
					t.Errorf("ParseFlags() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseFlags() unexpected error = %v", err)
				return
			}

			if opts.YamlFile != tt.wantOpts.YamlFile {
				t.Errorf("YamlFile = %v, want %v", opts.YamlFile, tt.wantOpts.YamlFile)
			}
			if opts.OutputDir != tt.wantOpts.OutputDir {
				t.Errorf("OutputDir = %v, want %v", opts.OutputDir, tt.wantOpts.OutputDir)
			}
			if opts.GenMask != tt.wantOpts.GenMask {
				t.Errorf("GenMask = %v, want %v", opts.GenMask, tt.wantOpts.GenMask)
			}
		})
	}
}

// TestRun 测试 Run 函数
func TestRun(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建测试 YAML 文件（数组格式）
	yamlContent := `- fname: user_id
  type: int64
  comment: 用户ID
- fname: user_name
  type: string
  comment: 用户名
`
	yamlFile := filepath.Join(tempDir, "test.yaml")
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to create test yaml: %v", err)
	}

	outputDir := filepath.Join(tempDir, "output")

	opts := &Options{
		YamlFile:  yamlFile,
		OutputDir: outputDir,
		GenMask:   false,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// 验证文件生成
	genFile := filepath.Join(outputDir, "logfields_gen.go")
	if _, err := os.Stat(genFile); os.IsNotExist(err) {
		t.Error("logfields_gen.go should be generated")
	}
}

// TestRunWithMask 测试带 mask 的 Run
func TestRunWithMask(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 创建测试 YAML 文件（带 mask，数组格式）
	yamlContent := `- fname: phone
  type: string
  comment: 手机号
  mask: true
`
	yamlFile := filepath.Join(tempDir, "test.yaml")
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to create test yaml: %v", err)
	}

	outputDir := filepath.Join(tempDir, "output")

	opts := &Options{
		YamlFile:  yamlFile,
		OutputDir: outputDir,
		GenMask:   true,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// 验证文件生成
	genFile := filepath.Join(outputDir, "logfields_gen.go")
	if _, err := os.Stat(genFile); os.IsNotExist(err) {
		t.Error("logfields_gen.go should be generated")
	}

	// 验证 mask.go 生成
	maskFile := filepath.Join(outputDir, "mask.go")
	if _, err := os.Stat(maskFile); os.IsNotExist(err) {
		t.Error("mask.go should be generated")
	}
}

// TestRunInvalidYaml 测试无效的 YAML 文件
func TestRunInvalidYaml(t *testing.T) {
	opts := &Options{
		YamlFile:  "nonexistent.yaml",
		OutputDir: "output",
		GenMask:   false,
	}

	err := Run(opts)
	if err == nil {
		t.Error("Run() should return error for invalid yaml file")
	}
}

// TestGenerateMaskFunctions 测试 generateMaskFunctions
func TestGenerateMaskFunctions(t *testing.T) {
	tempDir := t.TempDir()

	fields := []config.FieldConfig{
		{Name: "Phone", Type: "string", JSONName: "phone", Mask: true, Comment: "手机号"},
	}

	err := generateMaskFunctions(tempDir, fields)
	if err != nil {
		t.Fatalf("generateMaskFunctions() error = %v", err)
	}

	// 验证 mask.go 生成
	maskFile := filepath.Join(tempDir, "mask.go")
	if _, err := os.Stat(maskFile); os.IsNotExist(err) {
		t.Error("mask.go should be generated")
	}
}

// TestGenerateMaskFunctionsNoMask 测试没有 mask 字段的情况
func TestGenerateMaskFunctionsNoMask(t *testing.T) {
	tempDir := t.TempDir()

	fields := []config.FieldConfig{
		{Name: "UserId", Type: "int64", JSONName: "user_id", Mask: false, Comment: "用户ID"},
	}

	err := generateMaskFunctions(tempDir, fields)
	if err != nil {
		t.Fatalf("generateMaskFunctions() error = %v", err)
	}

	// 验证 mask.go 不会生成（因为没有 mask 字段）
	maskFile := filepath.Join(tempDir, "mask.go")
	if _, err := os.Stat(maskFile); !os.IsNotExist(err) {
		t.Error("mask.go should not be generated when there are no mask fields")
	}
}

// TestRunWithExistingMaskFile 测试已存在 mask 文件的情况
func TestRunWithExistingMaskFile(t *testing.T) {
	tempDir := t.TempDir()

	// 创建测试 YAML 文件
	yamlContent := `- fname: phone
  type: string
  comment: 手机号
  mask: true
- fname: email
  type: string
  comment: 邮箱
  mask: true
`
	yamlFile := filepath.Join(tempDir, "test.yaml")
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to create test yaml: %v", err)
	}

	outputDir := filepath.Join(tempDir, "output")

	// 先创建输出目录和已存在的 mask.go 文件
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	existingMaskContent := `package output

// maskPhone 脱敏手机号
func maskPhone(phone string) any {
	// 已实现
	return phone
}
`
	maskFile := filepath.Join(outputDir, "mask.go")
	if err := os.WriteFile(maskFile, []byte(existingMaskContent), 0644); err != nil {
		t.Fatalf("failed to create existing mask file: %v", err)
	}

	opts := &Options{
		YamlFile:  yamlFile,
		OutputDir: outputDir,
		GenMask:   true,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// 验证 mask.go 仍然存在并包含新函数
	content, err := os.ReadFile(maskFile)
	if err != nil {
		t.Fatalf("failed to read mask file: %v", err)
	}

	contentStr := string(content)
	// 验证已存在的函数仍然保留
	if !strings.Contains(contentStr, "maskPhone") {
		t.Error("existing maskPhone function should be preserved")
	}
}
