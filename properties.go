package cartocss

import (
	"bytes"
	"math"
	"sort"
	"strings"

	"github.com/flywave/go-cartocss/color"

	"fmt"
)

type attr struct {
	value       Value
	pos         position
	specificity specificity
}

type key struct {
	name     string
	instance string
}

type Properties struct {
	values          map[key]attr
	defaultInstance string
}

func (p *Properties) String() string {
	var buf bytes.Buffer
	buf.WriteString("Properties{")
	for k, v := range p.values {
		if k.instance != "" {
			buf.WriteString(k.instance)
			buf.WriteRune('/')
		}
		fmt.Fprintf(&buf, "%s: %#v", k.name, v.value)
	}
	buf.WriteRune('}')
	return buf.String()
}

func (p *Properties) get(name string) (Value, bool) {
	v, ok := p.values[key{name: name, instance: p.defaultInstance}]
	return v.value, ok
}

func (p *Properties) getKey(property key) Value {
	return p.values[property].value
}

// pos returns position of the property in the .mss file.
func (p *Properties) pos(property key) position {
	return p.values[property].pos
}

func (p *Properties) isEmpty() bool {
	return len(p.values) == 0
}

func (p *Properties) set(property string, val Value) {
	if p.values == nil {
		p.values = make(map[key]attr)
	}
	p.values[key{name: property}] = attr{value: val}
}

func (p *Properties) setKey(property key, val Value) {
	if p.values == nil {
		p.values = make(map[key]attr)
	}
	p.values[property] = attr{value: val}
}

func (p *Properties) setPos(property key, val Value, pos position) {
	if p.values == nil {
		p.values = make(map[key]attr)
	}
	p.values[property] = attr{value: val, pos: pos, specificity: specificity{index: pos.index}}
}

func (p *Properties) setSpecificity(property key, specificity specificity) {
	a := p.values[property]
	index := a.specificity.index // keep existing index
	a.specificity = specificity
	a.specificity.index = index
	p.values[property] = a
}

func (p *Properties) updateMissing(o *Properties) {
	for k, v := range o.values {
		existing, ok := p.values[k]
		if !ok || existing.specificity.less(v.specificity) {
			p.values[k] = v
		}
	}
}

func (p *Properties) keys() []key {
	keys := make([]key, len(p.values))
	i := 0
	for k := range p.values {
		keys[i] = k
		i += 1
	}
	return keys
}

func (p *Properties) clone() *Properties {
	result := Properties{}
	for k, v := range p.values {
		result.setPos(k, v.value, v.pos) // XXX
	}
	return &result
}

func (p *Properties) minPos() int {
	index := math.MaxInt32
	for _, v := range p.values {
		if v.specificity.index < index {
			index = v.specificity.index
		}
	}
	return index
}

func (p *Properties) minPrefixPos(prefix string) []prefixPos {
	instanceIndex := map[string]int{}
	for k, v := range p.values {
		if strings.HasPrefix(k.name, prefix) {
			index, ok := instanceIndex[k.instance]
			if !ok || v.specificity.index < index {
				instanceIndex[k.instance] = v.specificity.index
			}
		}
	}
	result := make([]prefixPos, 0, len(instanceIndex))
	for instance, index := range instanceIndex {
		result = append(result, prefixPos{prefix: prefix, instance: instance, index: index})
	}
	return result
}

type prefixPos struct {
	prefix   string
	instance string
	index    int
}

type byIndex []prefixPos

func (p byIndex) Len() int      { return len(p) }
func (p byIndex) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p byIndex) Less(i, j int) bool {
	if p[i].index != p[j].index {
		return p[i].index < p[j].index
	}

	return len(p[i].prefix) > len(p[j].prefix)
}

type Prefix struct {
	Name     string
	Instance string
}

func SortedPrefixes(p *Properties, prefixes []string) []Prefix {
	pp := make([]prefixPos, 0, len(prefixes))
	for _, prefix := range prefixes {
		pos := p.minPrefixPos(prefix)
		pp = append(pp, pos...)
	}
	sort.Sort(byIndex(pp))

	result := make([]Prefix, len(pp))
	for i, p := range pp {
		result[i] = Prefix{Name: p.prefix, Instance: p.instance}
	}
	return result
}

func (p *Properties) GetKV() ([]string, []Value) {
	ks := []string{}
	vs := []Value{}
	for k, v := range p.values {
		ks = append(ks, k.name)
		vs = append(vs, v.value)
	}
	return ks, vs
}

func (p *Properties) SetDefaultInstance(instance string) {
	p.defaultInstance = instance
}

func (p *Properties) GetBool(property string) (bool, bool) {
	v, ok := p.get(property)
	if !ok {
		return false, false
	}
	r, ok := v.(bool)
	return r, ok
}

func (p *Properties) GetString(property string) (string, bool) {
	v, ok := p.get(property)
	if !ok {
		return "", false
	}
	r, ok := v.(string)
	return r, ok
}

func (p *Properties) GetFloat(property string) (float64, bool) {
	v, ok := p.get(property)
	if !ok {
		return 0, false
	}
	r, ok := v.(float64)
	return r, ok
}

func (p *Properties) GetColor(property string) (color.Color, bool) {
	v, ok := p.get(property)
	if !ok {
		return color.Color{}, false
	}
	r, ok := v.(color.Color)
	return r, ok
}

func (p *Properties) GetFloatList(property string) ([]float64, bool) {
	v, ok := p.get(property)
	if !ok {
		return nil, false
	}
	l, ok := v.([]Value)
	if !ok {
		return nil, false
	}
	nums := make([]float64, len(l))
	for i := range l {
		nums[i], ok = l[i].(float64)
		if !ok {
			return nil, false
		}
	}
	return nums, true
}

func (p *Properties) GetStringList(property string) ([]string, bool) {
	v, ok := p.get(property)
	if !ok {
		return nil, false
	}
	if s, ok := v.(string); ok {
		return []string{s}, true
	}
	l, ok := v.([]Value)
	if !ok {
		return nil, false
	}
	strs := make([]string, len(l))
	for i := range l {
		strs[i], ok = l[i].(string)
		if !ok {
			return nil, false
		}
	}
	return strs, true
}

func (p *Properties) GetFieldList(property string) ([]interface{}, bool) {
	v, ok := p.get(property)
	if !ok {
		return nil, false
	}
	if s, ok := v.(string); ok {
		return []interface{}{Field(s)}, true
	}
	l, ok := v.([]Value)
	if !ok {
		return nil, false
	}
	vals := make([]interface{}, len(l))
	for i := range l {
		vals[i] = l[i]
	}
	return vals, true
}

func (p *Properties) GetStopList(property string) ([]Stop, bool) {
	v, ok := p.get(property)
	if !ok {
		return nil, false
	}
	l, ok := v.([]Value)
	if !ok {
		return nil, false
	}
	stops := make([]Stop, len(l))
	for i := range l {
		stops[i], ok = l[i].(Stop)
		if !ok {
			return nil, false
		}
	}
	return stops, true
}

func combineProperties(a, b *Properties) *Properties {
	r := &Properties{values: make(map[key]attr)}
	for k, v := range a.values {
		r.values[k] = v
	}
	r.updateMissing(b)
	return r
}

var propCounter int

func NewProperties(kv ...interface{}) *Properties {
	r := &Properties{}
	r.values = make(map[key]attr)
	for i := 0; i < (len(kv) - 1); i += 2 {
		k := kv[i].(string)
		v := kv[i+1]
		r.values[key{name: k}] = attr{value: v, specificity: specificity{index: propCounter}}
		propCounter += 1
	}
	return r
}

func NewPropertiesInstance(kiv ...interface{}) *Properties {
	r := &Properties{}
	r.values = make(map[key]attr)
	for i := 0; i < (len(kiv) - 2); i += 3 {
		k := kiv[i].(string)
		instance := kiv[i+1].(string)
		v := kiv[i+2]
		r.values[key{name: k, instance: instance}] = attr{value: v, specificity: specificity{index: propCounter}}
		propCounter += 1
	}
	return r
}
