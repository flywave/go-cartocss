package mapnik

import (
	"strings"

	"github.com/flywave/go-cartocss"
)

func filterItems(rules []cartocss.Rule) map[string]map[string]struct{} {
	result := make(map[string]map[string]struct{})
	for _, r := range rules {
		if len(r.Filters) == 0 {
			return nil
		}

		found := false
		for _, f := range r.Filters {
			if f.CompOp != cartocss.EQ {
				continue
			}
			v, ok := f.Value.(string)
			if !ok {
				continue
			}
			found = true
			if result[f.Field] == nil {
				result[f.Field] = make(map[string]struct{})
			}
			result[f.Field][v] = struct{}{}
		}
		if !found {
			return nil
		}
	}
	return result
}

func FilterString(rules []cartocss.Rule) string {
	var vals, parts []string

	items := filterItems(rules)

	for key, values := range items {
		vals = vals[:0]
		for v, _ := range values {
			vals = append(vals, "'"+v+"'")
		}
		parts = append(parts, "\""+key+"\" IN ("+strings.Join(vals, ", ")+")")
	}
	if parts == nil {
		return ""
	}
	return "(" + strings.Join(parts, " OR ") + ")"
}

func WrapWhere(query, where string) string {
	if where == "" {
		return query
	}
	return "(SELECT * FROM " + query + " WHERE " + where + ") as filtered"
}
