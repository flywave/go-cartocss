package mapnik

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	cartocss "github.com/flywave/go-cartocss"
)

// FilterCondition represents a single filter extracted from CartoCSS rules.
type FilterCondition struct {
	Field string
	Op    cartocss.CompOp
	Value interface{}
}

// FilterSet holds all filter conditions extracted from a set of CartoCSS rules,
// organized for efficient field lookup. This is data-source agnostic — consumers
// can use it to pre-filter features regardless of storage backend.
type FilterSet struct {
	fields map[string][]FilterCondition
}

// NewFilterSet extracts filter conditions from a slice of CartoCSS rules.
// Conditions are deduplicated across rules.
func NewFilterSet(rules []cartocss.Rule) *FilterSet {
	fs := &FilterSet{fields: make(map[string][]FilterCondition)}
	seen := make(map[string]map[filterKey]struct{})

	for _, r := range rules {
		for _, f := range r.Filters {
			k := filterKey{field: f.Field, op: f.CompOp}
			switch v := f.Value.(type) {
			case string:
				k.valStr = v
			case float64:
				k.valFloat = v
			case int:
				k.valFloat = float64(v)
			case cartocss.ModuloComparsion:
				k.modVal = &v
			}

			if _, ok := seen[f.Field]; !ok {
				seen[f.Field] = make(map[filterKey]struct{})
			}
			if _, ok := seen[f.Field][k]; ok {
				continue
			}
			seen[f.Field][k] = struct{}{}

			fs.fields[f.Field] = append(fs.fields[f.Field], FilterCondition{
				Field: f.Field,
				Op:    f.CompOp,
				Value: f.Value,
			})
		}
	}
	for _, conds := range fs.fields {
		if len(conds) == 0 {
			delete(fs.fields, conds[0].Field)
		}
	}
	return fs
}

// Fields returns all field names that have at least one filter condition.
func (fs *FilterSet) Fields() []string {
	keys := make([]string, 0, len(fs.fields))
	for k := range fs.fields {
		keys = append(keys, k)
	}
	return keys
}

// Conditions returns all filter conditions for all fields.
func (fs *FilterSet) Conditions() []FilterCondition {
	var all []FilterCondition
	for _, conds := range fs.fields {
		all = append(all, conds...)
	}
	return all
}

// FieldConditions returns all filter conditions for a specific field.
func (fs *FilterSet) FieldConditions(field string) []FilterCondition {
	return fs.fields[field]
}

// FieldValues returns all distinct equality (=) values for a given field.
func (fs *FilterSet) FieldValues(field string) []interface{} {
	conds := fs.fields[field]
	if len(conds) == 0 {
		return nil
	}
	seen := make(map[interface{}]struct{})
	var vals []interface{}
	for _, c := range conds {
		if c.Op != cartocss.EQ {
			continue
		}
		if _, ok := seen[c.Value]; ok {
			continue
		}
		seen[c.Value] = struct{}{}
		vals = append(vals, c.Value)
	}
	return vals
}

// HasField reports whether the given field has any filter conditions.
func (fs *FilterSet) HasField(field string) bool {
	_, ok := fs.fields[field]
	return ok
}

// FilteredFieldCount returns the number of distinct fields with filters.
func (fs *FilterSet) FilteredFieldCount() int {
	return len(fs.fields)
}

// ConditionCount returns the total number of filter conditions.
func (fs *FilterSet) ConditionCount() int {
	n := 0
	for _, conds := range fs.fields {
		n += len(conds)
	}
	return n
}

// Matches tests whether a set of attribute values satisfies all filter conditions.
// Returns true if there are no filter conditions (unfiltered).
// val can be string, float64, int, or nil.
func (fs *FilterSet) Matches(featureAttrs map[string]interface{}) bool {
	if len(fs.fields) == 0 {
		return true
	}
	for field, conds := range fs.fields {
		val, ok := featureAttrs[field]
		if !ok {
			return false
		}
		for _, c := range conds {
			if !compareFilter(val, c.Op, c.Value) {
				return false
			}
		}
	}
	return true
}

func compareFilter(actual interface{}, op cartocss.CompOp, expected interface{}) bool {
	switch op {
	case cartocss.EQ:
		return reflect.DeepEqual(actual, expected)
	case cartocss.NEQ:
		return !reflect.DeepEqual(actual, expected)
	case cartocss.GT, cartocss.GTE, cartocss.LT, cartocss.LTE:
		a, aok := toFloat(actual)
		e, eok := toFloat(expected)
		if !aok || !eok {
			return false
		}
		switch op {
		case cartocss.GT:
			return a > e
		case cartocss.GTE:
			return a >= e
		case cartocss.LT:
			return a < e
		case cartocss.LTE:
			return a <= e
		}
	case cartocss.REGEX:
		// string regex matching not implemented at filter level
		return true
	case cartocss.MODULO:
		// modulo comparison not implemented at filter level
		return true
	}
	return false
}

func toFloat(v interface{}) (float64, bool) {
	switch v := v.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	}
	return 0, false
}

type filterKey struct {
	field    string
	op       cartocss.CompOp
	valStr   string
	valFloat float64
	modVal   *cartocss.ModuloComparsion
}

// --- Data source organization ---

// RuleEntry pairs a cartocss Rule with its derived FilterSet.
// Index preserves the original rule order for style/ruleset construction.
type RuleEntry struct {
	Rule      cartocss.Rule
	FilterSet *FilterSet
	Index     int
}

// FilterGroup collects rules that share the same filter conditions.
// Consumers can create a single filtered data source per group.
type FilterGroup struct {
	// Label is a human-readable name for this filter group (e.g. "type=motorway").
	Label string
	// FilterSet contains the combined filter conditions for this group.
	FilterSet *FilterSet
	// RuleEntries lists the rules that belong to this group.
	RuleEntries []RuleEntry
}

// LayerFilterSet organizes all rules for one layer by filter conditions.
// Rules without filters go into a single group; rules with different filter
// combinations each get their own group.
type LayerFilterSet struct {
	LayerID string
	Groups  []FilterGroup
}

// NewLayerFilterSet analyzes rules for a single layer and groups them by
// their filter conditions. This is the primary entry point for datasource
// organization — consumers can create one in-memory datasource per group.
func NewLayerFilterSet(layerID string, rules []cartocss.Rule) *LayerFilterSet {
	lfs := &LayerFilterSet{LayerID: layerID}
	groupMap := make(map[string]*FilterGroup)
	groupOrder := []string{}

	for i, r := range rules {
		entry := RuleEntry{
			Rule:      r,
			FilterSet: NewFilterSet([]cartocss.Rule{r}),
			Index:     i,
		}

		label := buildFilterLabel(r.Filters)
		if g, ok := groupMap[label]; ok {
			g.RuleEntries = append(g.RuleEntries, entry)
		} else {
			g := &FilterGroup{
				Label:       label,
				FilterSet:   entry.FilterSet,
				RuleEntries: []RuleEntry{entry},
			}
			groupMap[label] = g
			groupOrder = append(groupOrder, label)
		}
	}

	for _, label := range groupOrder {
		lfs.Groups = append(lfs.Groups, *groupMap[label])
	}
	return lfs
}

func buildFilterLabel(filters []cartocss.Filter) string {
	if len(filters) == 0 {
		return "*"
	}
	parts := make([]string, 0, len(filters))
	for _, f := range filters {
		parts = append(parts, f.Field+" "+f.CompOp.String()+" "+fmtFilterValue(f.Value))
	}
	return strings.Join(parts, " & ")
}

func fmtFilterValue(v interface{}) string {
	switch v := v.(type) {
	case string:
		return "'" + v + "'"
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", v)
	}
}
