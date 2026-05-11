package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  FieldConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: FieldConfig{
				Name:     "UserId",
				Type:     "int64",
				JSONName: "user_id",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			config: FieldConfig{
				Name:     "",
				Type:     "string",
				JSONName: "test",
			},
			wantErr: true,
			errMsg:  "field name is required",
		},
		{
			name: "empty type",
			config: FieldConfig{
				Name:     "UserId",
				Type:     "",
				JSONName: "user_id",
			},
			wantErr: true,
			errMsg:  "type is required",
		},
		{
			name: "whitespace name",
			config: FieldConfig{
				Name:     "   ",
				Type:     "string",
				JSONName: "test",
			},
			wantErr: true,
			errMsg:  "field name is required",
		},
		{
			name: "whitespace type",
			config: FieldConfig{
				Name:     "UserId",
				Type:     "   ",
				JSONName: "user_id",
			},
			wantErr: true,
			errMsg:  "type is required",
		},
		{
			name: "invalid type",
			config: FieldConfig{
				Name:     "UserId",
				Type:     "invalidType",
				JSONName: "user_id",
			},
			wantErr: true,
			errMsg:  "invalid type",
		},
		{
			name: "valid string type",
			config: FieldConfig{
				Name:     "UserName",
				Type:     "string",
				JSONName: "user_name",
			},
			wantErr: false,
		},
		{
			name: "valid bool type",
			config: FieldConfig{
				Name:     "IsActive",
				Type:     "bool",
				JSONName: "is_active",
			},
			wantErr: false,
		},
		{
			name: "valid float64 type",
			config: FieldConfig{
				Name:     "Amount",
				Type:     "float64",
				JSONName: "amount",
			},
			wantErr: false,
		},
		{
			name: "valid byte type",
			config: FieldConfig{
				Name:     "Data",
				Type:     "byte",
				JSONName: "data",
			},
			wantErr: false,
		},
		{
			name: "valid rune type",
			config: FieldConfig{
				Name:     "Char",
				Type:     "rune",
				JSONName: "char",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() error = nil, wantErr = true")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() error = %v, wantErr = false", err)
				}
			}
		})
	}
}

func TestLoad(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	tests := []struct {
		name      string
		content   string
		wantErr   bool
		errMsg    string
		wantCount int
	}{
		{
			name: "valid yaml",
			content: `- fname: user_id
  type: int64
  mask: true
  comment: 用户ID
- fname: user_name
  type: string
  comment: 用户名
`,
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "invalid snake_case - PascalCase",
			content: `- fname: UserID
  type: int64
`,
			wantErr: true,
			errMsg:  "snake_case",
		},
		{
			name: "invalid snake_case - hyphen",
			content: `- fname: user-id
  type: string
`,
			wantErr: true,
			errMsg:  "snake_case",
		},
		{
			name: "invalid snake_case - double underscore",
			content: `- fname: user__id
  type: string
`,
			wantErr: true,
			errMsg:  "snake_case",
		},
		{
			name: "invalid snake_case - leading underscore",
			content: `- fname: _user_id
  type: string
`,
			wantErr: true,
			errMsg:  "snake_case",
		},
		{
			name: "invalid snake_case - trailing underscore",
			content: `- fname: user_id_
  type: string
`,
			wantErr: true,
			errMsg:  "snake_case",
		},
		{
			name: "invalid snake_case - starts with number",
			content: `- fname: 123_user
  type: string
`,
			wantErr: true,
			errMsg:  "snake_case",
		},
		{
			name: "empty fname",
			content: `- fname: ""
  type: string
`,
			wantErr: true,
			errMsg:  "fname is required",
		},
		{
			name: "invalid type",
			content: `- fname: user_id
  type: invalid_type
`,
			wantErr: true,
			errMsg:  "invalid type",
		},
		{
			name:      "empty yaml",
			content:   `[]`,
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "complex snake_case",
			content: `- fname: http_request_duration_ms
  type: int64
  comment: HTTP请求耗时
- fname: trace_id
  type: string
  comment: 追踪ID
`,
			wantErr:   false,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建临时文件
			tempFile := filepath.Join(tempDir, tt.name+".yaml")
			if err := os.WriteFile(tempFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			fields, err := Load(tempFile)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() error = nil, wantErr = true")
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Load() error = %v, want to contain %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Load() error = %v, wantErr = false", err)
					return
				}
				if len(fields) != tt.wantCount {
					t.Errorf("Load() returned %d fields, want %d", len(fields), tt.wantCount)
				}
			}
		})
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Load() should return error for non-existent file")
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user_id", "UserId"},
		{"user_name", "UserName"},
		{"phone", "Phone"},
		{"http_request_duration", "HttpRequestDuration"},
		{"trace_id", "TraceId"},
		{"a", "A"},
		{"abc_def_ghi", "AbcDefGhi"},
		{"", ""},
		{"single", "Single"},
		{"with_123_numbers", "With123Numbers"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toPascalCase(tt.input)
			if result != tt.expected {
				t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateSnakeCase(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "user_id", false},
		{"valid with numbers", "user123", false},
		{"valid multiple parts", "http_request_duration", false},
		{"valid single letter", "a", false},
		{"valid with number in middle", "user_123_id", false},
		{"empty", "", true},
		{"whitespace", "   ", true},
		{"PascalCase", "UserId", true},
		{"camelCase", "userId", true},
		{"kebab-case", "user-id", true},
		{"double underscore", "user__id", true},
		{"leading underscore", "_user_id", true},
		{"trailing underscore", "user_id_", true},
		{"starts with number", "123_user", true},
		{"contains space", "user id", true},
		{"contains uppercase", "User_ID", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSnakeCase(tt.input)
			if tt.wantErr && err == nil {
				t.Errorf("validateSnakeCase(%q) error = nil, wantErr = true", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateSnakeCase(%q) error = %v, wantErr = false", tt.input, err)
			}
		})
	}
}

func TestLoadFieldConversion(t *testing.T) {
	tempDir := t.TempDir()
	content := `- fname: user_id
  type: int64
  mask: true
  comment: 用户ID
- fname: email_address
  type: string
  mask: true
  comment: 邮箱地址
`
	tempFile := filepath.Join(tempDir, "test.yaml")
	if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	fields, err := Load(tempFile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(fields))
	}

	// 检查第一个字段
	if fields[0].Name != "UserId" {
		t.Errorf("fields[0].Name = %q, want %q", fields[0].Name, "UserId")
	}
	if fields[0].JSONName != "user_id" {
		t.Errorf("fields[0].JSONName = %q, want %q", fields[0].JSONName, "user_id")
	}
	if fields[0].Type != "int64" {
		t.Errorf("fields[0].Type = %q, want %q", fields[0].Type, "int64")
	}
	if !fields[0].Mask {
		t.Errorf("fields[0].Mask = %v, want %v", fields[0].Mask, true)
	}
	if fields[0].Comment != "用户ID" {
		t.Errorf("fields[0].Comment = %q, want %q", fields[0].Comment, "用户ID")
	}

	// 检查第二个字段
	if fields[1].Name != "EmailAddress" {
		t.Errorf("fields[1].Name = %q, want %q", fields[1].Name, "EmailAddress")
	}
	if fields[1].JSONName != "email_address" {
		t.Errorf("fields[1].JSONName = %q, want %q", fields[1].JSONName, "email_address")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
