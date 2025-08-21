package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	expr "github.com/expr-lang/expr"
)

// nullCoalesceFunc 实现空安全运算符 ?? 的逻辑
// nullCoalesceFunc 实现空安全运算符的逻辑
// 这个函数会在运行时被调用，如果第一个参数计算失败或为空值，则返回第二个参数
func nullCoalesceFunc(a, b any) any {
	if a == nil {
		return b
	}
	if s, ok := a.(string); ok && s == "" {
		return b
	}
	return a
}

// preprocessNullCoalescing 预处理空安全运算符 ??
// 将 "a ?? b" 转换为 "nullCoalesce(a, b)"
func preprocessNullCoalescing(code string) string {
	
	// 处理 ?? 运算符 - 递归处理所有的 ?? 运算符
	re := regexp.MustCompile(`([^?]+?)\s*\?\?\s*(.+)`)
	for re.MatchString(code) {
		matches := re.FindStringSubmatch(code)
		if len(matches) == 3 {
			left := strings.TrimSpace(matches[1])
			right := strings.TrimSpace(matches[2])
			
			// 对于嵌套属性访问，使用safeGet包装
			if strings.Contains(left, ".") && !strings.Contains(left, "(") {
				parts := strings.SplitN(left, ".", 2)
				left = fmt.Sprintf(`safeGet(%s, "%s")`, parts[0], parts[1])
			}
			
			// 对于数组索引访问，使用safeIndex包装
			if strings.Contains(left, "[") && strings.Contains(left, "]") {
				re2 := regexp.MustCompile(`(\w+)\[(\d+)\]`)
				left = re2.ReplaceAllString(left, "safeIndex($1, $2)")
			}
			
			// 递归处理右侧部分
			right = preprocessNullCoalescing(right)
			code = fmt.Sprintf("nullCoalesce(%s, %s)", left, right)
			break // 处理完一次就退出
		}
	}
	

	return code
}

// safeGet 安全地获取嵌套属性，如果路径中任何部分为 nil 则返回 nil
func safeGet(obj any, path string) any {
	if obj == nil {
		return nil
	}
	
	parts := strings.Split(path, ".")
	current := obj
	
	for _, part := range parts {
		if current == nil {
			return nil
		}
		
		switch v := current.(type) {
		case map[string]any:
			current = v[part]
		default:
			return nil
		}
	}
	
	return current
}

// safeIndex 安全地访问数组索引，如果索引越界则返回 nil
func safeIndex(arr any, index int) any {
	if arr == nil {
		return nil
	}
	
	switch v := arr.(type) {
	case []any:
		if index < 0 || index >= len(v) {
			return nil
		}
		return v[index]
	case []string:
		if index < 0 || index >= len(v) {
			return nil
		}
		return v[index]
	case []int:
		if index < 0 || index >= len(v) {
			return nil
		}
		return v[index]
	default:
		return nil
	}
}

// evalExpr 计算表达式的值
func evalExpr(code string, ctx map[string]any) (any, error) {
	// 预处理空安全运算符
	code = preprocessNullCoalescing(code)

	// 创建环境并添加自定义函数和Go标准库函数
	env := map[string]any{
		"nullCoalesce": nullCoalesceFunc,
		"safeGet":     safeGet,
		"safeIndex":   safeIndex,
		// strings 包函数
		"strings": map[string]any{
			"ToUpper":    strings.ToUpper,
			"ToLower":    strings.ToLower,
			"TrimSpace":  strings.TrimSpace,
			"Contains":   strings.Contains,
			"HasPrefix":  strings.HasPrefix,
			"HasSuffix":  strings.HasSuffix,
			"Replace":    strings.Replace,
			"ReplaceAll": strings.ReplaceAll,
			"Split":      strings.Split,
			"Join":       strings.Join,
			"Repeat":     strings.Repeat,
		},
		// strconv 包函数
		"strconv": map[string]any{
			"Atoi":     strconv.Atoi,
			"Itoa":     strconv.Itoa,
			"ParseInt": strconv.ParseInt,
			"FormatInt": strconv.FormatInt,
		},
		// time 包函数
		"time": map[string]any{
			"Now": time.Now,
		},
		// 常用的全局函数
		"len": func(v any) int {
			switch val := v.(type) {
			case string:
				return len(val)
			case []any:
				return len(val)
			case []string:
				return len(val)
			case []int:
				return len(val)
			case map[string]any:
				return len(val)
			default:
				return 0
			}
		},
	}
	// 合并用户上下文
	for k, v := range ctx {
		env[k] = v
	}

	program, err := expr.Compile(code, expr.Env(env), expr.AllowUndefinedVariables())
	if err != nil {
		return nil, err
	}
	
	result, err := expr.Run(program, env)
	if err != nil {
		// 对于数组越界访问，返回nil而不是错误
		if strings.Contains(err.Error(), "index out of range") {
			return nil, nil
		}
		return nil, err
	}
	
	// 检查除零错误 - 检查是否为无穷大或NaN
	if f, ok := result.(float64); ok {
		if math.IsInf(f, 0) || math.IsNaN(f) {
			return nil, fmt.Errorf("division by zero")
		}
	}
	
	return result, nil
}

// evalBool 计算布尔表达式的值，支持真值判断
func evalBool(code string, ctx map[string]any) (bool, error) {
	v, err := evalExpr(code, ctx)
	if err != nil {
		return false, err
	}
	
	// 如果是布尔值，直接返回
	if b, ok := v.(bool); ok {
		return b, nil
	}
	
	// 真值判断：nil、0、空字符串、空数组、空map为假，其他为真
	switch val := v.(type) {
	case nil:
		return false, nil
	case string:
		return val != "", nil
	case int:
		return val != 0, nil
	case int64:
		return val != 0, nil
	case float64:
		return val != 0, nil
	case []any:
		return len(val) > 0, nil
	case []string:
		return len(val) > 0, nil
	case []int:
		return len(val) > 0, nil
	case []map[string]any:
		return len(val) > 0, nil
	case map[string]any:
		return len(val) > 0, nil
	default:
		// 其他类型默认为真
		return true, nil
	}
}