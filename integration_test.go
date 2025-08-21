package main

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// TestIntegrationKubernetesTemplate 测试完整的Kubernetes模板集成
func TestIntegrationKubernetesTemplate(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	// 创建测试用的snippet文件
	err := os.WriteFile("test_snippet.tpl", []byte("# Snippet included\napp: ${appName}"), 0644)
	if err != nil {
		t.Fatalf("创建测试snippet文件失败: %v", err)
	}
	defer os.Remove("test_snippet.tpl")

	// 完整的Kubernetes模板，包含所有语法特性
	templateStr := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${appName}
  namespace: ${namespace ?? "default"}
  labels:
    version: ${version ?? "v1.0.0"}
spec:
  replicas: #(replicas)
  template:
    spec:
      containers:
        #for c in containers
        - name: ${c.name}
          image: ${c.image}:${c.tag ?? "latest"}
          env:
             - name: LOG_LEVEL
               value: ${c.logLevel ?? "info"}
          ports:
            #for p in c.ports
            - containerPort: ${p}
            #end
        #end
#if enableIngress
---
kind: Ingress
metadata:
  name: ${appName}-ing
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: ${ingress.rewriteTarget ?? "/"}
spec:
  rules:
  - host: ${ingress.host ?? "localhost"}
    http: { }
#end
#include "test_snippet.tpl"`

	// 转换$(...)为${...}格式
	templateStr = strings.ReplaceAll(templateStr, "$ (", "$(")
	re := regexp.MustCompile(`\$\(([^)]+)\)`)
	templateStr = re.ReplaceAllString(templateStr, "${$1}")

	tpl, err := eng.ParseString(templateStr)
	if err != nil {
		t.Fatalf("解析模板失败: %v", err)
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

	result, err := tpl.Render(ctx)
	if err != nil {
		t.Fatalf("渲染模板失败: %v", err)
	}

	expected := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-app
  namespace: default
  labels:
    version: v1.0.0
spec:
  replicas: 2
  template:
    spec:
      containers:
        - name: web
          image: nginx:1.25
          env:
             - name: LOG_LEVEL
               value: debug
          ports:
            - containerPort: 80
            - containerPort: 8080
        - name: sidecar
          image: busybox:latest
          env:
             - name: LOG_LEVEL
               value: info
          ports:
            - containerPort: 9000
---
kind: Ingress
metadata:
  name: demo-app-ing
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: demo.example.com
    http: { }
# Snippet included
app: demo-app`

	if result != expected {
		t.Errorf("期望: %q, 实际: %q", expected, result)
	}
}

// TestIntegrationWebPageTemplate 测试完整的网页模板集成
func TestIntegrationWebPageTemplate(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	// 创建测试用的包含文件
	headerTemplate := `<header>
  <h1>${site.title ?? "My Website"}</h1>
  <nav>
    #for link in navigation
    <a href="${link.url}">${link.text}</a>
    #end
  </nav>
</header>`

	footerTemplate := `<footer>
  <p>&copy; ${year ?? 2024} ${site.owner ?? "Unknown"}. All rights reserved.</p>
  #if showSocialLinks
  <div class="social">
    #for social in socialLinks
    <a href="${social.url}" target="_blank">${social.name}</a>
    #end
  </div>
  #end
</footer>`

	err := os.WriteFile("header.tpl", []byte(headerTemplate), 0644)
	if err != nil {
		t.Fatalf("创建header模板失败: %v", err)
	}
	defer os.Remove("header.tpl")

	err = os.WriteFile("footer.tpl", []byte(footerTemplate), 0644)
	if err != nil {
		t.Fatalf("创建footer模板失败: %v", err)
	}
	defer os.Remove("footer.tpl")

	// 主页面模板
	mainTemplate := `<!DOCTYPE html>
<html>
<head>
  <title>${page.title ?? site.title ?? "Untitled"}</title>
  <meta charset="utf-8">
</head>
<body>
  #include "header.tpl"
  
  <main>
    <h2>${page.heading ?? "Welcome"}</h2>
    <p>${page.description ?? "No description available"}</p>
    
    #if showArticles
    <section class="articles">
      <h3>Latest Articles</h3>
      #for article in articles
      <article>
        <h4>${article.title}</h4>
        <p class="meta">By ${article.author ?? "Anonymous"} on ${article.date}</p>
        <p>${article.excerpt ?? "No excerpt available"}</p>
        #if article.tags
        <div class="tags">
          #for tag in article.tags
          <span class="tag">${tag}</span>
          #end
        </div>
        #end
      </article>
      #end
    </section>
    #end
    
    #if showContactForm
    <section class="contact">
      <h3>Contact Us</h3>
      <form>
        <input type="text" placeholder="Name" required>
        <input type="email" placeholder="Email" required>
        <textarea placeholder="Message" required></textarea>
        <button type="submit">Send Message</button>
      </form>
    </section>
    #end
  </main>
  
  #include "footer.tpl"
</body>
</html>`

	tpl, err := eng.ParseString(mainTemplate)
	if err != nil {
		t.Fatalf("解析主模板失败: %v", err)
	}

	ctx := map[string]any{
		"site": map[string]any{
			"title": "Tech Blog",
			"owner": "John Doe",
		},
		"page": map[string]any{
			"title":       "Home - Tech Blog",
			"heading":     "Welcome to My Tech Blog",
			"description": "Sharing insights about technology and programming",
		},
		"navigation": []map[string]any{
			{"text": "Home", "url": "/"},
			{"text": "Articles", "url": "/articles"},
			{"text": "About", "url": "/about"},
			{"text": "Contact", "url": "/contact"},
		},
		"showArticles": true,
		"articles": []map[string]any{
			{
				"title":   "Getting Started with Go",
				"author":  "Jane Smith",
				"date":    "2024-01-15",
				"excerpt": "Learn the basics of Go programming language",
				"tags":    []string{"go", "programming", "tutorial"},
			},
			{
				"title": "Advanced Docker Techniques",
				// 故意省略author来测试默认值
				"date": "2024-01-10",
				// 故意省略excerpt来测试默认值
				"tags": []string{"docker", "devops"},
			},
		},
		"showContactForm":  true,
		"showSocialLinks": true,
		"socialLinks": []map[string]any{
			{"name": "Twitter", "url": "https://twitter.com/johndoe"},
			{"name": "GitHub", "url": "https://github.com/johndoe"},
		},
		"year": 2024,
	}

	result, err := tpl.Render(ctx)
	if err != nil {
		t.Fatalf("渲染模板失败: %v", err)
	}

	// 验证结果包含期望的内容
	expectedContents := []string{
		"<!DOCTYPE html>",
		"<title>Home - Tech Blog</title>",
		"<h1>Tech Blog</h1>",
		"<a href=\"/\">Home</a>",
		"<a href=\"/articles\">Articles</a>",
		"<h2>Welcome to My Tech Blog</h2>",
		"<p>Sharing insights about technology and programming</p>",
		"<h4>Getting Started with Go</h4>",
		"<p class=\"meta\">By Jane Smith on 2024-01-15</p>",
		"<h4>Advanced Docker Techniques</h4>",
		"<p class=\"meta\">By Anonymous on 2024-01-10</p>",
		"<p>No excerpt available</p>",
		"<span class=\"tag\">go</span>",
		"<span class=\"tag\">docker</span>",
		"<input type=\"text\" placeholder=\"Name\" required>",
		"<p>&copy; 2024 John Doe. All rights reserved.</p>",
		"<a href=\"https://twitter.com/johndoe\" target=\"_blank\">Twitter</a>",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(result, expected) {
			t.Errorf("结果中缺少期望内容: %q", expected)
		}
	}
}

// TestIntegrationConfigTemplate 测试配置文件模板集成
func TestIntegrationConfigTemplate(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	// 配置文件模板，包含多种语法特性
	configTemplate := `# Application Configuration
app:
  name: ${app.name}
  version: ${app.version ?? "1.0.0"}
  debug: ${app.debug ?? false}
  
server:
  host: ${server.host ?? "localhost"}
  port: ${server.port ?? 8080}
  #if server.ssl
  ssl:
    enabled: true
    cert: ${server.ssl.cert}
    key: ${server.ssl.key}
  #end
  
database:
  #if database.type == "mysql"
  type: mysql
  host: ${database.host}
  port: ${database.port ?? 3306}
  name: ${database.name}
  user: ${database.user}
  password: ${database.password}
  #else
  type: sqlite
  file: ${database.file ?? "app.db"}
  #end
  
features:
  #for feature in features
  ${feature.name}:
    enabled: ${feature.enabled ?? true}
    #if feature.config
    config:
      #for key, value in feature.config
      ${key}: ${value}
      #end
    #end
  #end
  
logging:
  level: ${logging.level ?? "info"}
  #if logging.outputs
  outputs:
    #for output in logging.outputs
    - type: ${output.type}
      #if output.type == "file"
      path: ${output.path}
      #end
      #if output.format
      format: ${output.format}
      #end
    #end
  #end
  
# Environment specific settings
#if env == "production"
production:
  cache:
    enabled: true
    ttl: ${cache.ttl ?? 3600}
  monitoring:
    enabled: true
    endpoint: ${monitoring.endpoint}
#else
development:
  hot_reload: true
  debug_sql: ${debug.sql ?? false}
#end`

	tpl, err := eng.ParseString(configTemplate)
	if err != nil {
		t.Fatalf("解析配置模板失败: %v", err)
	}

	ctx := map[string]any{
		"app": map[string]any{
			"name":    "MyApp",
			"version": "2.1.0",
			"debug":   true,
		},
		"server": map[string]any{
			"host": "0.0.0.0",
			"port": 9000,
			"ssl": map[string]any{
				"cert": "/path/to/cert.pem",
				"key":  "/path/to/key.pem",
			},
		},
		"database": map[string]any{
			"type":     "mysql",
			"host":     "db.example.com",
			"name":     "myapp_db",
			"user":     "dbuser",
			"password": "secret123",
			// 故意省略port来测试默认值
		},
		"features": []map[string]any{
			{
				"name":    "authentication",
				"enabled": true,
				"config": map[string]any{
					"jwt_secret": "mysecret",
					"expires_in": "24h",
				},
			},
			{
				"name": "caching",
				// 故意省略enabled来测试默认值
			},
			{
				"name":    "analytics",
				"enabled": false,
			},
		},
		"logging": map[string]any{
			"level": "debug",
			"outputs": []map[string]any{
				{
					"type":   "console",
					"format": "json",
				},
				{
					"type": "file",
					"path": "/var/log/myapp.log",
				},
			},
		},
		"env": "production",
		"cache": map[string]any{
			"ttl": 7200,
		},
		"monitoring": map[string]any{
			"endpoint": "https://monitoring.example.com",
		},
	}

	result, err := tpl.Render(ctx)
	if err != nil {
		t.Fatalf("渲染配置模板失败: %v", err)
	}
	
	// 验证结果包含期望的配置内容
	expectedContents := []string{
		"name: MyApp",
		"version: 2.1.0",
		"debug: true",
		"host: 0.0.0.0",
		"port: 9000",
		"ssl:",
		"enabled: true",
		"cert: /path/to/cert.pem",
		"type: mysql",
		"port: 3306", // 默认值
		"authentication:",
		"enabled: true",
		"jwt_secret: mysecret",
		"caching:",
		"enabled: true", // 默认值
		"analytics:",
		"enabled: false",
		"level: debug",
		"type: console",
		"format: json",
		"type: file",
		"path: /var/log/myapp.log",
		"production:",
		"cache:",
		"ttl: 7200",
		"monitoring:",
		"endpoint: https://monitoring.example.com",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(result, expected) {
			t.Errorf("配置结果中缺少期望内容: %q", expected)
		}
	}
}

// TestIntegrationErrorHandling 测试集成场景中的错误处理
func TestIntegrationErrorHandling(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	tests := []struct {
		name        string
		template    string
		context     map[string]any
		shouldError bool
		errorMsg    string
	}{
		{
			name: "包含不存在的文件",
			template: `Before
#include "nonexistent.tpl"
After`,
			context:     map[string]any{},
			shouldError: true,
			errorMsg:    "no such file",
		},
		{
			name: "循环中的表达式错误",
			template: `#for item in items
${item.nonexistent.field}
#end`,
			context: map[string]any{
				"items": []map[string]any{
					{"name": "test"},
				},
			},
			shouldError: true,
		},
		{
			name: "条件中的表达式错误",
			template: `#if invalidExpression.field
Should not reach
#end`,
			context:     map[string]any{},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := eng.ParseString(tt.template)
			if err != nil {
				if tt.shouldError {
					if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
						t.Errorf("期望错误消息包含 %q, 实际: %v", tt.errorMsg, err)
					}
					return
				}
				t.Fatalf("解析模板失败: %v", err)
			}

			_, err = tpl.Render(tt.context)
			if tt.shouldError {
				if err == nil {
					t.Errorf("期望出现错误，但成功执行了")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("期望错误消息包含 %q, 实际: %v", tt.errorMsg, err)
				}
			} else if err != nil {
				t.Fatalf("渲染模板失败: %v", err)
			}
		})
	}
}

// BenchmarkIntegrationComplex 复杂集成场景性能基准测试
func BenchmarkIntegrationComplex(b *testing.B) {
	loader := os.DirFS(".")
	eng := New(loader)

	// 创建复杂的模板，包含所有语法特性
	complexTemplate := `# Complex Template
App: ${app.name ?? "DefaultApp"}
#if app.features
Features:
#for feature in app.features
  - ${feature.name}: ${feature.enabled ?? true}
    #if feature.config
    Config:
    #for key, value in feature.config
      ${key}: ${value ?? "default"}
    #end
    #end
#end
#end

#if app.environments
Environments:
#for env in app.environments
  ${env.name}:
    host: ${env.host ?? "localhost"}
    port: ${env.port ?? 8080}
    #if env.database
    database:
      type: ${env.database.type ?? "sqlite"}
      #if env.database.type == "mysql"
      host: ${env.database.host}
      port: ${env.database.port ?? 3306}
      #end
    #end
#end
#end`

	// 创建复杂的上下文数据
	context := map[string]any{
		"app": map[string]any{
			"name": "BenchmarkApp",
			"features": []map[string]any{
				{
					"name":    "auth",
					"enabled": true,
					"config": map[string]any{
						"jwt_secret": "secret",
						"expires":    "24h",
					},
				},
				{
					"name": "cache",
					// 故意省略enabled和config来测试默认值
				},
				{
					"name":    "logging",
					"enabled": false,
					"config": map[string]any{
						"level": "debug",
					},
				},
			},
			"environments": []map[string]any{
				{
					"name": "development",
					"host": "dev.example.com",
					"port": 3000,
					"database": map[string]any{
						"type": "sqlite",
					},
				},
				{
					"name": "production",
					"host": "prod.example.com",
					// 故意省略port来测试默认值
					"database": map[string]any{
						"type": "mysql",
						"host": "db.prod.example.com",
						// 故意省略port来测试默认值
					},
				},
			},
		},
	}

	tpl, err := eng.ParseString(complexTemplate)
	if err != nil {
		b.Fatalf("解析复杂模板失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tpl.Render(context)
		if err != nil {
			b.Fatalf("渲染复杂模板失败: %v", err)
		}
	}
}