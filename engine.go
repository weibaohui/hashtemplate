package main

import (
	"io/fs"
	"strings"
)

// Engine 模板引擎结构体
type Engine struct {
	Loader fs.FS // where #include reads files from; use os.DirFS(root)
}

// Template 模板结构体
type Template struct {
	engine *Engine
	nodes  []node
}

// New 创建新的模板引擎实例
func New(loader fs.FS) *Engine {
	return &Engine{Loader: loader}
}

// ParseString 解析字符串模板
func (e *Engine) ParseString(s string) (*Template, error) {
	p := newParser(s)
	nodes, err := p.parse()
	if err != nil {
		return nil, err
	}
	return &Template{engine: e, nodes: nodes}, nil
}

// ParseFile 解析文件模板
func (e *Engine) ParseFile(path string) (*Template, error) {
	b, err := fs.ReadFile(e.Loader, path)
	if err != nil {
		return nil, err
	}
	return e.ParseString(string(b))
}

// Render 渲染模板，返回渲染后的字符串
func (t *Template) Render(ctx map[string]any) (string, error) {
	var sb strings.Builder
	for _, n := range t.nodes {
		if err := n.render(&sb, t.engine, ctx); err != nil {
			return "", err
		}
	}
	return sb.String(), nil
}