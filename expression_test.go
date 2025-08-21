package main

import (
	"os"
	"testing"
)

// TestExpressionCalculation 测试括号内表达式计算
func TestExpressionCalculation(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	tests := []struct {
		name     string
		template string
		context  map[string]any
		expected string
	}{
		{
			name:     "基本数学运算 - 加法",
			template: "Result: ${10 + 5}",
			context:  map[string]any{},
			expected: "Result: 15\n",
		},
		{
			name:     "基本数学运算 - 减法",
			template: "Result: ${20 - 8}",
			context:  map[string]any{},
			expected: "Result: 12\n",
		},
		{
			name:     "基本数学运算 - 乘法",
			template: "Result: ${6 * 7}",
			context:  map[string]any{},
			expected: "Result: 42\n",
		},
		{
			name:     "基本数学运算 - 除法",
			template: "Result: ${100 / 4}",
			context:  map[string]any{},
			expected: "Result: 25\n",
		},
		{
			name:     "变量参与计算 - 乘法",
			template: "total: ${price * quantity}",
			context: map[string]any{
				"price":    10.5,
				"quantity": 3,
			},
			expected: "total: 31.5\n",
		},
		{
			name:     "多步计算 - 总价和折扣",
			template: "total: ${price * quantity}\ndiscount: ${total * 0.1}",
			context: map[string]any{
				"price":    100,
				"quantity": 2,
				"total":    200, // 预先计算的总价
			},
			expected: "total: 200\ndiscount: 20\n",
		},
		{
			name:     "复杂表达式 - 含税价格计算",
			template: "subtotal: ${price * quantity}\ntax: ${price * quantity * 0.08}\ntotal: ${price * quantity * 1.08}",
			context: map[string]any{
				"price":    50,
				"quantity": 2,
			},
			expected: "subtotal: 100\ntax: 8\ntotal: 108\n",
		},
		{
			name:     "字符串拼接表达式",
			template: "fullName: ${firstName + ' ' + lastName}",
			context: map[string]any{
				"firstName": "张",
				"lastName":  "三",
			},
			expected: "fullName: 张 三\n",
		},
		{
			name:     "条件表达式 - 三元运算符",
			template: "status: ${age >= 18 ? 'adult' : 'minor'}",
			context: map[string]any{
				"age": 25,
			},
			expected: "status: adult\n",
		},
		{
			name:     "数组长度计算",
			template: "itemCount: ${len(items)}",
			context: map[string]any{
				"items": []string{"apple", "banana", "cherry"},
			},
			expected: "itemCount: 3\n",
		},
		{
			name:     "嵌套对象属性计算",
			template: "area: ${dimensions.width * dimensions.height}",
			context: map[string]any{
				"dimensions": map[string]any{
					"width":  10,
					"height": 5,
				},
			},
			expected: "area: 50\n",
		},
		{
			name:     "数组元素计算",
			template: "sum: ${numbers[0] + numbers[1] + numbers[2]}",
			context: map[string]any{
				"numbers": []any{10, 20, 30},
			},
			expected: "sum: 60\n",
		},
		{
			name:     "百分比计算",
			template: "percentage: ${(current / total) * 100}%",
			context: map[string]any{
				"current": 75,
				"total":   100,
			},
			expected: "percentage: 75%\n",
		},
		{
			name:     "浮点数精度计算",
			template: "result: ${price * 1.15}",
			context: map[string]any{
				"price": 99.99,
			},
			expected: "result: 114.98849999999999\n",
		},
		{
			name:     "布尔逻辑运算",
			template: "canAccess: ${isLoggedIn && hasPermission}",
			context: map[string]any{
				"isLoggedIn":    true,
				"hasPermission": true,
			},
			expected: "canAccess: true\n",
		},
		{
			name:     "Go标准库 - strings.ToUpper",
			template: "upperName: ${strings.ToUpper(appName)}",
			context: map[string]any{
				"appName": "hello-world",
			},
			expected: "upperName: HELLO-WORLD\n",
		},
		{
			name:     "Go标准库 - strings.ToLower",
			template: "lowerName: ${strings.ToLower(appName)}",
			context: map[string]any{
				"appName": "HELLO-WORLD",
			},
			expected: "lowerName: hello-world\n",
		},
		{
			name:     "Go标准库 - strings.Contains",
			template: "hasKeyword: ${strings.Contains(text, keyword)}",
			context: map[string]any{
				"text":    "Hello World",
				"keyword": "World",
			},
			expected: "hasKeyword: true\n",
		},
		{
			name:     "Go标准库 - strings.Replace",
			template: "replaced: ${strings.Replace(text, old, new, 1)}",
			context: map[string]any{
				"text": "hello world hello",
				"old":  "hello",
				"new":  "hi",
			},
			expected: "replaced: hi world hello\n",
		},
		{
			name:     "Go标准库 - strings.Split和len组合",
			template: "wordCount: ${len(strings.Split(sentence, ' '))}",
			context: map[string]any{
				"sentence": "hello world from go",
			},
			expected: "wordCount: 4\n",
		},
		{
			name:     "Go标准库 - strconv.Itoa",
			template: "numberStr: ${strconv.Itoa(number)}",
			context: map[string]any{
				"number": 42,
			},
			expected: "numberStr: 42\n",
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

// TestGoStandardLibrary 测试Go标准库函数调用
func TestGoStandardLibrary(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	tests := []struct {
		name     string
		template string
		context  map[string]any
		expected string
	}{
		{
			name:     "strings.ToUpper - 转换为大写",
			template: "upperName: ${strings.ToUpper(appName)}",
			context: map[string]any{
				"appName": "my-awesome-app",
			},
			expected: "upperName: MY-AWESOME-APP\n",
		},
		{
			name:     "strings.ToLower - 转换为小写",
			template: "lowerName: ${strings.ToLower(appName)}",
			context: map[string]any{
				"appName": "MY-AWESOME-APP",
			},
			expected: "lowerName: my-awesome-app\n",
		},
		{
			name:     "strings.TrimSpace - 去除空格",
			template: "trimmed: '${strings.TrimSpace(text)}'",
			context: map[string]any{
				"text": "  hello world  ",
			},
			expected: "trimmed: 'hello world'\n",
		},
		{
			name:     "strings.Contains - 检查包含",
			template: "contains: ${strings.Contains(text, keyword)}",
			context: map[string]any{
				"text":    "Hello Go Programming",
				"keyword": "Go",
			},
			expected: "contains: true\n",
		},
		{
			name:     "strings.HasPrefix - 检查前缀",
			template: "hasPrefix: ${strings.HasPrefix(filename, prefix)}",
			context: map[string]any{
				"filename": "config.yaml",
				"prefix":   "config",
			},
			expected: "hasPrefix: true\n",
		},
		{
			name:     "strings.HasSuffix - 检查后缀",
			template: "hasGoExt: ${strings.HasSuffix(filename, '.go')}",
			context: map[string]any{
				"filename": "main.go",
			},
			expected: "hasGoExt: true\n",
		},
		{
			name:     "strings.ReplaceAll - 替换所有",
			template: "replaced: ${strings.ReplaceAll(text, old, new)}",
			context: map[string]any{
				"text": "hello world hello universe",
				"old":  "hello",
				"new":  "hi",
			},
			expected: "replaced: hi world hi universe\n",
		},
		{
			name:     "strings.Split - 分割字符串",
			template: "parts: ${strings.Split(csv, ',')}",
			context: map[string]any{
				"csv": "apple,banana,cherry",
			},
			expected: "parts: [apple banana cherry]\n",
		},
		{
			name:     "strings.Join - 连接字符串",
			template: "joined: ${strings.Join(parts, ' | ')}",
			context: map[string]any{
				"parts": []string{"apple", "banana", "cherry"},
			},
			expected: "joined: apple | banana | cherry\n",
		},
		{
			name:     "strings.Repeat - 重复字符串",
			template: "repeated: ${strings.Repeat(char, count)}",
			context: map[string]any{
				"char":  "*",
				"count": 5,
			},
			expected: "repeated: *****\n",
		},
		{
			name:     "strconv.Itoa - 整数转字符串",
			template: "numberStr: ${strconv.Itoa(number)}",
			context: map[string]any{
				"number": 12345,
			},
			expected: "numberStr: 12345\n",
		},
		{
			name:     "组合使用 - 格式化名称",
			template: "formattedName: ${strings.ToUpper(strings.ReplaceAll(name, ' ', '_'))}",
			context: map[string]any{
				"name": "hello world app",
			},
			expected: "formattedName: HELLO_WORLD_APP\n",
		},
		{
			name:     "组合使用 - 文件名处理",
			template: "isGoFile: ${strings.HasSuffix(strings.ToLower(filename), '.go')}",
			context: map[string]any{
				"filename": "MAIN.GO",
			},
			expected: "isGoFile: true\n",
		},
		{
			name:     "组合使用 - 统计单词数",
			template: "wordCount: ${len(strings.Split(strings.TrimSpace(text), ' '))}",
			context: map[string]any{
				"text": "  hello world from golang  ",
			},
			expected: "wordCount: 4\n",
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

// TestComplexExpressions 测试复杂表达式计算
func TestComplexExpressions(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	tests := []struct {
		name     string
		template string
		context  map[string]any
		expected string
	}{
		{
			name:     "购物车总价计算",
			template: "itemTotal: ${item.price * item.quantity}\nshipping: ${itemTotal > 100 ? 0 : 10}\nfinalTotal: ${itemTotal + shipping}",
			context: map[string]any{
				"item": map[string]any{
					"price":    25.99,
					"quantity": 3,
				},
				"itemTotal": 77.97, // 预计算值
				"shipping":  10,     // 预计算值
			},
			expected: "itemTotal: 77.97\nshipping: 10\nfinalTotal: 87.97\n",
		},
		{
			name:     "员工薪资计算",
			template: "baseSalary: ${employee.baseSalary}\nbonus: ${employee.baseSalary * employee.performanceRating * 0.1}\ntotalSalary: ${employee.baseSalary + bonus}",
			context: map[string]any{
				"employee": map[string]any{
					"baseSalary":        5000,
					"performanceRating": 1.2,
				},
				"bonus": 600, // 预计算值
			},
			expected: "baseSalary: 5000\nbonus: 600\ntotalSalary: 5600\n",
		},
		{
			name:     "几何计算 - 圆形面积",
			template: "radius: ${circle.radius}\narea: ${3.14159 * circle.radius * circle.radius}",
			context: map[string]any{
				"circle": map[string]any{
					"radius": 5,
				},
			},
			expected: "radius: 5\narea: 78.53975\n",
		},
		{
			name:     "时间计算 - 小时转分钟",
			template: "hours: ${timeData.hours}\nminutes: ${timeData.hours * 60 + timeData.minutes}",
			context: map[string]any{
				"timeData": map[string]any{
					"hours":   2,
					"minutes": 30,
				},
			},
			expected: "hours: 2\nminutes: 150\n",
		},
		{
			name:     "数组统计计算",
			template: "count: ${len(scores)}\naverage: ${(scores[0] + scores[1] + scores[2]) / len(scores)}",
			context: map[string]any{
				"scores": []int{85, 92, 78},
			},
			expected: "count: 3\naverage: 85\n",
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

// TestExpressionErrorHandling 测试表达式错误处理
func TestExpressionErrorHandling(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	tests := []struct {
		name        string
		template    string
		context     map[string]any
		shouldError bool
	}{
		{
			name:        "除零错误",
			template:    "result: ${10 / 0}",
			context:     map[string]any{},
			shouldError: true,
		},
		{
			name:        "未定义变量",
			template:    "result: ${undefinedVar * 2}",
			context:     map[string]any{},
			shouldError: true,
		},
		{
			name:        "类型不匹配 - 字符串与数字相乘",
			template:    "result: ${'hello' * 5}",
			context:     map[string]any{},
			shouldError: true,
		},
		{
			name:        "数组越界访问",
			template:    "result: ${arr[10]}",
			context:     map[string]any{"arr": []int{1, 2, 3}},
			shouldError: false, // 应该返回 nil 而不是错误
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := eng.ParseString(tt.template)
			if err != nil {
				if tt.shouldError {
					return // 期望的解析错误
				}
				t.Fatalf("解析模板失败: %v", err)
			}

			_, err = tpl.Render(tt.context)
			if tt.shouldError && err == nil {
				t.Errorf("期望出现错误，但成功执行了")
			} else if !tt.shouldError && err != nil {
				t.Errorf("不期望出现错误，但出现了错误: %v", err)
			}
		})
	}
}

// BenchmarkExpressionCalculation 表达式计算性能基准测试
func BenchmarkExpressionCalculation(b *testing.B) {
	loader := os.DirFS(".")
	eng := New(loader)

	template := "total: ${price * quantity}\ndiscount: ${total * 0.1}\nfinal: ${total - discount}"
	context := map[string]any{
		"price":    99.99,
		"quantity": 3,
		"total":    299.97,
		"discount": 29.997,
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