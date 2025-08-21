package main

import (
	"os"
	"testing"
)

// TestIncludeStatements 测试包含文件语法
func TestIncludeStatements(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	// 创建测试用的包含文件
	setupIncludeTestFiles(t)
	defer cleanupIncludeTestFiles(t)

	tests := []struct {
		name     string
		template string
		context  map[string]any
		expected string
	}{
		{
			name:     "基本文件包含",
			template: `Before include
#include "test_header.tpl"
After include`,
			context:  map[string]any{"title": "Test Page"},
			expected: "Before include\n<h1>Test Page</h1>\nAfter include\n",
		},
		{
			name:     "多个文件包含",
			template: `#include "test_header.tpl"
<main>Content here</main>
#include "test_footer.tpl"`,
			context: map[string]any{
				"title": "Multi Include",
				"year":  2024,
			},
			expected: "<h1>Multi Include</h1>\n<main>Content here</main>\n<footer>Copyright 2024</footer>\n",
		},
		{
			name:     "包含文件中的变量插值",
			template: `#include "test_user_card.tpl"`,
			context: map[string]any{
				"user": map[string]any{
					"name":  "张三",
					"email": "zhangsan@example.com",
					"age":   30,
				},
			},
			expected: "<div class=\"user-card\">\n  <h3>张三</h3>\n  <p>Email: zhangsan@example.com</p>\n  <p>Age: 30</p>\n</div>\n",
		},
		{
			name:     "包含文件中的条件语句",
			template: `#include "test_conditional.tpl"`,
			context: map[string]any{
				"isLoggedIn": true,
				"username":   "admin",
			},
			expected: "Welcome back, admin!\n",
		},
		{
			name:     "包含文件中的循环语句",
			template: `#include "test_list.tpl"`,
			context: map[string]any{
				"items": []string{"Apple", "Banana", "Cherry"},
			},
			expected: "<ul>\n<li>Apple</li>\n<li>Banana</li>\n<li>Cherry</li>\n</ul>\n",
		},
		{
			name:     "嵌套包含文件",
			template: `#include "test_nested_parent.tpl"`,
			context: map[string]any{
				"pageTitle": "Nested Test",
				"content":   "This is nested content",
			},
			expected: "<html>\n<head><title>Nested Test</title></head>\n<body>This is nested content</body>\n</html>\n",
		},
		{
			name:     "包含文件路径中的变量",
			template: `Dynamic include test:
#include "test_dynamic.tpl"`,
			context: map[string]any{
				"message": "Hello from dynamic template!",
			},
			expected: "Dynamic include test:\nHello from dynamic template!\n",
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
				t.Errorf("期望: %q, 实际: %q", tt.expected, result)
			}
		})
	}
}

// TestIncludeEdgeCases 测试包含文件的边界情况
func TestIncludeEdgeCases(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	// 创建测试用的包含文件
	setupIncludeTestFiles(t)
	defer cleanupIncludeTestFiles(t)

	tests := []struct {
		name        string
		template    string
		context     map[string]any
		expected    string
		shouldError bool
	}{
		{
			name:        "包含不存在的文件",
			template:    `#include "nonexistent.tpl"`,
			context:     map[string]any{},
			shouldError: true,
		},
		{
			name:     "包含空文件",
			template: `Before
#include "test_empty.tpl"
After`,
			context:  map[string]any{},
			expected: "Before\n\nAfter\n",
		},
		{
			name:     "包含只有空白字符的文件",
			template: `Before
#include "test_whitespace.tpl"
After`,
			context:  map[string]any{},
			expected: "Before\n   \n\t\nAfter\n",
		},
		{
			name:     "包含文件中有语法错误",
			template: `#include "test_syntax_error.tpl"`,
			context:  map[string]any{},
			expected: "This has unclosed ${variable\n",
		},
		{
			name:     "多次包含同一文件",
			template: `#include "test_header.tpl"
Middle content
#include "test_header.tpl"`,
			context:  map[string]any{"title": "Repeated"},
			expected: "<h1>Repeated</h1>\nMiddle content\n<h1>Repeated</h1>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := eng.ParseString(tt.template)
			if err != nil {
				if tt.shouldError {
					return // 期望的错误
				}
				t.Fatalf("解析模板失败: %v", err)
			}

			result, err := tpl.Render(tt.context)
			if err != nil {
				if tt.shouldError {
					return // 期望的错误
				}
				t.Fatalf("渲染模板失败: %v", err)
			}

			if tt.shouldError {
				t.Errorf("期望出现错误，但成功执行了")
				return
			}

			if result != tt.expected {
				t.Errorf("期望: %q, 实际: %q", tt.expected, result)
			}
		})
	}
}

// setupIncludeTestFiles 创建测试用的包含文件
func setupIncludeTestFiles(t *testing.T) {
	testFiles := map[string]string{
		"test_header.tpl": "<h1>${title}</h1>",
		"test_footer.tpl": "<footer>Copyright ${year}</footer>",
		"test_user_card.tpl": `<div class="user-card">
  <h3>${user.name}</h3>
  <p>Email: ${user.email}</p>
  <p>Age: ${user.age}</p>
</div>`,
		"test_conditional.tpl": `#if isLoggedIn
Welcome back, ${username}!
#else
Please log in
#end`,
		"test_list.tpl": `<ul>
#for item in items
<li>${item}</li>
#end
</ul>`,
		"test_nested_parent.tpl": `<html>
#include "test_nested_child.tpl"
</html>`,
		"test_nested_child.tpl":  `<head><title>${pageTitle}</title></head>
<body>${content}</body>`,
		"test_dynamic.tpl":       "${message}",
		"test_empty.tpl":         "",
		"test_whitespace.tpl":    "   \n\t",
		"test_syntax_error.tpl":  "This has unclosed ${variable",
	}

	for filename, content := range testFiles {
		err := os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			t.Fatalf("创建测试文件 %s 失败: %v", filename, err)
		}
	}
}

// cleanupIncludeTestFiles 清理测试文件
func cleanupIncludeTestFiles(t *testing.T) {
	testFiles := []string{
		"test_header.tpl",
		"test_footer.tpl",
		"test_user_card.tpl",
		"test_conditional.tpl",
		"test_list.tpl",
		"test_nested_parent.tpl",
		"test_nested_child.tpl",
		"test_dynamic.tpl",
		"test_empty.tpl",
		"test_whitespace.tpl",
		"test_syntax_error.tpl",
	}

	for _, filename := range testFiles {
		_ = os.Remove(filename) // 忽略错误，因为文件可能不存在
	}
}

// TestIncludeWithComplexContext 测试包含文件与复杂上下文
func TestIncludeWithComplexContext(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	// 创建复杂的包含文件
	complexTemplate := `<div class="product">
  <h2>${product.name}</h2>
  <p>Price: $${product.price}</p>
  #if product.inStock
  <span class="in-stock">In Stock</span>
  #else
  <span class="out-of-stock">Out of Stock</span>
  #end
  <ul class="features">
  #for feature in product.features
    <li>${feature}</li>
  #end
  </ul>
</div>`

	err := os.WriteFile("test_complex.tpl", []byte(complexTemplate), 0644)
	if err != nil {
		t.Fatalf("创建复杂测试文件失败: %v", err)
	}
	defer os.Remove("test_complex.tpl")

	template := `Product Catalog:
#for product in products
#include "test_complex.tpl"
#end`

	context := map[string]any{
		"products": []any{
			map[string]any{
				"name":     "Laptop",
				"price":    999.99,
				"inStock":  true,
				"features": []string{"16GB RAM", "512GB SSD", "Intel i7"},
			},
			map[string]any{
				"name":     "Mouse",
				"price":    29.99,
				"inStock":  false,
				"features": []string{"Wireless", "Ergonomic"},
			},
		},
	}

	tpl, err := eng.ParseString(template)
	if err != nil {
		t.Fatalf("解析模板失败: %v", err)
	}

	result, err := tpl.Render(context)
	if err != nil {
		t.Fatalf("渲染模板失败: %v", err)
	}

	expected := `Product Catalog:
<div class="product">
  <h2>Laptop</h2>
  <p>Price: $999.99</p>
  <span class="in-stock">In Stock</span>
  <ul class="features">
    <li>16GB RAM</li>
    <li>512GB SSD</li>
    <li>Intel i7</li>
  </ul>
</div>
<div class="product">
  <h2>Mouse</h2>
  <p>Price: $29.99</p>
  <span class="out-of-stock">Out of Stock</span>
  <ul class="features">
    <li>Wireless</li>
    <li>Ergonomic</li>
  </ul>
</div>
`

	if result != expected {
		t.Errorf("期望: %q, 实际: %q", expected, result)
	}
}

// BenchmarkInclude 包含文件性能基准测试
func BenchmarkInclude(b *testing.B) {
	loader := os.DirFS(".")
	eng := New(loader)

	// 创建基准测试文件
	err := os.WriteFile("bench_include.tpl", []byte("Hello ${name}!"), 0644)
	if err != nil {
		b.Fatalf("创建基准测试文件失败: %v", err)
	}
	defer os.Remove("bench_include.tpl")

	template := `#include "bench_include.tpl"`
	context := map[string]any{"name": "World"}

	tpl, err := eng.ParseString(template)
	if err != nil {
		b.Fatalf("解析模板失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tpl.Render(context)
		if err != nil {
			b.Fatalf("渲染模板失败: %v", err)
		}
	}
}