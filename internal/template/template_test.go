package template

import (
	"strings"
	"testing"
	"text/template"
)

func TestToPascal(t *testing.T) {
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
		{"_leading", "Leading"},
		{"trailing_", "Trailing"},
		{"__double__", "Double"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToPascal(tt.input)
			if result != tt.expected {
				t.Errorf("ToPascal(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"UserId", "userid"},
		{"USER_NAME", "user_name"},
		{"Phone", "phone"},
		{"HTTPRequest", "httprequest"},
		{"", ""},
		{"already_lower", "already_lower"},
		{"Mixed123Case", "mixed123case"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToLower(tt.input)
			if result != tt.expected {
				t.Errorf("ToLower(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFuncMap(t *testing.T) {
	fm := FuncMap()

	// 检查函数是否存在
	if _, ok := fm["toPascal"]; !ok {
		t.Error("FuncMap should contain 'toPascal'")
	}
	if _, ok := fm["toLower"]; !ok {
		t.Error("FuncMap should contain 'toLower'")
	}

	// 验证函数可以正确调用
	if fn, ok := fm["toPascal"].(func(string) string); ok {
		result := fn("user_id")
		if result != "UserId" {
			t.Errorf("toPascal function in FuncMap returned %q, want %q", result, "UserId")
		}
	} else {
		t.Error("toPascal in FuncMap is not a func(string) string")
	}

	if fn, ok := fm["toLower"].(func(string) string); ok {
		result := fn("USER_ID")
		if result != "user_id" {
			t.Errorf("toLower function in FuncMap returned %q, want %q", result, "user_id")
		}
	} else {
		t.Error("toLower in FuncMap is not a func(string) string")
	}
}

func TestLogFieldsTemplate(t *testing.T) {
	// 创建模板
	tmpl, err := template.New("test").Funcs(FuncMap()).Parse(LogFieldsTemplate)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	type field struct {
		Name     string
		Type     string
		JSONName string
		Mask     bool
		Comment  string
	}

	tests := []struct {
		name string
		data struct {
			Fields      []field
			PackageName string
		}
		contains []string
	}{
		{
			name: "basic template",
			data: struct {
				Fields      []field
				PackageName string
			}{
				Fields: []field{
					{Name: "UserId", Type: "int64", JSONName: "user_id", Mask: false, Comment: "用户ID"},
				},
				PackageName: "logger",
			},
			contains: []string{
				"package logger",
				"type UserId int64",
				"WUserId",
				"fieldKeys.UserId",
				"用户ID",
			},
		},
		{
			name: "template with mask",
			data: struct {
				Fields      []field
				PackageName string
			}{
				Fields: []field{
					{Name: "Phone", Type: "string", JSONName: "phone", Mask: true, Comment: "手机号"},
				},
				PackageName: "testpkg",
			},
			contains: []string{
				"package testpkg",
				"type Phone string",
				"MaskSensitive",
				"maskPhone",
				"手机号",
			},
		},
		{
			name: "multiple fields",
			data: struct {
				Fields      []field
				PackageName string
			}{
				Fields: []field{
					{Name: "UserId", Type: "int64", JSONName: "user_id", Mask: false, Comment: "用户ID"},
					{Name: "UserName", Type: "string", JSONName: "user_name", Mask: false, Comment: "用户名"},
					{Name: "Phone", Type: "string", JSONName: "phone", Mask: true, Comment: "手机号"},
				},
				PackageName: "logger",
			},
			contains: []string{
				"UserId",
				"UserName",
				"Phone",
				"fieldKeys",
			},
		},
		{
			name: "empty fields",
			data: struct {
				Fields      []field
				PackageName string
			}{
				Fields:      []field{},
				PackageName: "empty",
			},
			contains: []string{
				"package empty",
				"fieldKeys",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			err := tmpl.Execute(&buf, tt.data)
			if err != nil {
				t.Fatalf("template execution failed: %v", err)
			}

			result := buf.String()
			for _, s := range tt.contains {
				if !strings.Contains(result, s) {
					t.Errorf("generated code should contain %q", s)
				}
			}
		})
	}
}

func TestLogFieldsTemplateStructure(t *testing.T) {
	tmpl, err := template.New("test").Funcs(FuncMap()).Parse(LogFieldsTemplate)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	type field struct {
		Name     string
		Type     string
		JSONName string
		Mask     bool
		Comment  string
	}

	data := struct {
		Fields      []field
		PackageName string
	}{
		Fields: []field{
			{Name: "UserId", Type: "int64", JSONName: "user_id", Mask: false, Comment: "用户ID"},
			{Name: "Phone", Type: "string", JSONName: "phone", Mask: true, Comment: "手机号"},
		},
		PackageName: "logger",
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("template execution failed: %v", err)
	}

	result := buf.String()

	// 验证基本结构
	checks := []struct {
		name    string
		pattern string
	}{
		{"package declaration", "package logger"},
		{"import statement", `github.com/zeromicro/go-zero/core/logx`},
		{"fieldKeys struct", "var fieldKeys = struct"},
		{"UserId field", "UserId string"},
		{"Phone field", "Phone string"},
		{"UserId type", "type UserId int64"},
		{"Phone type", "type Phone string"},
		{"WUserId function", "func WUserId(v int64) logx.LogField"},
		{"WPhone function", "func WPhone(v string) logx.LogField"},
		{"MaskSensitive for Phone", "func (v Phone) MaskSensitive() any"},
		{"maskPhone call", "return maskPhone(string(v))"},
	}

	for _, check := range checks {
		if !strings.Contains(result, check.pattern) {
			t.Errorf("generated code should contain %s: %q", check.name, check.pattern)
		}
	}

	// 验证 UserId 没有 MaskSensitive（因为 Mask: false）
	if strings.Contains(result, "func (v UserId) MaskSensitive()") {
		t.Error("UserId should not have MaskSensitive method (Mask: false)")
	}
}

func TestToPascalEdgeCases(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// 边界情况
		{"", ""},
		{"a", "A"},
		{"A", "A"},
		{"_", ""},
		{"__", ""},
		{"___", ""},
		// 多个下划线
		{"a__b", "AB"},
		{"a___b___c", "ABC"},
		// 数字
		{"123", "123"},
		{"a1_b2", "A1B2"},
		// 混合
		{"user_id_v2", "UserIdV2"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToPascal(tt.input)
			if result != tt.expected {
				t.Errorf("ToPascal(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
