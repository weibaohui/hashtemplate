package main

import (
	"os"
	"testing"
)

// TestNullSafetyOperator 测试空安全运算符 ??
func TestNullSafetyOperator(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	tests := []struct {
		name     string
		template string
		context  map[string]any
		expected string
	}{
		{
			name:     "基本空安全运算符 - 值存在",
			template: "Name: ${name ?? 'Anonymous'}",
			context:  map[string]any{"name": "张三"},
			expected: "Name: 张三\n",
		},
		{
			name:     "基本空安全运算符 - 值不存在",
			template: "Name: ${name ?? 'Anonymous'}",
			context:  map[string]any{},
			expected: "Name: Anonymous\n",
		},
		{
			name:     "空安全运算符 - nil值",
			template: "Value: ${value ?? 'Default'}",
			context:  map[string]any{"value": nil},
			expected: "Value: Default\n",
		},
		{
			name:     "空安全运算符 - 空字符串",
			template: "Text: ${text ?? 'No text'}",
			context:  map[string]any{"text": ""},
			expected: "Text: No text\n",
		},
		{
			name:     "空安全运算符 - 零值数字",
			template: "Count: ${count ?? 10}",
			context:  map[string]any{"count": 0},
			expected: "Count: 0\n", // 零值不被认为是空值
		},
		{
			name:     "空安全运算符 - false值",
			template: "Enabled: ${enabled ?? true}",
			context:  map[string]any{"enabled": false},
			expected: "Enabled: false\n", // false不被认为是空值
		},
		{
			name:     "嵌套对象空安全运算符",
			template: "Email: ${user.email ?? 'no-email@example.com'}",
			context: map[string]any{
				"user": map[string]any{
					"name": "张三",
					// 故意省略email字段
				},
			},
			expected: "Email: no-email@example.com\n",
		},
		{
			name:     "多层嵌套空安全运算符",
			template: "Address: ${user.profile.address ?? 'No address'}",
			context: map[string]any{
				"user": map[string]any{
					"name": "张三",
					// 故意省略profile字段
				},
			},
			expected: "Address: No address\n",
		},
		{
			name:     "数组索引空安全运算符",
			template: "First item: ${items[0] ?? 'No items'}",
			context:  map[string]any{"items": []string{}},
			expected: "First item: No items\n",
		},
		{
			name:     "链式空安全运算符",
			template: "Value: ${a ?? b ?? c ?? 'Final default'}",
			context: map[string]any{
				"a": nil,
				"b": "",
				"c": nil,
			},
			expected: "Value: Final default\n",
		},
		{
			name:     "链式空安全运算符 - 中间有值",
			template: "Value: ${a ?? b ?? c ?? 'Final default'}",
			context: map[string]any{
				"a": nil,
				"b": "Found!",
				"c": "Should not reach",
			},
			expected: "Value: Found!\n",
		},
		{
			name:     "空安全运算符与表达式",
			template: "Result: ${(value * 2) ?? 0}",
			context:  map[string]any{"value": 5},
			expected: "Result: 10\n",
		},
		{
			name:     "空安全运算符与字符串连接",
			template: "Full name: ${(firstName + ' ' + lastName) ?? 'Unknown'}",
			context: map[string]any{
				"firstName": "John",
				"lastName":  "Doe",
			},
			expected: "Full name: John Doe\n",
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



// TestNullSafetyInConditionals 测试条件语句中的空安全运算符
func TestNullSafetyInConditionals(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	tests := []struct {
		name     string
		template string
		context  map[string]any
		expected string
	}{
		{
			name: "条件语句中的空安全运算符",
			template: `#if user.isActive ?? false
User is active
#else
User is not active
#end`,
			context: map[string]any{
				"user": map[string]any{
					"name": "张三",
					"isActive": true,
				},
			},
			expected: "User is active\n",
		},
		{
			name: "条件语句中的空安全运算符 - 匹配",
			template: `#if user.isActive ?? false
User is active
#else
User is not active
#end`,
			context: map[string]any{
				"user": map[string]any{
					"name": "张三",
					// 故意省略isActive字段
				},
			},
			expected: "User is not active\n",
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

// TestNullSafetyInLoops 测试循环中的空安全运算符
func TestNullSafetyInLoops(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	tests := []struct {
		name     string
		template string
		context  map[string]any
		expected string
	}{
		{
			name: "循环中的空安全运算符",
			template: `#for user in users
Name: ${user.name ?? 'Unknown'}, Role: ${user.role ?? 'Guest'}
#end`,
			context: map[string]any{
				"users": []any{
					map[string]any{"name": "张三", "role": "admin"},
					map[string]any{"name": "李四"}, // 故意省略role
					map[string]any{"role": "user"}, // 故意省略name
				},
			},
			expected: "Name: 张三, Role: admin\nName: 李四, Role: Guest\nName: Unknown, Role: user\n",
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

// TestNullSafetyEdgeCases 测试空安全运算符的边界情况
func TestNullSafetyEdgeCases(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	tests := []struct {
		name     string
		template string
		context  map[string]any
		expected string
	}{
		{
			name:     "空安全运算符与数字零值",
			template: "Count: ${count ?? -1}",
			context:  map[string]any{"count": 0},
			expected: "Count: 0\n", // 零值不被认为是空值
		},
		{
			name:     "空安全运算符与布尔false",
			template: "Enabled: ${enabled ?? true}",
			context:  map[string]any{"enabled": false},
			expected: "Enabled: false\n", // false不被认为是空值
		},
		{
			name:     "空安全运算符与空数组",
			template: "Items: ${items ?? 'No items'}",
			context:  map[string]any{"items": []string{}},
			expected: "Items: []\n", // 空数组不被认为是空值
		},
		{
			name:     "空安全运算符与空Map",
			template: "Config: ${config ?? 'No config'}",
			context:  map[string]any{"config": map[string]any{}},
			expected: "Config: map[]\n", // 空Map不被认为是空值
		},
		{
			name:     "简单空安全运算符链",
			template: "Result: ${a ?? b ?? 'default'}",
			context: map[string]any{
				"a": nil,
				"b": "found",
			},
			expected: "Result: found\n",
		},
		{
			name:     "多层空安全运算符",
			template: "Value: ${x ?? y ?? 'fallback'}",
			context: map[string]any{
				"x": nil,
				"y": "success",
			},
			expected: "Value: success\n",
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

// BenchmarkNullSafety 空安全运算符性能基准测试
func BenchmarkNullSafety(b *testing.B) {
	loader := os.DirFS(".")
	eng := New(loader)

	template := "Name: ${name ?? 'Anonymous'}, Age: ${age ?? 0}, Email: ${email ?? 'no-email'}"
	context := map[string]any{
		"name": "Test User",
		// 故意省略age和email来测试默认值
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