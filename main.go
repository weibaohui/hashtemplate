package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// main 主函数，演示模板引擎的使用
func main() {
	loader := os.DirFS(".")
	eng := New(loader)

	tplStr := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: $(appName)
  namespace: $(namespace ?? "default")
  labels:
    version: $(version ?? "v1.0.0")
spec:
  replicas: #(replicas)
  template:
    spec:
      containers:
        #for c in containers
        - name: $(c.name)
          image: $(c.image):$(c.tag ?? "latest")
          env:
             - name: LOG_LEVEL
               value: $(c.logLevel ?? "info")
          ports:
            #for p in c.ports
            - containerPort: $(p)
            #end
        #end
#if enableIngress
---
kind: Ingress
metadata:
  name: $(appName)-ing
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: $(ingress.rewriteTarget ?? "/")
spec:
  rules:
  - host: $(ingress.host ?? "localhost")
    http: { }
#end
#include "snippet.tpl"`

	// Note: we support both $(...) and ${...} & #( ... ) formats.
	// For convenience, alias $(...) to ${...}
	tplStr = strings.ReplaceAll(tplStr, "$ (", "$(") // no-op guard
	// 使用正则表达式将 $(x) 转换为 ${x}
	re := regexp.MustCompile(`\$\(([^)]+)\)`)
	tplStr = re.ReplaceAllString(tplStr, "${$1}")

	tpl, err := eng.ParseString(tplStr)
	must(err)

	ctx := map[string]any{
		"appName":       "demo-app",
		"replicas":      2,
		"enableIngress": true,
		// 故意省略 namespace 和 version 来测试默认值
		"ingress": map[string]any{
			"host": "demo.example.com",
			// 故意省略 rewriteTarget 来测试默认值
		},
		"containers": []any{
			// 第一个容器有完整信息
			map[string]any{"name": "web", "image": "nginx", "tag": "1.25", "logLevel": "debug", "ports": []int{80, 8080}},
			// 第二个容器故意省略 tag 和 logLevel 来测试默认值
			map[string]any{"name": "sidecar", "image": "busybox", "ports": []int{9000}},
		},
	}

	// Prepare an included snippet file at runtime for the demo
	_ = writeFileIfMissing("snippet.tpl", "# Simple include demo\n#(appName) included!\n")

	out, err := tpl.Render(ctx)
	must(err)

	// Print result
	w := bufio.NewWriter(os.Stdout)
	_, _ = w.WriteString(out)
	_ = w.Flush()

	// 额外测试空安全运算符的各种用法
	fmt.Println("\n=== 空安全运算符测试 ===")
	testTemplate := `
测试结果:
1. 基本 ?? 运算符: ${name ?? "匿名用户"}
2. 嵌套字段: ${user.email ?? "no-email@example.com"}
3. 描述字段: ${description ?? "无描述"}
4. 数字默认值: ${count ?? 0}
5. 空字符串处理: ${emptyField ?? "默认值"}
`

	testTpl, err := eng.ParseString(testTemplate)
	must(err)

	// 测试上下文 - 故意省略一些字段来演示默认值
	testCtx := map[string]any{
		"name": "张三",
		"user": map[string]any{
			// 故意省略 email 字段
		},
		// 故意省略 description, count
		"emptyField": "", // 空字符串测试
	}

	testOut, err := testTpl.Render(testCtx)
	must(err)
	fmt.Print(testOut)
}

// must 错误处理辅助函数
func must(err error) {
	if err != nil {
		panic(err)
	}
}

// writeFileIfMissing 如果文件不存在则创建文件
func writeFileIfMissing(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	return os.WriteFile(path, []byte(content), 0644)
}
