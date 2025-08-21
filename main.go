package main

import (
	"bufio"
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
`

	// Note: we support both $(...) and ${...} & #( ... ) formats.
	// For convenience, alias $(...) to ${...}
	tplStr = strings.ReplaceAll(tplStr, "$ (", "$(") // no-op guard
	// 使用正则表达式将 $(x) 转换为 ${x}
	re := regexp.MustCompile(`\$\(([^)]+)\)`)
	tplStr = re.ReplaceAllString(tplStr, "${$1}")

	tpl, err := eng.ParseString(tplStr)
	if err != nil {
		panic(err)
	}

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

	if out, err := tpl.Render(ctx); err == nil {
		// Print result
		w := bufio.NewWriter(os.Stdout)
		_, _ = w.WriteString(out)
		_ = w.Flush()
	}

}
