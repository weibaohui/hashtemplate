package main

import (
	"os"
	"testing"
)

// TestConditionalStatements 测试条件语句语法
func TestConditionalStatements(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	tests := []struct {
		name     string
		template string
		context  map[string]any
		expected string
	}{
		{
			name: "基本if语句 - 条件为真",
			template: `#if enabled
Feature is enabled
#end`,
			context:  map[string]any{"enabled": true},
			expected: "Feature is enabled\n",
		},
		{
			name: "基本if语句 - 条件为假",
			template: `#if enabled
Feature is enabled
#end`,
			context:  map[string]any{"enabled": false},
			expected: "",
		},
		{
			name: "if-else语句 - 条件为真",
			template: `#if isAdmin
Admin panel
#else
User panel
#end`,
			context:  map[string]any{"isAdmin": true},
			expected: "Admin panel\n",
		},
		{
			name: "if-else语句 - 条件为假",
			template: `#if isAdmin
Admin panel
#else
User panel
#end`,
			context:  map[string]any{"isAdmin": false},
			expected: "User panel\n",
		},
		{
			name: "变量比较条件",
			template: `#if age >= 18
Adult
#else
Minor
#end`,
			context:  map[string]any{"age": 25},
			expected: "Adult\n",
		},
		{
			name: "字符串比较条件",
			template: `#if role == "admin"
Welcome, Administrator!
#else
Welcome, User!
#end`,
			context:  map[string]any{"role": "admin"},
			expected: "Welcome, Administrator!\n",
		},
		{
			name: "逻辑AND条件",
			template: `#if isLoggedIn && hasPermission
Access granted
#else
Access denied
#end`,
			context: map[string]any{
				"isLoggedIn":    true,
				"hasPermission": true,
			},
			expected: "Access granted\n",
		},
		{
			name: "逻辑OR条件",
			template: `#if isVip || isPremium
Special content
#else
Regular content
#end`,
			context: map[string]any{
				"isVip":     false,
				"isPremium": true,
			},
			expected: "Special content\n",
		},
		{
			name: "嵌套对象属性条件",
			template: `#if user.isActive
User ${user.name} is active
#else
User ${user.name} is inactive
#end`,
			context: map[string]any{
				"user": map[string]any{
					"name":     "张三",
					"isActive": true,
				},
			},
			expected: "User 张三 is active\n",
		},
		{
			name: "数组长度条件",
			template: `#if len(items) > 0
Items found: ${len(items)}
#else
No items found
#end`,
			context: map[string]any{
				"items": []string{"apple", "banana"},
			},
			expected: "Items found: 2\n",
		},
		{
			name: "条件中包含插值",
			template: `#if enabled
Status: ${status}
Message: ${message}
#end`,
			context: map[string]any{
				"enabled": true,
				"status":  "OK",
				"message": "Everything is working",
			},
			expected: "Status: OK\nMessage: Everything is working\n",
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

// TestNestedConditionals 测试嵌套条件语句
func TestNestedConditionals(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	tests := []struct {
		name     string
		template string
		context  map[string]any
		expected string
	}{
		{
			name: "嵌套if语句",
			template: `#if isLoggedIn
#if isAdmin
Admin Dashboard
#else
User Dashboard
#end
#else
Please login
#end`,
			context: map[string]any{
				"isLoggedIn": true,
				"isAdmin":    true,
			},
			expected: "Admin Dashboard\n",
		},
		{
			name: "深层嵌套条件",
			template: `#if level1
Level 1
#if level2
Level 2
#if level3
Level 3
#end
#end
#end`,
			context: map[string]any{
				"level1": true,
				"level2": true,
				"level3": true,
			},
			expected: "Level 1\nLevel 2\nLevel 3\n",
		},
		{
			name: "复杂嵌套逻辑",
			template: `#if user.type == "premium"
#if user.credits > 0
Premium user with ${user.credits} credits
#else
Premium user with no credits
#end
#else
#if user.type == "basic"
Basic user
#else
Guest user
#end
#end`,
			context: map[string]any{
				"user": map[string]any{
					"type":    "premium",
					"credits": 100,
				},
			},
			expected: "Premium user with 100 credits\n",
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

// TestConditionalEdgeCases 测试条件语句的边界情况
func TestConditionalEdgeCases(t *testing.T) {
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
			name:     "空条件块",
			template: `#if true
#end`,
			context:  map[string]any{},
			expected: "",
		},
		{
			name:     "空else块",
			template: `#if false
Never shown
#else
#end`,
			context:  map[string]any{},
			expected: "",
		},
		{
			name:     "nil值条件",
			template: `#if nullValue == nil
Null is nil
#else
Not null
#end`,
			context:  map[string]any{"nullValue": nil},
			expected: "Null is nil\n",
		},
		{
			name:     "零值条件",
			template: `#if zeroValue == 0
Zero value
#else
Not zero
#end`,
			context:  map[string]any{"zeroValue": 0},
			expected: "Zero value\n",
		},
		{
			name:     "空字符串条件",
			template: `#if emptyString == ""
Empty string
#else
Not empty
#end`,
			context:  map[string]any{"emptyString": ""},
			expected: "Empty string\n",
		},
		{
			name:     "未定义变量条件",
			template: `#if hasValue
Has value
#else
No value
#end`,
			context:  map[string]any{"hasValue": false},
			expected: "No value\n",
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

// BenchmarkConditional 条件语句性能基准测试
func BenchmarkConditional(b *testing.B) {
	loader := os.DirFS(".")
	eng := New(loader)

	template := `#if isEnabled
Feature is enabled for ${user}
#else
Feature is disabled
#end`
	context := map[string]any{
		"isEnabled": true,
		"user":      "testuser",
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