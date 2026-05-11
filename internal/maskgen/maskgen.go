// Package maskgen 处理 mask 函数的生成和补全
package maskgen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuan-shuo/gologfields/internal/config"
)

// Generator 负责生成和补全 mask 函数
type Generator struct {
	packageName string
}

// New 创建一个新的 mask 生成器
func New(packageName string) *Generator {
	return &Generator{
		packageName: packageName,
	}
}

// Generate 生成或补全 mask 函数
// maskFilePath: mask.go 文件路径
// fields: 需要生成 mask 函数的字段列表
func (g *Generator) Generate(maskFilePath string, fields []config.FieldConfig) error {
	// 过滤出需要 mask 的字段
	maskFields := filterMaskFields(fields)
	if len(maskFields) == 0 {
		return nil // 没有需要 mask 的字段
	}

	// 检查 mask 文件是否已存在
	existingFuncs := make(map[string]bool)
	var existingContent string

	if _, err := os.Stat(maskFilePath); err == nil {
		// 文件存在，解析已实现的函数
		content, funcs, err := g.parseExistingMaskFile(maskFilePath)
		if err != nil {
			return fmt.Errorf("parsing existing mask file: %w", err)
		}
		existingContent = content
		existingFuncs = funcs
	}

	// 找出需要新增的 mask 函数
	var newFields []config.FieldConfig
	for _, field := range maskFields {
		funcName := "mask" + field.Name
		if !existingFuncs[funcName] {
			newFields = append(newFields, field)
		}
	}

	if len(newFields) == 0 {
		return nil // 所有 mask 函数都已实现
	}

	// 生成新的 mask 函数代码
	newCode := g.generateMaskCode(newFields, existingContent == "")

	// 写入文件
	if existingContent == "" {
		// 新文件
		if err := os.WriteFile(maskFilePath, []byte(newCode), 0644); err != nil {
			return fmt.Errorf("writing mask file: %w", err)
		}
	} else {
		// 追加到现有文件
		// 确保文件末尾有换行符，然后追加新代码
		existingContent = strings.TrimRight(existingContent, "\n")
		fullContent := existingContent + "\n\n" + newCode + "\n"
		if err := os.WriteFile(maskFilePath, []byte(fullContent), 0644); err != nil {
			return fmt.Errorf("writing mask file: %w", err)
		}
	}

	return nil
}

// parseExistingMaskFile 解析已存在的 mask 文件，返回内容和已实现的函数名
func (g *Generator) parseExistingMaskFile(path string) (string, map[string]bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, content, parser.ParseComments)
	if err != nil {
		return "", nil, err
	}

	funcs := make(map[string]bool)
	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			funcs[fn.Name.Name] = true
		}
	}

	return string(content), funcs, nil
}

// toLowerFirst 将字符串首字母小写
func toLowerFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// generateMaskCode 生成 mask 函数代码
func (g *Generator) generateMaskCode(fields []config.FieldConfig, isNewFile bool) string {
	var sb strings.Builder

	if isNewFile {
		sb.WriteString(fmt.Sprintf("package %s\n\n", g.packageName))
	}

	for _, field := range fields {
		funcName := "mask" + field.Name
		paramName := toLowerFirst(field.Name)
		desc := field.Comment
		if desc == "" {
			desc = field.Name
		}
		sb.WriteString(fmt.Sprintf("// %s 对%s进行脱敏\n", funcName, desc))
		sb.WriteString("// 请在此实现具体的脱敏逻辑\n")
		sb.WriteString(fmt.Sprintf("func %s(%s %s) any {\n", funcName, paramName, field.Type))
		sb.WriteString(fmt.Sprintf("\t// TODO: 实现%s脱敏逻辑\n", desc))
		sb.WriteString(fmt.Sprintf("\treturn %s\n", paramName))
		sb.WriteString("}\n\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}

// filterMaskFields 过滤出需要 mask 的字段
func filterMaskFields(fields []config.FieldConfig) []config.FieldConfig {
	var result []config.FieldConfig
	for _, field := range fields {
		if field.Mask {
			result = append(result, field)
		}
	}
	return result
}

// GetMaskFilePath 获取 mask 文件的默认路径
func GetMaskFilePath(outputDir string) string {
	return filepath.Join(outputDir, "mask.go")
}
