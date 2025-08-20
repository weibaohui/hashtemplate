// enjoygo: a tiny Enjoy-like template engine prototype in Go
// -----------------------------------------------------------
// Features in this MVP:
//   - Output expressions:  #( ... )  and  ${ ... }
//   - Directives:          #if / #else / #end,  #for x in expr / #end,  #include "file"
//   - Context:             map[string]any passed to Render
//   - Expression eval:     via github.com/antonmedv/expr (safe, fast)
//
// Limitations (kept simple on purpose):
//   - No #define / macro yet (easy to add later)
//   - #include reads from a user-provided loader (here: os.DirFS)
//   - Minimal error reporting; tune as you extend
//
// Usage:
//   go mod init enjoygo-demo
//   go get github.com/antonmedv/expr
//   go run .
//
// Author: ChatGPT (prototype for discussion)

package main

import (
	"bufio"
	"errors"
	"fmt"
	expr "github.com/expr-lang/expr"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// -----------------------------------------------------------
// Engine & Template
// -----------------------------------------------------------

type Engine struct {
	Loader fs.FS // where #include reads files from; use os.DirFS(root)
}

type Template struct {
	engine *Engine
	nodes  []node
}

func New(loader fs.FS) *Engine {
	return &Engine{Loader: loader}
}

func (e *Engine) ParseString(s string) (*Template, error) {
	p := newParser(s)
	nodes, err := p.parse()
	if err != nil {
		return nil, err
	}
	return &Template{engine: e, nodes: nodes}, nil
}

func (e *Engine) ParseFile(path string) (*Template, error) {
	b, err := fs.ReadFile(e.Loader, path)
	if err != nil {
		return nil, err
	}
	return e.ParseString(string(b))
}

func (t *Template) Render(ctx map[string]any) (string, error) {
	var sb strings.Builder
	for _, n := range t.nodes {
		if err := n.render(&sb, t.engine, ctx); err != nil {
			return "", err
		}
	}
	return sb.String(), nil
}

// -----------------------------------------------------------
// AST nodes
// -----------------------------------------------------------

type node interface {
	render(sb *strings.Builder, eng *Engine, ctx map[string]any) error
}

type textNode struct{ text string }

func (n *textNode) render(sb *strings.Builder, _ *Engine, _ map[string]any) error {
	sb.WriteString(n.text)
	return nil
}

type exprNode struct{ code string }

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

type ifNode struct {
	cond  string
	thenN []node
	elseN []node
}

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

type includeNode struct{ path string }

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

// -----------------------------------------------------------
// Parser
// -----------------------------------------------------------

type parser struct {
	lines  []string
	cursor int
}

// newParser 创建新的模板解析器
func newParser(s string) *parser {
	// Normalize line endings
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return &parser{lines: strings.Split(s, "\n")}
}

var (
	reIf      = regexp.MustCompile(`^\s*#if\s+(.+)$`)
	reElse    = regexp.MustCompile(`^\s*#else\s*$`)
	reEnd     = regexp.MustCompile(`^\s*#end\s*$`)
	reFor     = regexp.MustCompile(`^\s*#for\s+([a-zA-Z_][a-zA-Z0-9_]*)\s+in\s+(.+)$`)
	reInclude = regexp.MustCompile(`^\s*#include\s+"([^"]+)"\s*$`)
)

func (p *parser) parse() ([]node, error) {
	var nodes []node
	for p.cursor < len(p.lines) {
		line := p.lines[p.cursor]

		// Directive: #if
		if m := reIf.FindStringSubmatch(line); m != nil {
			p.cursor++
			thenBlock, elseBlock, err := p.parseIfBlocks(m[1])
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, &ifNode{cond: m[1], thenN: thenBlock, elseN: elseBlock})
			continue
		}

		// Directive: #for x in expr
		if m := reFor.FindStringSubmatch(line); m != nil {
			p.cursor++
			body, err := p.parseUntilEnd()
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, &forNode{varName: m[1], iter: m[2], body: body})
			continue
		}

		// Directive: #include "file"
		if m := reInclude.FindStringSubmatch(line); m != nil {
			p.cursor++
			nodes = append(nodes, &includeNode{path: m[1]})
			continue
		}

		// Plain line (may contain expressions)
		p.cursor++
		parts := splitExprs(line)
		nodes = append(nodes, parts...)
	}
	return nodes, nil
}

func (p *parser) parseIfBlocks(cond string) (thenBlock, elseBlock []node, err error) {
	thenBlock = []node{}
	for p.cursor < len(p.lines) {
		line := p.lines[p.cursor]

		if reEnd.MatchString(line) {
			p.cursor++
			return thenBlock, elseBlock, nil
		}
		if reElse.MatchString(line) {
			p.cursor++
			elseBlock, err = p.parseUntilEnd()
			return thenBlock, elseBlock, err
		}

		// Nested directives are supported via re-parse of line kinds
		if m := reIf.FindStringSubmatch(line); m != nil {
			p.cursor++
			th, el, err := p.parseIfBlocks(m[1])
			if err != nil {
				return nil, nil, err
			}
			thenBlock = append(thenBlock, &ifNode{cond: m[1], thenN: th, elseN: el})
			continue
		}
		if m := reFor.FindStringSubmatch(line); m != nil {
			p.cursor++
			body, err := p.parseUntilEnd()
			if err != nil {
				return nil, nil, err
			}
			thenBlock = append(thenBlock, &forNode{varName: m[1], iter: m[2], body: body})
			continue
		}
		if m := reInclude.FindStringSubmatch(line); m != nil {
			p.cursor++
			thenBlock = append(thenBlock, &includeNode{path: m[1]})
			continue
		}

		p.cursor++
		parts := splitExprs(line)
		thenBlock = append(thenBlock, parts...)
	}
	return nil, nil, errors.New("unterminated #if: missing #end")
}

func (p *parser) parseUntilEnd() ([]node, error) {
	var nodes []node
	for p.cursor < len(p.lines) {
		line := p.lines[p.cursor]
		if reEnd.MatchString(line) {
			p.cursor++
			return nodes, nil
		}
		// nested
		if m := reIf.FindStringSubmatch(line); m != nil {
			p.cursor++
			th, el, err := p.parseIfBlocks(m[1])
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, &ifNode{cond: m[1], thenN: th, elseN: el})
			continue
		}
		if m := reFor.FindStringSubmatch(line); m != nil {
			p.cursor++
			body, err := p.parseUntilEnd()
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, &forNode{varName: m[1], iter: m[2], body: body})
			continue
		}
		if m := reInclude.FindStringSubmatch(line); m != nil {
			p.cursor++
			nodes = append(nodes, &includeNode{path: m[1]})
			continue
		}

		p.cursor++
		parts := splitExprs(line)
		nodes = append(nodes, parts...)
	}
	return nil, errors.New("unterminated block: missing #end")
}

// splitExprs splits a line into text/expr nodes.
// Supports:  #( ... )   and   ${ ... }
var (
	reHashExpr  = regexp.MustCompile(`#\((.*?)\)`)  // non-greedy
	reDollarExp = regexp.MustCompile(`\$\{(.*?)\}`) // non-greedy
)

// splitExprs 将包含表达式的行分割成文本节点和表达式节点
func splitExprs(line string) []node {
	// First process #( ... )
	nodes := splitByRegex(line, reHashExpr, func(code string) node { return &exprNode{code: code} })
	// For each text node, further split by ${ ... }
	var out []node
	for _, n := range nodes {
		if t, ok := n.(*textNode); ok {
			out = append(out, splitByRegex(t.text, reDollarExp, func(code string) node { return &exprNode{code: code} })...)
		} else {
			out = append(out, n)
		}
	}
	// 为每行添加换行符
	out = append(out, &textNode{text: "\n"})
	return out
}

type nodeFactory func(code string) node

func splitByRegex(s string, re *regexp.Regexp, makeNode nodeFactory) []node {
	locs := re.FindAllStringSubmatchIndex(s, -1)
	if len(locs) == 0 {
		return []node{&textNode{text: s}}
	}
	var nodes []node
	prevEnd := 0
	for _, loc := range locs {
		start, end := loc[0], loc[1]
		codeStart, codeEnd := loc[2], loc[3]
		if start > prevEnd {
			nodes = append(nodes, &textNode{text: s[prevEnd:start]})
		}
		code := strings.TrimSpace(s[codeStart:codeEnd])
		nodes = append(nodes, makeNode(code))
		prevEnd = end
	}
	if prevEnd < len(s) {
		nodes = append(nodes, &textNode{text: s[prevEnd:]})
	}
	return nodes
}

// -----------------------------------------------------------
// Expression evaluation helpers
// -----------------------------------------------------------

func evalExpr(code string, ctx map[string]any) (any, error) {
	program, err := expr.Compile(code, expr.Env(ctx), expr.AllowUndefinedVariables())
	if err != nil {
		return nil, err
	}
	return expr.Run(program, ctx)
}

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

// -----------------------------------------------------------
// Demo (YAML flavored)
// -----------------------------------------------------------

func main() {
	loader := os.DirFS(".")
	eng := New(loader)

	tplStr := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: $(appName)
spec:
  replicas: #(replicas)
  template:
    spec:
      containers:
        #for c in containers
        - name: $(c.name)
          image: $(c.image):$(c.tag)
          ports:
            #for p in c.ports
            - containerPort: $(p)
            #end
        #end
#if enableIngress
---
kind: Ingress
metadata:
  name: $(appName)-ing
spec:
  rules:
  - host: $(ingress.host)
    http: { }
#end
#include "snippet.tpl"`

	// Note: we support both $(...) and ${...} & #( ... ) formats.
	// For convenience, alias $(...) to ${...}
	tplStr = strings.ReplaceAll(tplStr, "$ (", "$(") // no-op guard
	// 使用正则表达式将 $(x) 转换为 ${x}
	re := regexp.MustCompile(`\$\(([^)]+)\)`)
	tplStr = re.ReplaceAllString(tplStr, "${$1}")

	tpl, err := eng.ParseString(tplStr)
	must(err)

	ctx := map[string]any{
		"appName":       "demo-app",
		"replicas":      2,
		"enableIngress": true,
		"ingress": map[string]any{
			"host": "demo.example.com",
		},
		"containers": []any{
			map[string]any{"name": "web", "image": "nginx", "tag": "1.25", "ports": []int{80, 8080}},
			map[string]any{"name": "sidecar", "image": "busybox", "tag": "stable", "ports": []int{9000}},
		},
	}

	// Prepare an included snippet file at runtime for the demo
	_ = writeFileIfMissing("snippet.tpl", "# Simple include demo\n#(appName) included!\n")

	out, err := tpl.Render(ctx)
	must(err)

	// Print result
	w := bufio.NewWriter(os.Stdout)
	_, _ = w.WriteString(out)
	_ = w.Flush()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func writeFileIfMissing(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	return os.WriteFile(path, []byte(content), 0644)
}
