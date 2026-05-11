// Package config 处理 YAML 配置文件的解析
package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

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

// FieldConfig 表示日志字段配置
type FieldConfig struct {
	Name     string `yaml:"name"`      // 字段名（PascalCase）
	Type     string `yaml:"type"`      // 类型（int64, string, float64等）
	JSONName string `yaml:"json_name"` // JSON字段名（可选，默认转snake_case）
	Mask     bool   `yaml:"mask"`      // 是否需要脱敏
	Comment  string `yaml:"comment"`   // 注释说明
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

	var fields []FieldConfig
	if err := yaml.Unmarshal(data, &fields); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	// 填充默认值并校验
	for i := range fields {
		if fields[i].JSONName == "" {
			fields[i].JSONName = toSnakeCase(fields[i].Name)
		}

		if err := fields[i].Validate(); err != nil {
			return nil, err
		}
	}

	return fields, nil
}

// toSnakeCase 将 PascalCase 转换为 snake_case
func toSnakeCase(s string) string {
	if s == "" {
		return ""
	}

	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}

	// 转换为小写
	for i := range result {
		if result[i] >= 'A' && result[i] <= 'Z' {
			result[i] = result[i] + ('a' - 'A')
		}
	}

	return string(result)
}
