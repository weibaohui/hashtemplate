package main

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// node 接口定义了所有节点类型必须实现的渲染方法
type node interface {
	render(sb *strings.Builder, eng *Engine, ctx map[string]any) error
}

// textNode 文本节点，直接输出文本内容
type textNode struct{ text string }

// render 渲染文本节点
func (n *textNode) render(sb *strings.Builder, _ *Engine, _ map[string]any) error {
	sb.WriteString(n.text)
	return nil
}

// exprNode 表达式节点，计算表达式并输出结果
type exprNode struct{ code string }

// render 渲染表达式节点
func (n *exprNode) render(sb *strings.Builder, _ *Engine, ctx map[string]any) error {
	val, err := evalExpr(n.code, ctx)
	if err != nil {
		return err
	}
	if val != nil {
		sb.WriteString(fmt.Sprintf("%v", val))
	}
	return nil
}

// ifNode 条件节点，根据条件执行不同的分支
type ifNode struct {
	cond  string
	thenN []node
	elseN []node
}

// render 渲染条件节点
func (n *ifNode) render(sb *strings.Builder, eng *Engine, ctx map[string]any) error {
	condResult, err := evalBool(n.cond, ctx)
	if err != nil {
		return err
	}

	var nodes []node
	if condResult {
		nodes = n.thenN
	} else {
		nodes = n.elseN
	}

	for _, child := range nodes {
		if err := child.render(sb, eng, ctx); err != nil {
			return err
		}
	}
	return nil
}

// forNode 循环节点，支持迭代多种数据类型
type forNode struct {
	varName  string // 第一个变量名（或唯一变量名）
	varName2 string // 第二个变量名（用于 key, value 语法）
	iter     string // expression that should evaluate to slice/array/map/string
	body     []node
}

// render 方法渲染 for 循环节点，支持迭代多种数据类型
func (n *forNode) render(sb *strings.Builder, eng *Engine, ctx map[string]any) error {
	val, err := evalExpr(n.iter, ctx)
	if err != nil {
		return fmt.Errorf("#for eval failed: %w", err)
	}

	switch v := val.(type) {
	case []any:
		for i, item := range v {
			if n.varName2 != "" {
				// key, value 语法
				ctx[n.varName] = i
				ctx[n.varName2] = item
			} else {
				// 单变量语法
				ctx[n.varName] = item
			}
			for _, c := range n.body {
				if err := c.render(sb, eng, ctx); err != nil {
					return err
				}
			}
		}
	case []string:
		for i, item := range v {
			if n.varName2 != "" {
				// key, value 语法
				ctx[n.varName] = i
				ctx[n.varName2] = item
			} else {
				// 单变量语法
				ctx[n.varName] = item
			}
			for _, c := range n.body {
				if err := c.render(sb, eng, ctx); err != nil {
					return err
				}
			}
		}
	case []int:
		for i, item := range v {
			if n.varName2 != "" {
				// key, value 语法
				ctx[n.varName] = i
				ctx[n.varName2] = item
			} else {
				// 单变量语法
				ctx[n.varName] = item
			}
			for _, c := range n.body {
				if err := c.render(sb, eng, ctx); err != nil {
					return err
				}
			}
		}
	case []map[string]any:
		for i, item := range v {
			if n.varName2 != "" {
				// key, value 语法
				ctx[n.varName] = i
				ctx[n.varName2] = item
			} else {
				// 单变量语法
				ctx[n.varName] = item
			}
			for _, c := range n.body {
				if err := c.render(sb, eng, ctx); err != nil {
					return err
				}
			}
		}
	case map[string]interface{}:
		for k, item := range v {
			if n.varName2 != "" {
				// key, value 语法
				ctx[n.varName] = k
				ctx[n.varName2] = item
			} else {
				// 单变量语法，只设置 key
				ctx[n.varName] = k
			}
			for _, c := range n.body {
				if err := c.render(sb, eng, ctx); err != nil {
					return err
				}
			}
		}
	case map[string]string:
		for k, item := range v {
			if n.varName2 != "" {
				// key, value 语法
				ctx[n.varName] = k
				ctx[n.varName2] = item
			} else {
				// 单变量语法，只设置 key
				ctx[n.varName] = k
			}
			for _, c := range n.body {
				if err := c.render(sb, eng, ctx); err != nil {
					return err
				}
			}
		}
	case string:
		for i, r := range v {
			if n.varName2 != "" {
				// key, value 语法
				ctx[n.varName] = i
				ctx[n.varName2] = string(r)
			} else {
				// 单变量语法
				ctx[n.varName] = string(r)
			}
			for _, c := range n.body {
				if err := c.render(sb, eng, ctx); err != nil {
					return err
				}
			}
		}
	default:
		return fmt.Errorf("#for does not support iterating %T", val)
	}
	return nil
}

// includeNode 包含文件节点，用于包含其他模板文件
type includeNode struct{ path string }

// render 渲染包含文件节点
func (n *includeNode) render(sb *strings.Builder, eng *Engine, ctx map[string]any) error {
	// Resolve include path safely
	p := filepath.Clean(n.path)
	b, err := fs.ReadFile(eng.Loader, p)
	if err != nil {
		return err
	}
	t, err := eng.ParseString(string(b))
	if err != nil {
		return err
	}
	out, err := t.Render(ctx)
	if err != nil {
		return err
	}
	sb.WriteString(out)
	return nil
}