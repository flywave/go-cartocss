package mapnik

import (
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
// This is useful for pre-filtering in key-value or column-family stores.
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

type filterKey struct {
	field    string
	op       cartocss.CompOp
	valStr   string
	valFloat float64
	modVal   *cartocss.ModuloComparsion
}
