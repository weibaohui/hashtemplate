package main

import (
	"errors"
	"regexp"
	"strings"
)

// parser 模板解析器
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

// 正则表达式模式
var (
	reIf      = regexp.MustCompile(`^\s*#if\s+(.+)$`)
	reElse    = regexp.MustCompile(`^\s*#else\s*$`)
	reEnd     = regexp.MustCompile(`^\s*#end\s*$`)
	reFor     = regexp.MustCompile(`^\s*#for\s+([a-zA-Z_][a-zA-Z0-9_]*(?:\s*,\s*[a-zA-Z_][a-zA-Z0-9_]*)?)\s+in\s+(.+)$`)
	reInclude = regexp.MustCompile(`^\s*#include\s+"([^"]+)"\s*$`)
)

// parse 解析模板内容
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

		// Directive: #for x in expr 或 #for key, value in expr
		if m := reFor.FindStringSubmatch(line); m != nil {
			p.cursor++
			body, err := p.parseUntilEnd()
			if err != nil {
				return nil, err
			}
			
			// 解析变量名，支持 key, value 语法
			vars := strings.Split(m[1], ",")
			var varName, varName2 string
			if len(vars) == 1 {
				varName = strings.TrimSpace(vars[0])
			} else if len(vars) == 2 {
				varName = strings.TrimSpace(vars[0])
				varName2 = strings.TrimSpace(vars[1])
			}
			
			nodes = append(nodes, &forNode{varName: varName, varName2: varName2, iter: m[2], body: body})
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

// parseIfBlocks 解析 if 块
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

// parseUntilEnd 解析直到遇到 #end
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

// nodeFactory 节点工厂函数类型
type nodeFactory func(code string) node

// splitByRegex 根据正则表达式分割字符串并创建节点
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