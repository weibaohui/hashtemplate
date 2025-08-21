package main

import (
	"os"
	"testing"
)

// TestVariableInterpolation 测试变量插值语法
func TestVariableInterpolation(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	tests := []struct {
		name     string
		template string
		context  map[string]any
		expected string
	}{
		{
			name:     "基本字符串插值 - ${} 格式",
			template: "Hello ${name}!",
			context:  map[string]any{"name": "World"},
			expected: "Hello World!\n",
		},
		{
			name:     "基本字符串插值 - #() 格式",
			template: "Hello #(name)!",
			context:  map[string]any{"name": "Go"},
			expected: "Hello Go!\n",
		},
		{
			name:     "数字插值",
			template: "Count: ${count}",
			context:  map[string]any{"count": 42},
			expected: "Count: 42\n",
		},
		{
			name:     "布尔值插值",
			template: "Enabled: ${enabled}",
			context:  map[string]any{"enabled": true},
			expected: "Enabled: true\n",
		},
		{
			name:     "嵌套对象属性访问",
			template: "User: ${user.name}, Email: ${user.email}",
			context: map[string]any{
				"user": map[string]any{
					"name":  "张三",
					"email": "zhangsan@example.com",
				},
			},
			expected: "User: 张三, Email: zhangsan@example.com\n",
		},
		{
			name:     "数组索引访问",
			template: "First: ${items[0]}, Second: ${items[1]}",
			context: map[string]any{
				"items": []string{"apple", "banana", "cherry"},
			},
			expected: "First: apple, Second: banana\n",
		},
		{
			name:     "混合格式插值",
			template: "${name} has #(count) items",
			context: map[string]any{
				"name":  "Alice",
				"count": 5,
			},
			expected: "Alice has 5 items\n",
		},
		{
			name:     "表达式计算",
			template: "Total: ${price * quantity}",
			context: map[string]any{
				"price":    10.5,
				"quantity": 3,
			},
			expected: "Total: 31.5\n",
		},
		{
			name:     "字符串连接",
			template: "Full name: ${firstName + ' ' + lastName}",
			context: map[string]any{
				"firstName": "John",
				"lastName":  "Doe",
			},
			expected: "Full name: John Doe\n",
		},
		{
			name:     "多行模板插值",
			template: "Name: ${name}\nAge: ${age}\nCity: ${city}",
			context: map[string]any{
				"name": "李四",
				"age":  25,
				"city": "北京",
			},
			expected: "Name: 李四\nAge: 25\nCity: 北京\n",
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

// TestInterpolationEdgeCases 测试插值的边界情况
func TestInterpolationEdgeCases(t *testing.T) {
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
			name:     "空值插值",
			template: "Value: ${emptyValue}",
			context:  map[string]any{"emptyValue": nil},
			expected: "Value: \n",
		},
		{
			name:     "零值插值",
			template: "Count: ${zero}",
			context:  map[string]any{"zero": 0},
			expected: "Count: 0\n",
		},
		{
			name:     "空字符串插值",
			template: "Text: '${empty}'",
			context:  map[string]any{"empty": ""},
			expected: "Text: ''\n",
		},
		{
			name:     "特殊字符插值",
			template: "Special: ${special}",
			context:  map[string]any{"special": "Hello\nWorld\t!"},
			expected: "Special: Hello\nWorld\t!\n",
		},
		{
			name:     "Unicode字符插值",
			template: "Unicode: ${unicode}",
			context:  map[string]any{"unicode": "你好世界 🌍"},
			expected: "Unicode: 你好世界 🌍\n",
		},
		{
			name:     "连续插值",
			template: "${a}${b}${c}",
			context: map[string]any{
				"a": "Hello",
				"b": " ",
				"c": "World",
			},
			expected: "Hello World\n",
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

// BenchmarkInterpolation 性能基准测试
func BenchmarkInterpolation(b *testing.B) {
	loader := os.DirFS(".")
	eng := New(loader)

	template := "Hello ${name}! You have ${count} messages."
	context := map[string]any{
		"name":  "User",
		"count": 42,
	}

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