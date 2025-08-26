# Hash Template 模板引擎使用教程

## 项目简介

Hash Template 是一个功能强大的 Go 语言模板引擎，专为配置文件生成和文本模板处理而设计。它支持条件语句、循环、表达式求值、空安全运算符等高级功能，特别适用于 Kubernetes 配置文件、配置模板等场景。

## 主要特性

- 🚀 **高性能**: 基于 Go 语言开发，执行效率高
- 🛡️ **空安全**: 内置空安全运算符 `??`，避免空指针异常
- 🔄 **循环支持**: 支持 `#for` 循环，可遍历数组、切片、映射等
- 🎯 **条件判断**: 支持 `#if/#else/#end` 条件语句
- 📁 **文件包含**: 支持 `#include` 指令包含其他模板文件
- 🧮 **表达式求值**: 基于 expr-lang/expr 库，支持复杂表达式计算

## 安装

### 前置要求

- Go 1.24.0 或更高版本

### 获取代码

```bash
git clone https://github.com/weibaohui/hashtemplate.git
cd hashtemplate
go mod tidy
```

### 运行示例

```bash
go run .
```

## 快速开始

### 基本用法

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    // 创建模板引擎
    loader := os.DirFS(".")
    eng := New(loader)
    
    // 定义模板字符串
    templateStr := `
应用名称: ${appName}
版本: ${version ?? "v1.0.0"}
副本数: ${replicas}
`
    
    // 解析模板
    tpl, err := eng.ParseString(templateStr)
    if err != nil {
        panic(err)
    }
    
    // 准备数据上下文
    ctx := map[string]any{
        "appName":  "my-app",
        "replicas": 3,
        // 注意：故意省略 version 来测试默认值
    }
    
    // 渲染模板
    result, err := tpl.Render(ctx)
    if err != nil {
        panic(err)
    }
    
    fmt.Println(result)
}
```

## 语法详解

### 1. 表达式语法


- `${expression}` - 标准表达式语法

```yaml
# 基本变量替换
name: ${appName}
port: ${port}

# 支持嵌套属性访问
image: ${container.image}
tag: ${container.tag}
```

### 2. 空安全运算符 (??)

空安全运算符 `??` 用于提供默认值，当左侧表达式为 `nil` 或空字符串时，返回右侧的默认值。

```yaml
# 基本用法
namespace: ${namespace ?? "default"}
version: ${version ?? "v1.0.0"}

# 嵌套属性的空安全访问
logLevel: ${container.logLevel ?? "info"}
replicas: ${config.replicas ?? 1}
```

### 3. 条件语句

使用 `#if`、`#else`、`#end` 实现条件渲染：

```yaml
#if enableIngress
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ${appName}-ingress
spec:
  rules:
  - host: ${ingress.host ?? "localhost"}
#else
# Ingress 已禁用
#end
```

支持复杂的条件表达式：

```yaml
#if replicas > 1 && enableHA
strategy:
  type: RollingUpdate
#end

#if environment == "production"
resources:
  limits:
    memory: "1Gi"
    cpu: "500m"
#end
```

### 4. 循环语句

#### 遍历数组/切片

```yaml
containers:
#for container in containers
- name: ${container.name}
  image: ${container.image}:${container.tag ?? "latest"}
  ports:
  #for port in container.ports
  - containerPort: ${port}
  #end
#end
```

#### 遍历映射 (key, value 语法)

```yaml
env:
#for key, value in environment
- name: ${key}
  value: ${value}
#end
```

#### 遍历字符串

```yaml
#for char in "hello"
char: ${char}
#end
```

### 5. 文件包含

使用 `#include` 指令包含其他模板文件：

```yaml
# main.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ${appName}-config
data:
#include "config-data.yaml"
```

```yaml
# config-data.yaml
app.properties: |
  app.name=${appName}
  app.version=${version ?? "1.0.0"}
  app.debug=${debug ?? false}
```

## 完整示例

### Kubernetes Deployment 模板

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${appName}
  namespace: ${namespace ?? "default"}
  labels:
    app: ${appName}
    version: ${version ?? "v1.0.0"}
spec:
  replicas: ${replicas}
  selector:
    matchLabels:
      app: ${appName}
  template:
    metadata:
      labels:
        app: ${appName}
        version: ${version ?? "v1.0.0"}
    spec:
      containers:
      #for container in containers
      - name: ${container.name}
        image: ${container.image}:${container.tag ?? "latest"}
        env:
        - name: LOG_LEVEL
          value: ${container.logLevel ?? "info"}
        #if container.env
        #for key, value in container.env
        - name: ${key}
          value: ${value}
        #end
        #end
        ports:
        #for port in container.ports
        - containerPort: ${port}
        #end
        #if container.resources
        resources:
          limits:
            memory: ${container.resources.memory ?? "512Mi"}
            cpu: ${container.resources.cpu ?? "500m"}
        #end
      #end
#if enableIngress
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ${appName}-ingress
  namespace: ${namespace ?? "default"}
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: ${ingress.rewriteTarget ?? "/"}
spec:
  rules:
  - host: ${ingress.host ?? "localhost"}
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: ${appName}
            port:
              number: 80
#end
```

### 对应的数据上下文

```go
ctx := map[string]any{
    "appName":       "demo-app",
    "namespace":     "production",
    "version":       "v2.1.0",
    "replicas":      3,
    "enableIngress": true,
    "ingress": map[string]any{
        "host":          "demo.example.com",
        "rewriteTarget": "/api/v1",
    },
    "containers": []any{
        map[string]any{
            "name":     "web",
            "image":    "nginx",
            "tag":      "1.25",
            "logLevel": "debug",
            "ports":    []int{80, 8080},
            "env": map[string]any{
                "ENVIRONMENT": "production",
                "DEBUG":       "false",
            },
            "resources": map[string]any{
                "memory": "1Gi",
                "cpu":    "1000m",
            },
        },
        map[string]any{
            "name":  "sidecar",
            "image": "busybox",
            "ports": []int{9000},
            // 故意省略某些字段来测试默认值
        },
    },
}
```

## API 参考

### Engine 类型

```go
type Engine struct {
    Loader fs.FS // 文件加载器，用于 #include 指令
}
```

#### 方法

- `New(loader fs.FS) *Engine` - 创建新的模板引擎实例
- `ParseString(s string) (*Template, error)` - 解析字符串模板
- `ParseFile(path string) (*Template, error)` - 解析文件模板

### Template 类型

```go
type Template struct {
    engine *Engine
    nodes  []node
}
```

#### 方法

- `Render(ctx map[string]any) (string, error)` - 渲染模板，返回结果字符串

## 高级功能

### 1. 表达式求值

HashTemplate 基于 [expr-lang/expr](https://github.com/expr-lang/expr) 库，支持丰富的表达式功能：

```yaml
# 数学运算
total: ${price * quantity}
discount: ${total * 0.1}

# 字符串操作
fullName: ${firstName + " " + lastName}
upperName: ${strings.ToUpper(appName)}

# 条件表达式
status: ${replicas > 1 ? "HA" : "Single"}

# 数组操作
firstContainer: ${containers[0].name}
containerCount: ${len(containers)}
```

### 2. 安全的属性访问

HashTemplate 提供了安全的嵌套属性访问，避免空指针异常：

```yaml
# 安全访问嵌套属性
host: ${config.database.host ?? "localhost"}
port: ${config.database.port ?? 5432}

# 安全访问数组元素
firstPort: ${container.ports[0] ?? 8080}
```

### 3. 类型转换

```yaml
# 自动类型转换
replicas: ${string(replicas)}
enabled: ${string(enableFeature)}
```

## 最佳实践

### 1. 模板组织

- 将复杂模板拆分为多个文件
- 使用 `#include` 指令组合模板
- 为模板文件使用有意义的命名

### 2. 数据结构设计

- 使用嵌套的 map 结构组织数据
- 为可选字段提供合理的默认值
- 保持数据结构的一致性

### 3. 错误处理

- 始终检查解析和渲染的错误
- 使用空安全运算符避免运行时错误
- 在开发阶段充分测试模板

### 4. 性能优化

- 重用 Engine 实例
- 缓存解析后的 Template 对象
- 避免在循环中进行复杂计算

## 测试

运行所有测试：

```bash
go test -v
```

运行特定测试：

```bash
go test -v -run TestIntegrationKubernetesTemplate
```

运行基准测试：

```bash
go test -bench=.
```

## 故障排除

### 常见问题

1. **模板解析失败**
   - 检查语法是否正确
   - 确保 `#if`、`#for` 有对应的 `#end`
   - 验证表达式语法

2. **渲染时出错**
   - 检查数据上下文是否包含所需字段
   - 使用空安全运算符提供默认值
   - 验证表达式中的变量名

3. **文件包含失败**
   - 确保文件路径正确
   - 检查文件系统权限
   - 验证 Loader 配置

### 调试技巧

- 使用简单的模板测试基本功能
- 逐步添加复杂性
- 检查中间渲染结果
- 使用单元测试验证模板逻辑

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

本项目采用 MIT 许可证。

## 更多资源

- [expr-lang/expr 文档](https://github.com/expr-lang/expr)
- [Go 模板最佳实践](https://golang.org/pkg/text/template/)
- [Kubernetes 配置文件参考](https://kubernetes.io/docs/reference/)