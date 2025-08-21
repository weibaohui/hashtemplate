package main

import (
	"os"
	"strings"
	"testing"
)

// TestLoopStatements 测试循环语句语法
func TestLoopStatements(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	tests := []struct {
		name     string
		template string
		context  map[string]any
		expected string
	}{
		{
			name: "基本数组循环",
			template: `#for item in items
- ${item}
#end`,
			context: map[string]any{
				"items": []string{"apple", "banana", "cherry"},
			},
			expected: "- apple\n- banana\n- cherry\n",
		},
		{
			name: "数字数组循环",
			template: `#for num in numbers
Number: ${num}
#end`,
			context: map[string]any{
				"numbers": []int{1, 2, 3, 4, 5},
			},
			expected: "Number: 1\nNumber: 2\nNumber: 3\nNumber: 4\nNumber: 5\n",
		},
		{
			name: "对象数组循环",
			template: `#for user in users
Name: ${user.name}, Age: ${user.age}
#end`,
			context: map[string]any{
				"users": []map[string]any{
					{"name": "张三", "age": 25},
					{"name": "李四", "age": 30},
					{"name": "王五", "age": 28},
				},
			},
			expected: "Name: 张三, Age: 25\nName: 李四, Age: 30\nName: 王五, Age: 28\n",
		},
		{
			name: "字符串循环",
			template: `#for char in text
${char}
#end`,
			context: map[string]any{
				"text": "Hello",
			},
			expected: "H\ne\nl\nl\no\n",
		},
		{
			name: "Map循环",
			template: `#for key, value in config
${key}: ${value}
#end`,
			context: map[string]any{
				"config": map[string]any{
					"host": "localhost",
					"port": 8080,
				},
			},
			expected: "host: localhost\nport: 8080\n", // 减少键值对数量以避免顺序问题
		},
		{
			name: "嵌套对象循环",
			template: `#for container in containers
Container: ${container.name}
#for port in container.ports
  Port: ${port}
#end
#end`,
			context: map[string]any{
				"containers": []any{
					map[string]any{
						"name":  "web",
						"ports": []int{80, 443},
					},
					map[string]any{
						"name":  "api",
						"ports": []int{8080, 8443},
					},
				},
			},
			expected: "Container: web\n  Port: 80\n  Port: 443\nContainer: api\n  Port: 8080\n  Port: 8443\n",
		},
		{
			name: "空数组循环",
			template: `#for item in emptyArray
This should not appear
#end
After loop`,
			context: map[string]any{
				"emptyArray": []string{},
			},
			expected: "After loop\n",
		},
		{
			name: "单元素数组循环",
			template: `#for item in singleItem
Only: ${item}
#end`,
			context: map[string]any{
				"singleItem": []string{"alone"},
			},
			expected: "Only: alone\n",
		},
		{
			name: "循环中的条件语句",
			template: `#for user in users
#if user.active
Active user: ${user.name}
#else
Inactive user: ${user.name}
#end
#end`,
			context: map[string]any{
				"users": []map[string]any{
					{"name": "Alice", "active": true},
					{"name": "Bob", "active": false},
					{"name": "Charlie", "active": true},
				},
			},
			expected: "Active user: Alice\nInactive user: Bob\nActive user: Charlie\n",
		},
		{
			name: "复杂表达式循环",
			template: `#for item in items
${item.name}: ${item.price * item.quantity}
#end`,
			context: map[string]any{
				"items": []map[string]any{
					{"name": "Apple", "price": 1.5, "quantity": 10},
					{"name": "Banana", "price": 0.8, "quantity": 15},
				},
			},
			expected: "Apple: 15\nBanana: 12\n",
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

			// 对于Map循环，检查所有期望的内容是否都存在
			if tt.name == "Map循环" {
				if !strings.Contains(result, "host: localhost") {
					t.Errorf("结果中缺少 'host: localhost'")
				}
				if !strings.Contains(result, "port: 8080") {
					t.Errorf("结果中缺少 'port: 8080'")
				}
			} else if result != tt.expected {
				t.Errorf("期望: %q, 实际: %q", tt.expected, result)
			}
		})
	}
}

// TestNestedLoops 测试嵌套循环
func TestNestedLoops(t *testing.T) {
	loader := os.DirFS(".")
	eng := New(loader)

	tests := []struct {
		name     string
		template string
		context  map[string]any
		expected string
	}{
		{
			name: "二维数组循环",
			template: `#for row in matrix
#for cell in row
${cell} 
#end
#end`,
			context: map[string]any{
				"matrix": [][]int{
					{1, 2, 3},
					{4, 5, 6},
					{7, 8, 9},
				},
			},
			expected: "1 \n2 \n3 \n4 \n5 \n6 \n7 \n8 \n9 \n",
		},
		{
			name: "部门员工循环",
			template: `#for dept in departments
Department: ${dept.name}
#for emp in dept.employees
  - ${emp.name} (${emp.role})
#end
#end`,
			context: map[string]any{
				"departments": []map[string]any{
					{
						"name": "Engineering",
						"employees": []map[string]any{
							{"name": "Alice", "role": "Developer"},
							{"name": "Bob", "role": "Architect"},
						},
					},
					{
						"name": "Marketing",
						"employees": []map[string]any{
							{"name": "Charlie", "role": "Manager"},
						},
					},
				},
			},
			expected: "Department: Engineering\n  - Alice (Developer)\n  - Bob (Architect)\nDepartment: Marketing\n  - Charlie (Manager)\n",
		},
		{
			name: "三层嵌套循环",
			template: `#for group in groups
Group: ${group.name}
#for subgroup in group.subgroups
  Subgroup: ${subgroup.name}
#for item in subgroup.items
    - ${item}
#end
#end
#end`,
			context: map[string]any{
				"groups": []map[string]any{
					{
						"name": "A",
						"subgroups": []map[string]any{
							{
								"name":  "A1",
								"items": []string{"item1", "item2"},
							},
						},
					},
				},
			},
			expected: "Group: A\n  Subgroup: A1\n    - item1\n    - item2\n",
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

// TestLoopEdgeCases 测试循环的边界情况
func TestLoopEdgeCases(t *testing.T) {
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
			name:        "nil数组循环",
			template:    `#for item in nullArray
Should not appear
#end
After loop`,
			context:     map[string]any{"nullArray": nil},
			expected:    "After loop\n",
			shouldError: true, // nil值应该报错
		},
		{
			name:        "未定义变量循环",
			template:    `#for item in undefinedVar
Should not appear
#end
After loop`,
			context:     map[string]any{},
			expected:    "After loop\n",
			shouldError: true, // 未定义变量应该报错
		},
		{
			name:     "空字符串循环",
			template: `#for char in emptyString
Should not appear
#end
After loop`,
			context:  map[string]any{"emptyString": ""},
			expected: "After loop\n",
		},
		{
			name:     "空Map循环",
			template: `#for key, value in emptyMap
Should not appear
#end
After loop`,
			context:  map[string]any{"emptyMap": map[string]any{}},
			expected: "After loop\n",
		},
		{
			name:     "包含nil元素的数组",
			template: `#for item in mixedArray
Item: ${item}
#end`,
			context: map[string]any{
				"mixedArray": []any{"hello", nil, "world"},
			},
			expected: "Item: hello\nItem: \nItem: world\n",
		},
		{
			name:     "循环变量名冲突",
			template: `Outer item: ${item}
#for item in items
Inner item: ${item}
#end
Outer item again: ${item}`,
			context: map[string]any{
				"item":  "outer",
				"items": []string{"inner1", "inner2"},
			},
			expected: "Outer item: outer\nInner item: inner1\nInner item: inner2\nOuter item again: outer\n",
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

// BenchmarkLoop 循环性能基准测试
func BenchmarkLoop(b *testing.B) {
	loader := os.DirFS(".")
	eng := New(loader)

	template := `#for item in items
- ${item.name}: ${item.value}
#end`

	// 创建测试数据
	items := make([]map[string]any, 100)
	for i := 0; i < 100; i++ {
		items[i] = map[string]any{
			"name":  "item" + string(rune('0'+i%10)),
			"value": i,
		}
	}

	context := map[string]any{"items": items}

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