// Package config 处理 YAML 配置文件的解析
package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// snakeCaseRegex 匹配有效的 snake_case 格式（OpenTelemetry 规范）
// 规则：小写字母、数字、下划线，不能以数字开头，不能连续下划线，不能以下划线开头或结尾
var snakeCaseRegex = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`)

// 支持的 Go 基础类型
var validTypes = map[string]bool{
	"bool":       true,
	"string":     true,
	"int":        true,
	"int8":       true,
	"int16":      true,
	"int32":      true,
	"int64":      true,
	"uint":       true,
	"uint8":      true,
	"uint16":     true,
	"uint32":     true,
	"uint64":     true,
	"uintptr":    true,
	"byte":       true,
	"rune":       true,
	"float32":    true,
	"float64":    true,
	"complex64":  true,
	"complex128": true,
}

// rawFieldConfig 表示 YAML 原始配置
type rawFieldConfig struct {
	Name    string `yaml:"name"`    // 字段名（snake_case），同时作为 JSON 字段名
	Type    string `yaml:"type"`    // 类型（int64, string, float64等）
	Mask    bool   `yaml:"mask"`    // 是否需要脱敏
	Comment string `yaml:"comment"` // 注释说明
}

// rawConfig 表示 YAML 根配置
type rawConfig struct {
	Service   string           `yaml:"service"`
	LogFields []rawFieldConfig `yaml:"logfields"`
}

// FieldConfig 表示日志字段配置（内部使用，包含生成的字段）
type FieldConfig struct {
	Name     string // 字段名（PascalCase，从 name 自动生成）
	Type     string // 类型
	JSONName string // JSON字段名
	Mask     bool   // 是否需要脱敏
	Comment  string // 注释说明
}

// Validate 校验字段配置是否合法
func (f *FieldConfig) Validate() error {
	// 校验字段名不能为空
	if strings.TrimSpace(f.Name) == "" {
		return fmt.Errorf("field name is required")
	}

	// 校验类型不能为空
	if strings.TrimSpace(f.Type) == "" {
		return fmt.Errorf("field '%s': type is required", f.Name)
	}

	// 校验类型是否为有效的 Go 类型
	if !validTypes[f.Type] {
		return fmt.Errorf("field '%s': invalid type '%s', must be one of valid Go types (e.g., string, int64, float64, bool, etc.)", f.Name, f.Type)
	}

	return nil
}

// Load 从指定路径加载 YAML 配置文件
func Load(path string) ([]FieldConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading YAML file %s: %w", path, err)
	}

	var rawCfg rawConfig
	if err := yaml.Unmarshal(data, &rawCfg); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	// 转换并填充默认值
	fields := make([]FieldConfig, len(rawCfg.LogFields))
	for i, raw := range rawCfg.LogFields {
		// 校验 name 是否符合 snake_case 规范
		if err := validateSnakeCase(raw.Name); err != nil {
			return nil, err
		}

		// 从 name 生成 Name（PascalCase），name 同时作为 JSON 字段名
		fields[i].Name = toPascalCase(raw.Name)
		fields[i].JSONName = raw.Name
		fields[i].Type = raw.Type
		fields[i].Mask = raw.Mask
		fields[i].Comment = raw.Comment

		if err := fields[i].Validate(); err != nil {
			return nil, err
		}
	}

	return fields, nil
}

// validateSnakeCase 校验字符串是否符合 snake_case 规范（OpenTelemetry 命名规范）
// 规则：
//   - 必须以小写字母开头
//   - 只能包含小写字母、数字和下划线
//   - 不能连续出现多个下划线
//   - 不能以下划线开头或结尾
//   - 示例：user_id, http_request_duration, trace_id
func validateSnakeCase(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("name is required")
	}

	if !snakeCaseRegex.MatchString(s) {
		return fmt.Errorf("invalid name '%s': must be snake_case format (lowercase letters, numbers, underscores; start with letter; no consecutive/trailing underscores), see OpenTelemetry naming conventions", s)
	}

	return nil
}

// toPascalCase 将 snake_case 转换为 PascalCase
// 例如：user_id -> UserId, phone_number -> PhoneNumber
func toPascalCase(s string) string {
	if s == "" {
		return ""
	}

	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}
