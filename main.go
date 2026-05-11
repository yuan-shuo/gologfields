// gologfields 是一个代码生成工具，根据 YAML 配置生成结构化日志字段代码
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yuan-shuo/gologfields/internal/config"
	"github.com/yuan-shuo/gologfields/internal/generator"
	"github.com/yuan-shuo/gologfields/internal/maskgen"
)

func main() {
	// 解析命令行参数
	var (
		yamlFile  = flag.String("f", "", "Path to the YAML configuration file (required)")
		outputDir = flag.String("d", "", "Output directory for the generated Go file (required)")
		genMask   = flag.Bool("m", false, "Generate mask functions (default: false)")
	)
	flag.Parse()

	// 验证必选参数
	if *yamlFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -f is required\n")
		flag.Usage()
		os.Exit(1)
	}
	if *outputDir == "" {
		fmt.Fprintf(os.Stderr, "Error: -d is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// 加载配置
	fields, err := config.Load(*yamlFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// 创建生成器
	gen, err := generator.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// 生成代码
	opts := generator.Options{
		OutputDir: *outputDir,
	}
	if err := gen.Generate(fields, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s/logfields_gen.go\n", *outputDir)

	// 生成 mask 函数（仅在指定 -m 时）
	if *genMask {
		maskPath := filepath.Join(*outputDir, "mask.go")

		packageName := filepath.Base(*outputDir)
		if packageName == "." || packageName == "/" || packageName == "" {
			packageName = "logger"
		}

		maskGen := maskgen.New(packageName)
		if err := maskGen.Generate(maskPath, fields); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating mask functions: %v\n", err)
			os.Exit(1)
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
	}
}
