# gologfields

[![CI](https://github.com/yuan-shuo/gologfields/workflows/ci/badge.svg)](https://github.com/yuan-shuo/gologfields/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/yuan-shuo/gologfields)](https://goreportcard.com/report/github.com/yuan-shuo/gologfields)
[![codecov](https://codecov.io/gh/yuan-shuo/gologfields/branch/main/graph/badge.svg)](https://codecov.io/gh/yuan-shuo/gologfields)
[![Release](https://img.shields.io/github/release/yuan-shuo/gologfields.svg)](https://github.com/yuan-shuo/gologfields/releases/latest)
[![Go Version](https://img.shields.io/badge/go%20version-%3E=1.25-61CFDD.svg)](https://golang.org/)

一个 Go 代码生成工具，根据 YAML 配置生成结构化日志字段代码，支持自动脱敏功能。

## 功能特性

- **类型安全**：为每个日志字段生成专门的类型
- **IDE 友好**：通过 `W` 前缀的构造器函数，IDE 自动补全一目了然
- **自动脱敏**：支持敏感数据自动脱敏，符合 `go-zero` 的 `Sensitive` 接口
- **自动补全 mask 函数**：自动生成未实现的脱敏函数存根，避免手动查看类型
- **灵活配置**：通过 YAML 文件配置字段，支持自定义 JSON 字段名
- **类型校验**：自动校验 YAML 中的类型是否为有效的 Go 类型

## 安装

```bash
go install github.com/yuan-shuo/gologfields@latest
```

或者本地编译：

```bash
git clone https://github.com/yuan-shuo/gologfields.git
cd gologfields
go build -o gologfields .
```

## 使用方法

### 1. 创建 YAML 配置文件

```yaml
# logfields.yaml
- fname: user_id
  type: int64
  mask: true
  comment: 用户ID

- fname: user_name
  type: string
  comment: 用户名

- fname: phone
  type: string
  mask: true
  comment: 手机号

- fname: email
  type: string
  mask: true
  comment: 邮箱
```

**字段说明：**
- `fname`: 日志字段名（snake_case），会自动转换为 PascalCase 作为 Go 类型名
- `type`: Go 类型（string, int64, float64, bool 等）
- `mask`: 是否需要脱敏（可选，默认 false）
- `comment`: 字段注释说明（可选）

**snake_case 命名规范（OpenTelemetry）：**
- 必须以小写字母开头
- 只能包含小写字母、数字和下划线
- 不能连续出现多个下划线
- 不能以下划线开头或结尾
- 有效示例：`user_id`, `http_request_duration`, `trace_id`
- 无效示例：`UserID`, `user-id`, `user__id`, `_user_id`, `user_id_`, `123_user`

### 2. 运行生成工具

```bash
# 仅生成日志字段代码
gologfields -f logfields.yaml -d ./logger

# 同时生成日志字段和 mask 函数存根
gologfields -f logfields.yaml -d ./logger -m
```

参数说明：
- `-f`: YAML 配置文件路径（必填）
- `-d`: 输出目录路径（必填）
- `-m`: 生成/追加 mask 函数存根（可选，默认不生成）

### 3. 实现脱敏函数（仅对 mask: true 的字段）

如果需要脱敏功能，使用 `-m` 参数运行工具，会自动生成 `mask.go` 文件：

```go
package logger

// maskUserId 对用户ID进行脱敏
// 请在此实现具体的脱敏逻辑
func maskUserId(userId int64) any {
    // TODO: 实现用户ID脱敏逻辑
    return userId
}

// maskPhone 对手机号进行脱敏
// 请在此实现具体的脱敏逻辑
func maskPhone(phone string) any {
    // TODO: 实现手机号脱敏逻辑
    return phone
}
```

你只需要填充具体的脱敏逻辑即可。当你添加新的 mask 字段后，再次使用 `-m` 运行工具，会自动追加新的函数存根，不会覆盖已有的实现。

**示例：实现脱敏逻辑**

```go
package logger

import "strconv"

// maskUserId 对用户ID进行脱敏
// 脱敏规则：显示前2位和后2位，中间用 **** 替换
func maskUserId(userId int64) any {
    s := strconv.FormatInt(userId, 10)
    if len(s) <= 4 {
        return "****"
    }
    return s[:2] + "****" + s[len(s)-2:]
}

// maskPhone 对手机号进行脱敏
// 脱敏规则：显示前3位和后4位，中间用 **** 替换
func maskPhone(phone string) any {
    if len(phone) < 7 {
        return "****"
    }
    return phone[:3] + "****" + phone[len(phone)-4:]
}
```

### 4. 使用生成的代码

```go
package main

import (
    "context"
    "yourproject/logger"
    "github.com/zeromicro/go-zero/core/logx"
)

func main() {
    ctx := context.Background()
    log := logx.WithContext(ctx)

    // 结构化日志，自动脱敏
    log.Infow("用户登录",
        logger.WUserId(12345678),           // 输出: "user_id": "12****78"
        logger.WPhone("13812345678"),       // 输出: "phone": "138****5678"
        logger.WUserName("张三"),            // 输出: "user_name": "张三"
    )
}
```

## YAML 配置说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `fname` | string | 是 | 字段名（snake_case），如 `user_id`, `phone_number` |
| `type` | string | 是 | Go 类型，如 `string`, `int64`, `float64`, `bool` 等 |
| `mask` | bool | 否 | 是否需要脱敏，默认为 `false` |
| `comment` | string | 否 | 字段注释说明 |

**命名转换规则：**
- `fname: user_id` → Go 类型名: `UserId`, JSON 字段名: `user_id`
- `fname: phone_number` → Go 类型名: `PhoneNumber`, JSON 字段名: `phone_number`

### 支持的 Go 类型

- 布尔型：`bool`
- 字符串：`string`
- 有符号整数：`int`, `int8`, `int16`, `int32`, `int64`
- 无符号整数：`uint`, `uint8`, `uint16`, `uint32`, `uint64`, `uintptr`
- 别名类型：`byte`, `rune`
- 浮点数：`float32`, `float64`
- 复数：`complex64`, `complex128`

## 生成的代码结构

```go
package logger

import "github.com/zeromicro/go-zero/core/logx"

// fieldKeys 日志字段名常量
var fieldKeys = struct {
    UserID   string
    UserName string
    Phone    string
}{
    UserID:   "user_id",
    UserName: "user_name",
    Phone:    "phone",
}

// UserID 用户ID
type UserID int64

// WUserID 创建用户ID日志字段
func WUserID(v int64) logx.LogField {
    return logx.Field(fieldKeys.UserID, UserID(v))
}

// MaskSensitive 实现 Sensitive 接口
func (v UserID) MaskSensitive() any {
    return maskUserID(int64(v))
}

// ... 其他字段
```

## 命名约定

1. **类型定义**：使用业务语义的命名（如 `UserID`, `Phone`, `Email`）
2. **字段构造器**：使用 `W` + 类型名的命名格式（`W` 代表 WithField）
   - 例如：`WUserID()`, `WPhone()`, `WEmail()`
3. **字段名常量**：在 `fieldKeys` 结构体中定义，与类型名保持一致，写法为 snake_case
   - 例如：`fieldKeys.UserID = "user_id"`

## 项目结构

```
gologfields/
├── main.go                          # 主程序入口
├── internal/
│   ├── config/
│   │   └── config.go                # YAML 配置解析和校验
│   ├── template/
│   │   └── template.go              # 代码生成模板
│   └── generator/
│       └── generator.go             # 代码生成器
├── logfields.example.yaml           # 示例 YAML 配置
├── README.md                        # 本文档
├── .gitignore                       # Git 忽略文件
└── go.mod                           # Go 模块定义
```

## 示例

查看 [logfields.example.yaml](logfields.example.yaml) 获取完整的配置示例。