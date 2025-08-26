package main

import (
	"os"
	"strings"
	"testing"
)

// TestYAMLCommentBasic 测试基础YAML注释兼容性
func TestYAMLCommentBasic(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	tests := []struct {
		name     string
		template string
		context  map[string]any
		expected string
	}{
		{
			name: "行尾注释与变量插值",
			template: `apiVersion: apps/v1  # API版本注释
kind: Deployment  # 资源类型注释
metadata:
  name: ${appName}  # 应用名称
  namespace: ${namespace ?? "default"}  # 命名空间，默认为default`,
			context: map[string]any{
				"appName": "test-app",
			},
			expected: `apiVersion: apps/v1  # API版本注释
kind: Deployment  # 资源类型注释
metadata:
  name: test-app  # 应用名称
  namespace: default  # 命名空间，默认为default
`,
		},
		{
			name: "行首注释块",
			template: `# YAML文件头部注释
# 这是一个Kubernetes Deployment模板
# 作者: HashTemplate团队
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${appName}
  # 中间注释
  labels:
    app: ${appName}  # 标签注释`,
			context: map[string]any{
				"appName": "my-app",
			},
			expected: `# YAML文件头部注释
# 这是一个Kubernetes Deployment模板
# 作者: HashTemplate团队
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  # 中间注释
  labels:
    app: my-app  # 标签注释
`,
		},
		{
			name: "多级缩进注释",
			template: `apiVersion: v1
kind: ConfigMap
metadata:
  name: ${configName}
data:
  config.yaml: |
    # 应用配置文件
    app:
      name: ${appName}  # 应用名称
      # 端口配置
      port: ${port ?? 8080}  # 默认端口8080`,
			context: map[string]any{
				"configName": "app-config",
				"appName":    "demo-app",
				"port":       9000,
			},
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  config.yaml: |
    # 应用配置文件
    app:
      name: demo-app  # 应用名称
      # 端口配置
      port: 9000  # 默认端口8080
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := eng.ParseString(tt.template)
			if err != nil {
				t.Fatalf("解析模板失败: %v", err)
			}

			result, err := tpl.Render(tt.context)
			if err != nil {
				t.Fatalf("渲染模板失败: %v", err)
			}

			if result != tt.expected {
				t.Errorf("渲染结果不匹配:\n期望:\n%q\n实际:\n%q", tt.expected, result)
			}
		})
	}
}

// TestYAMLCommentEdgeCases 测试边界情况和潜在冲突
func TestYAMLCommentEdgeCases(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	tests := []struct {
		name        string
		template    string
		context     map[string]any
		expected    string
		shouldError bool
	}{
		{
			name: "注释中包含类似指令的文本",
			template: `# 这个注释包含 #for 和 #if 关键字但不应被解析为指令
apiVersion: v1
kind: Pod
metadata:
  name: ${podName}  # 注释：使用 #for 循环和 #if 条件
  annotations:
    description: "这个pod使用了#for和#if语法在注释中"  # 这里有#for
spec:
  containers:
  - name: main`,
			context: map[string]any{
				"podName": "test-pod",
			},
			expected: `# 这个注释包含 #for 和 #if 关键字但不应被解析为指令
apiVersion: v1
kind: Pod
metadata:
  name: test-pod  # 注释：使用 #for 循环和 #if 条件
  annotations:
    description: "这个pod使用了#for和#if语法在注释中"  # 这里有#for
spec:
  containers:
  - name: main
`,
		},
		{
			name: "行中间的#字符",
			template: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${appName}
  annotations:
    # 这是正常注释
    example.com/hash-config: "key1#value1,key2#value2"  # 值中包含#
    example.com/command: "echo 'process #1 is running'"  # 命令中的#
spec:
  replicas: ${replicas}`,
			context: map[string]any{
				"appName":  "hash-app",
				"replicas": 3,
			},
			expected: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: hash-app
  annotations:
    # 这是正常注释
    example.com/hash-config: "key1#value1,key2#value2"  # 值中包含#
    example.com/command: "echo 'process #1 is running'"  # 命令中的#
spec:
  replicas: 3
`,
		},
		{
			name: "空行和空白字符处理",
			template: `apiVersion: v1
kind: Service

# 空行上方和下方

metadata:
  name: ${serviceName}
	# 制表符缩进的注释
    # 空格缩进的注释
spec:
  
  # 包含空白字符的行
  	
  ports:
  - port: ${port}`,
			context: map[string]any{
				"serviceName": "my-service",
				"port":        80,
			},
			expected: `apiVersion: v1
kind: Service

# 空行上方和下方

metadata:
  name: my-service
	# 制表符缩进的注释
    # 空格缩进的注释
spec:
  
  # 包含空白字符的行
  	
  ports:
  - port: 80
`,
		},
		{
			name: "注释与表达式在同一行",
			template: `apiVersion: v1
kind: ConfigMap
metadata:
  name: ${configName}  # 配置名称：${configName ?? "default"}
data:
  config: |
    # 内嵌YAML配置
    app_name: ${appName}  # 应用名：${appName}
    debug: ${debug ?? false}  # 调试模式，默认false`,
			context: map[string]any{
				"configName": "app-config",
				"appName":    "myapp",
				"debug":      true,
			},
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config  # 配置名称：app-config
data:
  config: |
    # 内嵌YAML配置
    app_name: myapp  # 应用名：myapp
    debug: true  # 调试模式，默认false
`,
		},
		{
			name: "多行字符串中的注释",
			template: `apiVersion: v1
kind: ConfigMap
metadata:
  name: ${configName}
data:
  script.sh: |
    #!/bin/bash
    # 这是shell脚本注释
    echo "Starting ${appName}"  # 输出应用名称
    # 更多注释
    if [ "${env}" = "production" ]; then  # 环境检查
      echo "Production mode"
    fi`,
			context: map[string]any{
				"configName": "script-config",
				"appName":    "webapp",
				"env":        "production",
			},
			expected: `apiVersion: v1
kind: ConfigMap
metadata:
  name: script-config
data:
  script.sh: |
    #!/bin/bash
    # 这是shell脚本注释
    echo "Starting webapp"  # 输出应用名称
    # 更多注释
    if [ "production" = "production" ]; then  # 环境检查
      echo "Production mode"
    fi
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := eng.ParseString(tt.template)
			if tt.shouldError {
				if err == nil {
					t.Errorf("期望解析失败，但解析成功了")
				}
				return
			}
			if err != nil {
				t.Fatalf("解析模板失败: %v", err)
			}

			result, err := tpl.Render(tt.context)
			if err != nil {
				t.Fatalf("渲染模板失败: %v", err)
			}

			if result != tt.expected {
				t.Errorf("渲染结果不匹配:\n期望:\n%q\n实际:\n%q", tt.expected, result)
			}
		})
	}
}

// TestYAMLCommentKubernetes 测试完整的Kubernetes实际场景
func TestYAMLCommentKubernetes(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	// 简化的Kubernetes模板
	template := `# Kubernetes部署配置文件
# 包含Deployment资源
# 作者: HashTemplate团队

apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${appName}  # 应用名称
  namespace: ${namespace ?? "default"}  # 命名空间，默认为default
  labels:
    app: ${appName}  # 应用标签
    version: ${version ?? "v1.0.0"}  # 版本标签，默认v1.0.0
  annotations:
    # 部署相关的注释
    deployment.kubernetes.io/revision: "1"
    example.com/deployment-config: "${appName}#${version ?? 'v1.0.0'}"  # 配置信息
spec:
  replicas: ${replicas}  # 副本数量
  selector:
    matchLabels:
      app: ${appName}  # 选择器标签
  template:
    metadata:
      labels:
        app: ${appName}  # Pod标签
        version: ${version ?? "v1.0.0"}  # Pod版本标签
    spec:
      containers:
      - name: ${containerName}  # 容器名称
        image: ${containerImage}:${containerTag ?? "latest"}  # 容器镜像，默认latest标签
        env:
        - name: APP_NAME
          value: ${appName}  # 应用名称环境变量
        - name: APP_VERSION
          value: ${version ?? "v1.0.0"}  # 应用版本环境变量
        ports:
        - containerPort: ${containerPort}  # 容器端口
          name: http  # 端口名称
          protocol: TCP  # 协议`

	context := map[string]any{
		"appName":        "web-application",
		"namespace":      "production",
		"version":        "v2.1.0",
		"replicas":       3,
		"containerName":  "web",
		"containerImage": "nginx",
		"containerTag":   "1.25",
		"containerPort":  80,
	}

	tpl, err := eng.ParseString(template)
	if err != nil {
		t.Fatalf("解析Kubernetes模板失败: %v", err)
	}

	result, err := tpl.Render(context)
	if err != nil {
		t.Fatalf("渲染Kubernetes模板失败: %v", err)
	}

	// 验证关键内容存在
	expectedContents := []string{
		"# Kubernetes部署配置文件",                                     // 文件头注释
		"name: web-application  # 应用名称",                           // 应用名称和注释
		"namespace: production  # 命名空间，默认为default",                // 命名空间和注释
		"version: v2.1.0  # 版本标签，默认v1.0.0",                       // 版本和注释
		"replicas: 3  # 副本数量",                                    // 副本数和注释
		"# 部署相关的注释",                                               // 注释区块
		"name: web  # 容器名称",                                      // 容器名称和注释
		"image: nginx:1.25  # 容器镜像，默认latest标签",                   // 镜像和注释
		"value: web-application  # 应用名称环境变量",                     // 环境变量和注释
		"containerPort: 80  # 容器端口",                              // 端口和注释
		"name: http  # 端口名称",                                     // 端口名称和注释
	}

	for _, expected := range expectedContents {
		if !strings.Contains(result, expected) {
			t.Errorf("Kubernetes模板结果中缺少期望内容: %q", expected)
		}
	}

	// 验证注释不会被误解析为指令
	unexpectedContents := []string{
		"# 容器配置段落\n- name",  // 确保注释后不会直接跟容器配置
		"# 环境变量配置\nvalue",  // 确保注释后不会直接跟环境变量
	}

	for _, unexpected := range unexpectedContents {
		if strings.Contains(result, unexpected) {
			t.Errorf("Kubernetes模板结果中包含不期望的内容（可能是注释解析错误）: %q", unexpected)
		}
	}

	// 验证结果是有效的YAML格式（基本检查）
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			// 这是注释行，应该保持原样
			continue
		}
		if strings.TrimSpace(line) == "" {
			// 空行，正常
			continue
		}
		if strings.Contains(line, "${" ) {
			t.Errorf("第%d行包含未处理的表达式: %q", i+1, line)
		}
	}

	t.Logf("Kubernetes模板渲染成功，总行数: %d", len(lines))
}

// TestYAMLCommentWithDirectives 测试注释与模板指令混合场景
func TestYAMLCommentWithDirectives(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	tests := []struct {
		name     string
		template string
		context  map[string]any
		expected string
	}{
		{
			name: "简单注释与表达式混合",
			template: `# 这是一个Kubernetes Service模板
apiVersion: v1
kind: Service
metadata:
  name: ${serviceName}  # 服务名称
  annotations:
    # 服务相关注释
    example.com/description: "这个服务使用了#for和#if语法"  # 描述注释
spec:
  type: ${serviceType ?? "ClusterIP"}  # 服务类型
  ports:
  - port: ${port}  # 服务端口
    name: http  # 端口名称`,
			context: map[string]any{
				"serviceName": "my-service",
				"serviceType": "LoadBalancer",
				"port":        80,
			},
			expected: `# 这是一个Kubernetes Service模板
apiVersion: v1
kind: Service
metadata:
  name: my-service  # 服务名称
  annotations:
    # 服务相关注释
    example.com/description: "这个服务使用了#for和#if语法"  # 描述注释
spec:
  type: LoadBalancer  # 服务类型
  ports:
  - port: 80  # 服务端口
    name: http  # 端口名称
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := eng.ParseString(tt.template)
			if err != nil {
				t.Fatalf("解析模板失败: %v", err)
			}

			result, err := tpl.Render(tt.context)
			if err != nil {
				t.Fatalf("渲染模板失败: %v", err)
			}

			if result != tt.expected {
				t.Errorf("渲染结果不匹配:\n期望:\n%q\n实际:\n%q", tt.expected, result)
			}
		})
	}
}