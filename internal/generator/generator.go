// Package generator 处理日志字段代码的生成
package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/yuan-shuo/gologfields/internal/config"
	tmplpkg "github.com/yuan-shuo/gologfields/internal/template"
)

// Options 包含代码生成选项
type Options struct {
	OutputDir string
}

// Generator 负责生成日志字段代码
type Generator struct {
	tmpl *template.Template
}

// New 创建一个新的生成器
func New() (*Generator, error) {
	tmpl, err := template.New("logfields").Funcs(tmplpkg.FuncMap()).Parse(tmplpkg.LogFieldsTemplate)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	return &Generator{tmpl: tmpl}, nil
}

// Generate 根据配置生成代码文件
func (g *Generator) Generate(fields []config.FieldConfig, opts Options) error {
	// 准备模板数据
	data := struct {
		Fields      []config.FieldConfig
		PackageName string
	}{
		Fields:      fields,
		PackageName: getPackageName(opts.OutputDir),
	}

	// 执行模板
	var buf bytes.Buffer
	if err := g.tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	// 格式化代码
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting code: %w\nRaw output:\n%s", err, buf.String())
	}

	// 确保输出目录存在
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory %s: %w", opts.OutputDir, err)
	}

	// 写入文件
	outputPath := filepath.Join(opts.OutputDir, "logfields_gen.go")
	if err := os.WriteFile(outputPath, formatted, 0644); err != nil {
		return fmt.Errorf("writing file %s: %w", outputPath, err)
	}

	return nil
}

// getPackageName 从输出目录路径提取包名
// 例如: "aaa/aaa/bbb" -> "bbb", "." -> "logger"
func getPackageName(outputDir string) string {
	// 统一使用 / 作为分隔符处理
	normalized := strings.ReplaceAll(outputDir, "\\", "/")
	cleanPath := path.Clean(normalized)
	base := path.Base(cleanPath)

	// 如果路径是 "." 或 "/" 等，使用默认包名
	if base == "." || base == "/" || base == "" {
		return "logger"
	}

	return base
}
