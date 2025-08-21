package main

import (
	"fmt"
	"regexp"
	"strings"

	expr "github.com/expr-lang/expr"
)

// preprocessNullCoalescing 预处理空安全运算符 ??
// 将 "a ?? b" 转换为 "default(a, b)"
func preprocessNullCoalescing(code string) string {
	// 使用更精确的正则表达式匹配 ?? 运算符
	// 支持括号、引号等复杂表达式
	re := regexp.MustCompile(`([^?\s]+(?:\([^)]*\))?(?:\"[^\"]*\")?(?:\s*[^?\s]+)*)\s*\?\?\s*([^?\s]+(?:\([^)]*\))?(?:\"[^\"]*\")?(?:\s*[^?\s]+)*)`)
	for re.MatchString(code) {
		matches := re.FindStringSubmatch(code)
		if len(matches) == 3 {
			left := strings.TrimSpace(matches[1])
			right := strings.TrimSpace(matches[2])
			code = re.ReplaceAllString(code, fmt.Sprintf("default(%s, %s)", left, right))
		}
	}
	return code
}

// defaultFunc 实现默认值函数，如果第一个参数为nil、空字符串或不存在，返回第二个参数
func defaultFunc(a, b any) any {
	if a == nil {
		return b
	}
	if s, ok := a.(string); ok && s == "" {
		return b
	}
	return a
}

// evalExpr 计算表达式的值
func evalExpr(code string, ctx map[string]any) (any, error) {
	// 预处理空安全运算符
	code = preprocessNullCoalescing(code)

	// 创建包含默认函数的环境
	env := map[string]any{
		"default": defaultFunc,
	}
	// 合并用户上下文
	for k, v := range ctx {
		env[k] = v
	}

	program, err := expr.Compile(code, expr.Env(env), expr.AllowUndefinedVariables())
	if err != nil {
		return nil, err
	}
	return expr.Run(program, env)
}

// evalBool 计算布尔表达式的值
func evalBool(code string, ctx map[string]any) (bool, error) {
	v, err := evalExpr(code, ctx)
	if err != nil {
		return false, err
	}
	b, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("expression is not boolean: %v", code)
	}
	return b, nil
}