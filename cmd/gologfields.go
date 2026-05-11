// Package cmd 包含 gologfields 工具的核心逻辑
package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yuan-shuo/gologfields/internal/config"
	"github.com/yuan-shuo/gologfields/internal/generator"
	"github.com/yuan-shuo/gologfields/internal/maskgen"
)

// Options 包含命令行选项
type Options struct {
	YamlFile  string
	OutputDir string
	GenMask   bool
}

// ParseFlags 解析命令行参数
func ParseFlags() (*Options, error) {
	var opts Options

	flag.StringVar(&opts.YamlFile, "f", "", "Path to the YAML configuration file (required)")
	flag.StringVar(&opts.OutputDir, "d", "", "Output directory for the generated Go file (required)")
	flag.BoolVar(&opts.GenMask, "m", false, "Generate mask functions (default: false)")

	flag.Parse()

	// 验证必选参数
	if opts.YamlFile == "" {
		return nil, fmt.Errorf("-f is required")
	}
	if opts.OutputDir == "" {
		return nil, fmt.Errorf("-d is required")
	}

	return &opts, nil
}

// Run 执行代码生成
func Run(opts *Options) error {
	// 加载配置
	fields, err := config.Load(opts.YamlFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// 创建生成器
	gen, err := generator.New()
	if err != nil {
		return fmt.Errorf("creating generator: %w", err)
	}

	// 生成代码
	genOpts := generator.Options{
		OutputDir: opts.OutputDir,
	}
	if err := gen.Generate(fields, genOpts); err != nil {
		return fmt.Errorf("generating code: %w", err)
	}

	fmt.Printf("Generated %s/logfields_gen.go\n", opts.OutputDir)

	// 生成 mask 函数（仅在指定 -m 时）
	if opts.GenMask {
		if err := generateMaskFunctions(opts.OutputDir, fields); err != nil {
			return fmt.Errorf("generating mask functions: %w", err)
		}
	}

	return nil
}

// generateMaskFunctions 生成 mask 函数
func generateMaskFunctions(outputDir string, fields []config.FieldConfig) error {
	maskPath := filepath.Join(outputDir, "mask.go")

	packageName := filepath.Base(outputDir)
	if packageName == "." || packageName == "/" || packageName == "" {
		packageName = "logger"
	}

	maskGen := maskgen.New(packageName)
	if err := maskGen.Generate(maskPath, fields); err != nil {
		return err
	}

	// 检查是否有 mask 字段
	hasMask := false
	for _, f := range fields {
		if f.Mask {
			hasMask = true
			break
		}
	}

	if hasMask {
		fmt.Printf("Generated/Updated %s\n", maskPath)
	}

	return nil
}

// Execute 是主入口函数
func Execute() {
	opts, err := ParseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	if err := Run(opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
