package mapnik

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	cartocss "github.com/flywave/go-cartocss"

	"github.com/flywave/go-cartocss/color"
)

var isOgrConnection = regexp.MustCompile(`^[a-zA-Z]{2,}:`)

func fmtField(vals []interface{}, ok bool) *string {
	if !ok {
		return nil
	}
	parts := []string{}
	for _, v := range vals {
		switch v := v.(type) {
		case cartocss.Field:
			parts = append(parts, string(v))
		case string:
			parts = append(parts, "'"+v+"'")
		}
	}
	r := strings.Join(parts, " + ")
	return &r
}

func fmtPattern(v []float64, scale float64, ok bool) *string {
	if !ok {
		return nil
	}
	return fmtPatternStr(v, scale)
}

func fmtPatternStr(v []float64, scale float64) *string {
	parts := make([]string, 0, len(v))
	for i := range v {
		parts = append(parts, strconv.FormatFloat(v[i]*scale, 'f', -1, 64))
	}
	r := strings.Join(parts, ", ")
	return &r
}

func fmtFloatProp(p *cartocss.Properties, name string, scale float64) *string {
	v, ok := p.GetFloat(name)
	if !ok {
		return nil
	}
	r := strconv.FormatFloat(v*scale, 'f', -1, 64)
	return &r
}

func fmtFloat(v float64, ok bool) *string {
	if !ok {
		return nil
	}
	r := strconv.FormatFloat(v, 'f', -1, 64)
	return &r
}

func fmtString(v string, ok bool) *string {
	if !ok {
		return nil
	}
	return &v
}

func fmtBool(v bool, ok bool) *string {
	if !ok {
		return nil
	}
	var r string
	if v {
		r = "true"
	} else {
		r = "false"
	}
	return &r
}

func fmtColor(v color.Color, ok bool) *string {
	if !ok {
		return nil
	}
	r := v.String()
	return &r
}

func fmtFilters(filters []cartocss.Filter) string {
	parts := []string{}
	for _, f := range filters {
		var value string
		switch v := f.Value.(type) {
		case nil:
			value = "null"
		case string:
			value = `'` + v + `'`
		case float64:
			value = string(*fmtFloat(v, true))
		case cartocss.ModuloComparsion:
			value = fmt.Sprintf("%d %s %d", v.Div, v.CompOp, v.Value)
		default:
			log.Printf("unknown type of filter value: %s", v)
			value = ""
		}
		field := f.Field
		if len(field) > 2 && field[0] == '"' && field[len(field)-1] == '"' {
			field = field[1 : len(field)-1]
		}
		if f.CompOp == cartocss.REGEX {
			parts = append(parts, "(["+field+"].match("+value+"))")
		} else {
			parts = append(parts, "(["+field+"] "+f.CompOp.String()+" "+value+")")
		}
	}

	s := strings.Join(parts, " and ")
	if len(filters) > 1 {
		s = "(" + s + ")"
	}
	return s
}

var webmercZoomScales = []int{
	500000000,
	200000000,
	100000000,
	50000000,
	25000000,
	12500000,
	6500000,
	3000000,
	1500000,
	750000,
	400000,
	200000,
	100000,
	50000,
	25000,
	12500,
	5000,
	2500,
	1500,
	750,
	500,
	250,
	100,
}
