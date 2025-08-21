package main

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// node AST节点接口
type node interface {
	render(sb *strings.Builder, eng *Engine, ctx map[string]any) error
}

// textNode 文本节点
type textNode struct{ text string }

// render 渲染文本节点
func (n *textNode) render(sb *strings.Builder, _ *Engine, _ map[string]any) error {
	sb.WriteString(n.text)
	return nil
}

// exprNode 表达式节点
type exprNode struct{ code string }

// render 渲染表达式节点
func (n *exprNode) render(sb *strings.Builder, _ *Engine, ctx map[string]any) error {
	v, err := evalExpr(n.code, ctx)
	if err != nil {
		return err
	}
	if v != nil {
		sb.WriteString(fmt.Sprint(v))
	}
	return nil
}

// ifNode 条件节点
type ifNode struct {
	cond  string
	thenN []node
	elseN []node
}

// render 渲染条件节点
func (n *ifNode) render(sb *strings.Builder, eng *Engine, ctx map[string]any) error {
	res, err := evalBool(n.cond, ctx)
	if err != nil {
		return err
	}
	var list []node
	if res {
		list = n.thenN
	} else {
		list = n.elseN
	}
	for _, c := range list {
		if err := c.render(sb, eng, ctx); err != nil {
			return err
		}
	}
	return nil
}

// forNode 循环节点
type forNode struct {
	varName string
	iter    string // expression that should evaluate to slice/array/map/string
	body    []node
}

// render 方法渲染 for 循环节点，支持迭代多种数据类型
func (n *forNode) render(sb *strings.Builder, eng *Engine, ctx map[string]any) error {
	val, err := evalExpr(n.iter, ctx)
	if err != nil {
		return fmt.Errorf("#for eval failed: %w", err)
	}
	switch v := val.(type) {
	case []any:
		for _, item := range v {
			ctx[n.varName] = item
			for _, c := range n.body {
				if err := c.render(sb, eng, ctx); err != nil {
					return err
				}
			}
		}
	case []string:
		for _, item := range v {
			ctx[n.varName] = item
			for _, c := range n.body {
				if err := c.render(sb, eng, ctx); err != nil {
					return err
				}
			}
		}
	case []int:
		for _, item := range v {
			ctx[n.varName] = item
			for _, c := range n.body {
				if err := c.render(sb, eng, ctx); err != nil {
					return err
				}
			}
		}
	case map[string]any:
		for k, item := range v {
			ctx[n.varName] = map[string]any{"key": k, "value": item}
			for _, c := range n.body {
				if err := c.render(sb, eng, ctx); err != nil {
					return err
				}
			}
		}
	case string:
		for _, r := range v {
			ctx[n.varName] = string(r)
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

// includeNode 包含节点
type includeNode struct{ path string }

// render 渲染包含节点
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