package tagexpr

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

type selectorExprNode struct {
	exprBackground
	field, name  string
	subExprs     []ExprNode
	boolOpposite *bool
	signOpposite *bool
}

func (se *selectorExprNode) String() string {
	return fmt.Sprintf("(%s)%s", se.field, se.name)
}

func (p *Expr) readSelectorExprNode(expr *string) ExprNode {
	field, name, subSelector, boolOpposite, signOpposite, found := findSelector(expr)
	if !found {
		return nil
	}
	operand := &selectorExprNode{
		field:        field,
		name:         name,
		boolOpposite: boolOpposite,
		signOpposite: signOpposite,
	}
	operand.subExprs = make([]ExprNode, 0, len(subSelector))
	for _, s := range subSelector {
		grp := newGroupExprNode()
		err := p.parseExprNode(&s, grp)
		if err != nil {
			return nil
		}
		sortPriority(grp)
		operand.subExprs = append(operand.subExprs, grp)
	}
	return operand
}

var selectorRegexp = regexp.MustCompile(`^([\!\+\-]*)(\([ \t]*[A-Za-z_]+[A-Za-z0-9_\.]*[ \t]*\))?(\$)([\)\[\],\+\-\*\/%><\|&!=\^ \t\\]|$)`)

func findSelector(expr *string) (field string, name string, subSelector []string, boolOpposite, signOpposite *bool, found bool) {
	raw := *expr
	a := selectorRegexp.FindAllStringSubmatch(raw, -1)
	if len(a) != 1 {
		return
	}
	r := a[0]
	if s0 := r[2]; len(s0) > 0 {
		field = strings.TrimSpace(s0[1 : len(s0)-1])
	}
	name = r[3]
	*expr = (*expr)[len(a[0][0])-len(r[4]):]
	for {
		sub := readPairedSymbol(expr, '[', ']')
		if sub == nil {
			break
		}
		if *sub == "" || (*sub)[0] == '[' {
			*expr = raw
			return "", "", nil, nil, nil, false
		}
		subSelector = append(subSelector, strings.TrimSpace(*sub))
	}
	prefix := r[1]
	if len(prefix) == 0 {
		found = true
		return
	}
	_, boolOpposite, signOpposite = getBoolAndSignOpposite(&prefix)
	found = true
	return
}

func (se *selectorExprNode) Run(ctx context.Context, currField string, tagExpr *TagExpr) interface{} {
	var subFields []interface{}
	if n := len(se.subExprs); n > 0 {
		subFields = make([]interface{}, n)
		for i, e := range se.subExprs {
			subFields[i] = e.Run(ctx, currField, tagExpr)
		}
	}
	field := se.field
	if field == "" {
		field = currField
	}
	v := tagExpr.getValue(field, subFields)
	return realValue(v, se.boolOpposite, se.signOpposite)
}
