package cartocss

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
	"os"
	"sort"
)

var debugRules = 0

type Selector struct {
	Layer      string
	Class      string
	Attachment string
	Zoom       ZoomRange
	Filters    []Filter
}

type Filter struct {
	Field  string
	CompOp CompOp
	Value  Value
}

func (f Filter) String() string {
	return fmt.Sprintf("%s %s %v", f.Field, f.CompOp, f.Value)
}

type byField []Filter

func (f byField) Len() int      { return len(f) }
func (f byField) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f byField) Less(i, j int) bool {
	return f[i].Field < f[j].Field
}

type bySpecifity struct {
	rules       []Rule
	attachments map[string]int
}

type specificity struct {
	layer   int
	class   int
	filters int
	index   int
}

func (s specificity) less(other specificity) bool {
	if s.layer != other.layer {
		return s.layer < other.layer
	}
	if s.class != other.class {
		return s.class < other.class
	}
	if s.filters != other.filters {
		return s.filters < other.filters
	}
	return s.index < other.index
}

func (r *Rule) specificity() specificity {
	s := specificity{}
	if r.Layer != "" {
		s.layer += 1
	}
	if r.Class != "" {
		s.class += 1
	}
	s.filters += len(r.Filters)
	if r.Zoom != AllZoom {
		s.filters += 1
	}
	return s
}

func (s bySpecifity) Len() int      { return len(s.rules) }
func (s bySpecifity) Swap(i, j int) { s.rules[i], s.rules[j] = s.rules[j], s.rules[i] }
func (s bySpecifity) Less(i, j int) bool {
	if s.rules[i].Layer != s.rules[j].Layer {
		return s.rules[i].Layer < s.rules[j].Layer
	}
	if s.rules[i].Attachment != s.rules[j].Attachment {
		return s.attachments[s.rules[i].Attachment] > s.attachments[s.rules[j].Attachment]
	}
	if len(s.rules[i].Filters) != len(s.rules[j].Filters) {
		return len(s.rules[i].Filters) < len(s.rules[j].Filters)
	}
	if s.rules[i].Zoom != s.rules[j].Zoom {
		if s.rules[i].Zoom.Levels() != s.rules[j].Zoom.Levels() {
			return s.rules[i].Zoom.Levels() > s.rules[j].Zoom.Levels()
		}
		return s.rules[i].Zoom < s.rules[j].Zoom
	}
	return s.rules[i].order < s.rules[j].order
}

func filterIsSubset(a, b []Filter) bool {
	if len(a) > len(b) {
		return false
	}
	var ib int
	for ia := range a {
		found := false
		for ; ib < len(b); ib++ {
			if a[ia].Field > b[ib].Field {
				continue
			}
			if a[ia].Field < b[ib].Field {
				return false
			}

			if a[ia].CompOp == b[ib].CompOp && a[ia].Value == b[ib].Value {
				found = true
				break
			}
			if filterContains(a[ia], b[ib]) {
				found = true
				break
			}
			return false
		}
		if !found {
			return false
		}
	}
	return true
}

func filterContains(a, b Filter) bool {
	var av, bv float64
	var ok bool

	av, ok = a.Value.(float64)
	if !ok {
		tmp, ok := a.Value.(int)
		if !ok {
			return false
		}
		av = float64(tmp)
	}
	bv, ok = b.Value.(float64)
	if !ok {
		tmp, ok := b.Value.(int)
		if !ok {
			return false
		}
		bv = float64(tmp)
	}
	switch a.CompOp {
	case GT:
		if b.CompOp == EQ && bv > av {
			return true
		}
		if b.CompOp == GT && bv >= av {
			return true
		}
		if b.CompOp == GTE && bv >= av {
			return true
		}
	case GTE:
		if b.CompOp == EQ && bv >= av {
			return true
		}
		if b.CompOp == GT && bv >= av {
			return true
		}
		if b.CompOp == GTE && bv >= av {
			return true
		}
	case LT:
		if b.CompOp == EQ && bv < av {
			return true
		}
		if b.CompOp == LT && bv <= av {
			return true
		}
		if b.CompOp == LTE && bv <= av {
			return true
		}
	case LTE:
		if b.CompOp == EQ && bv <= av {
			return true
		}
		if b.CompOp == LT && bv <= av {
			return true
		}
		if b.CompOp == LTE && bv <= av {
			return true
		}
	}
	return false
}

func filterOverlap(a, b []Filter) bool {
	for ia := range a {
		for ib := range b {
			if a[ia].Field != b[ib].Field {
				continue
			}
			if a[ia].CompOp != b[ib].CompOp || a[ia].Value != b[ib].Value {
				return false
			} else {
				break
			}
		}
	}
	return true
}

func filterEqual(a, b []Filter) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Field != b[i].Field {
			return false
		}
		if a[i].CompOp != b[i].CompOp {
			return false
		}
		if a[i].Value != b[i].Value {
			return false
		}
	}
	return true
}

type Rule struct {
	Layer      string
	Attachment string
	Class      string
	Filters    []Filter
	Zoom       ZoomRange
	Properties *Properties
	order      int
}

func (r *Rule) hash() uint64 {
	h := fnv.New64()
	h.Write([]byte(r.Layer))
	h.Write([]byte(r.Attachment))
	h.Write([]byte(r.Class))
	binary.Write(h, binary.LittleEndian, r.Zoom)
	for i := range r.Filters {
		h.Write([]byte(r.Filters[i].String()))
	}
	return h.Sum64()
}

func (r *Rule) String() string {
	return fmt.Sprintf("Rule{%#v %#v %#v %v %v %s}", r.Layer, r.Attachment, r.Class, r.Filters, r.Zoom, r.Properties.String())
}

func (r Rule) childOf(o Rule) bool {
	if !(r.Layer == o.Layer || o.Layer == "") {
		return false
	}
	if !(r.Attachment == o.Attachment || o.Attachment == "") {
		return false
	}
	if !(r.Class == o.Class || o.Class == "") {
		return false
	}
	if !(r.Zoom&o.Zoom == r.Zoom || o.Zoom == AllZoom) {
		return false
	}
	if !filterIsSubset(o.Filters, r.Filters) {
		return false
	}
	return true
}

func (r Rule) same(o Rule) bool {
	if r.Layer != o.Layer {
		return false
	}
	if r.Attachment != o.Attachment {
		return false
	}
	if r.Class != o.Class {
		return false
	}
	if r.Zoom != o.Zoom {
		return false
	}
	if !filterEqual(r.Filters, o.Filters) {
		return false
	}
	return true
}

func (r Rule) sameExceptClass(o Rule) bool {
	if r.Layer != o.Layer {
		return false
	}
	if r.Attachment != o.Attachment {
		return false
	}
	if r.Zoom != o.Zoom {
		return false
	}
	if !filterEqual(r.Filters, o.Filters) {
		return false
	}
	return true
}

func (r Rule) overlaps(o Rule) bool {
	if !(r.Layer == o.Layer || o.Layer == "") {
		return false
	}
	if !(r.Attachment == o.Attachment || o.Attachment == "") {
		return false
	}
	if !(r.Zoom.combine(o.Zoom).Levels() > 0 || r.Zoom == o.Zoom) {
		return false
	}
	if !filterOverlap(o.Filters, r.Filters) {
		return false
	}
	return true
}

func (m *MSS) Layers() []string {
	layerNames := []string{}
	layersAdded := map[string]struct{}{}
	for _, b := range m.root.blocks {
		for _, s := range b.selectors {
			if _, ok := layersAdded[s.Layer]; !ok {
				layerNames = append(layerNames, s.Layer)
				layersAdded[s.Layer] = struct{}{}
			}
		}
	}
	return layerNames
}

func (m *MSS) LayerRules(layer string, cssIds []string, classes ...string) []Rule {
	return m.LayerZoomRules(layer, cssIds, InvalidZoom, classes...)
}

func (m *MSS) LayerZoomRules(layer string, cssIds []string, zoom ZoomRange, classes ...string) []Rule {
	attachments := make(map[string]int)
	rules := []Rule{}
	order := 1
	var collect func(*block, Rule)
	collect = func(node *block, parent Rule) {
		if len(node.selectors) == 0 {
			for _, n := range node.blocks {
				collect(n, parent)
			}
		}

		for _, s := range node.selectors {
			current := Rule{
				Layer:      parent.Layer,
				Class:      parent.Class,
				Attachment: parent.Attachment,
				Filters:    append([]Filter{}, parent.Filters...),
				Zoom:       parent.Zoom,
			}
			if s.Layer != "" {
				for _, id := range cssIds {
					if s.Layer != id {
						continue
					}
					current.Layer = s.Layer
					break
				}
			}
			if current.Layer == "" {
				continue
			}
			foundClass := false
			if s.Class != "" {
				for _, c := range classes {
					if c == s.Class {
						foundClass = true
						break
					}
				}
				if !foundClass {
					continue
				}
				current.Class = s.Class
			}
			if s.Attachment != "" {
				if _, ok := attachments[s.Attachment]; !ok {
					attachments[s.Attachment] = order
				}
				current.Attachment = s.Attachment
			}
			if s.Filters != nil {
				sort.Sort(byField(s.Filters))
				f, ok := mergeFilters(current.Filters, s.Filters)
				if !ok {
					continue
				}
				current.Filters = f
			}

			if s.Zoom != InvalidZoom {
				if current.Zoom != 0 {
					current.Zoom = current.Zoom.combine(s.Zoom)
					if current.Zoom == InvalidZoom {
						continue
					}
				} else {
					current.Zoom = s.Zoom
				}
			}

			if node.properties != nil && !node.properties.isEmpty() {
				order += 1
				r := Rule{
					Layer:      current.Layer,
					Class:      current.Class,
					Attachment: current.Attachment,
					Filters:    append([]Filter{}, current.Filters...),
					Zoom:       current.Zoom,
					Properties: node.properties.clone(),
					order:      order,
				}
				spec := r.specificity()
				for _, k := range r.Properties.keys() {
					r.Properties.setSpecificity(k, spec)
				}
				rules = append(rules, r)
			}
			for _, n := range node.blocks {
				collect(n, current)
			}
		}
	}
	collect(&m.root, Rule{Zoom: zoom})
	if len(rules) > 0 {
		rules = sortedRules(rules, attachments, classes)
	}
	for i := range rules {
		if rules[i].Layer == "" {
			rules[i].Layer = layer
		}
	}

	return rules
}

func combineRules(a, b Rule) Rule {
	r := Rule{
		Layer:      a.Layer,
		Class:      a.Class,
		Attachment: a.Attachment,
		Zoom:       a.Zoom.combine(b.Zoom),
	}

	r.Filters = combineFilters(a.Filters, b.Filters)
	r.Properties = combineProperties(a.Properties, b.Properties)

	if debugRules >= 1 {
		fmt.Fprintln(os.Stderr, "        a", a)
		fmt.Fprintln(os.Stderr, "      + b", b)
		fmt.Fprintln(os.Stderr, "      ===", r)
	}

	return r
}

func combineFilters(a, b []Filter) []Filter {
	combined := make([]Filter, len(a))
	copy(combined, a)
nextFilter:
	for _, f := range b {
		for _, c := range combined {
			if f.Field == c.Field {
				continue nextFilter
			}
		}
		combined = append(combined, f)
	}
	sort.Sort(byField(combined))
	return combined
}

func extendRule(base Rule, rules []Rule, pos int) (int, []Rule) {
	var addedTotal, added int
	if newRules := fillProperties(base, rules[pos+1:]); len(newRules) > 0 {
		for _, r := range newRules {
			added, rules = extendRule(r, rules, pos)
			addedTotal += added
		}
		rules = append(rules[:pos+addedTotal], append(newRules, rules[pos+addedTotal:]...)...)
		addedTotal += len(newRules)
	}
	return addedTotal, rules
}

func fillProperties(r Rule, subRules []Rule) []Rule {
	newRules := []Rule{}
	for _, o := range subRules {
		if debugRules >= 2 {
			fmt.Fprintln(os.Stderr, " compare ", r, o)
		}

		if r.same(o) {
			r.Properties.updateMissing(o.Properties)
			continue
		} else if r.childOf(o) {
			// e.g. {a=1, b=1}.chilldOf{b=1} -> add missing properties
			if debugRules >= 1 {
				fmt.Fprintln(os.Stderr, " child of", r, o)
			}
			r.Properties.updateMissing(o.Properties)
		} else if r.overlaps(o) {
			// {a=1, b=1}.overlaps{c=1} -> create new combined rule
			if debugRules >= 1 {
				fmt.Fprintln(os.Stderr, " overlaps", r, o)
			}
			newRule := combineRules(r, o)
			if o.same(newRule) {
				o.Properties.updateMissing(newRule.Properties)
			} else if r.same(newRule) {
				r.Properties.updateMissing(newRule.Properties)
			} else {
				dup := false
				for i, nr := range newRules {
					if newRule.same(nr) {
						newRules[i].Properties.updateMissing(nr.Properties)
						dup = true
						break
					}
				}
				if !dup {
					for i, nr := range subRules {
						if newRule.same(nr) {
							subRules[i].Properties.updateMissing(nr.Properties)
							dup = true
							break
						}
					}
					if !dup {
						newRules = append(newRules, newRule)
					}
				}
			}
		}
	}
	return newRules
}

func sortedRules(rules []Rule, attachments map[string]int, classes []string) []Rule {
	if len(rules) == 0 {
		return nil
	}
	sort.Sort(sort.Reverse(bySpecifity{rules: rules, attachments: attachments}))

	if debugRules >= 1 {
		fmt.Fprintln(os.Stderr, "sorted ruled, before fillProperties")
		for _, rr := range rules {
			fmt.Fprintln(os.Stderr, "  ", rr)
		}
		fmt.Fprintln(os.Stderr, "\nfilling rules")
	}

	var added int
	for pos := 0; pos < len(rules); {
		if debugRules >= 1 {
			fmt.Fprintln(os.Stderr, "pre-extend")
			for i, rr := range rules {
				if i == pos {
					fmt.Fprintln(os.Stderr, "* ", rr)
				} else {
					fmt.Fprintln(os.Stderr, "  ", rr)
				}
			}
		}
		added, rules = extendRule(rules[pos], rules, pos)
		pos += added
		if debugRules >= 1 {
			fmt.Fprintln(os.Stderr, "post-extend")
			for i, rr := range rules {
				if i == pos {
					fmt.Fprintln(os.Stderr, "* ", rr)
				} else {
					fmt.Fprintln(os.Stderr, "  ", rr)
				}
			}
			fmt.Fprintln(os.Stderr)
		}
		pos++
	}

	if len(classes) > 0 {
		return dedupMergeClasses(rules, classes)
	}
	return dedup(rules)
}

func dedup(rules []Rule) []Rule {
	added := make(map[uint64]struct{}, len(rules))
	result := []Rule{}
	for i := range rules {
		hash := rules[i].hash()
		if _, ok := added[hash]; !ok {
			result = append(result, rules[i])
			added[hash] = struct{}{}
		}
	}
	return result
}

func dedupMergeClasses(rules []Rule, classes []string) []Rule {
	classIdx := func(class string) int {
		if class == "" {
			return math.MaxInt32
		}
		for i := range classes {
			if classes[i] == class {
				return i
			}
		}
		return math.MaxInt32
	}

	result := []Rule{}
	for i := range rules {
		found := false
		for j := range result {
			if rules[i].sameExceptClass(result[j]) {
				classIdxI := classIdx(rules[i].Class)
				classIdxJ := classIdx(result[j].Class)

				if classIdxI < classIdxJ {
					rules[i].Properties.updateMissing(result[j].Properties)
					result[j] = rules[i]
				} else if classIdxJ < classIdxI {
					result[j].Properties.updateMissing(rules[i].Properties)
				}
				found = true
				break
			}
		}
		if !found {
			result = append(result, rules[i])
		}
	}
	return result
}

func assertSortedFilters(filters []Filter) {
	for i := range filters {
		if i == 0 {
			continue
		}
		if filters[i-1].Field > filters[i].Field {
			panic(fmt.Sprintf("filter not sorted: %v\n", filters))
		}
	}
}

func mergeFilters(a, b []Filter) ([]Filter, bool) {
	result := make([]Filter, 0, len(a)+len(b))

	var ai, bi int

	for ai < len(a) && bi < len(b) {
		if a[ai].Field < b[bi].Field {
			result = append(result, a[ai])
			ai++
			continue
		} else if b[bi].Field < a[ai].Field {
			result = append(result, b[bi])
			bi++
			continue
		} else {
			f, ok := mergeFilter(a[ai], b[bi])
			if !ok {
				return nil, false
			}
			result = append(result, f)
			ai++
			bi++
		}
	}

	result = append(result, a[ai:]...)
	result = append(result, b[bi:]...)

	return result, true
}

func mergeFilter(a, b Filter) (Filter, bool) {
	if a.Field != b.Field {
		return Filter{}, false
	}
	if a.CompOp == b.CompOp && a.Value == b.Value {
		return a, true
	}
	if a.CompOp == LT {
		a.CompOp = LTE
		a.Value = a.Value.(float64) - 1
	}
	if a.CompOp == GT {
		a.CompOp = GTE
		a.Value = a.Value.(float64) + 1
	}
	if b.CompOp == LT {
		b.CompOp = LTE
		b.Value = b.Value.(float64) - 1
	}
	if b.CompOp == GT {
		b.CompOp = GTE
		b.Value = b.Value.(float64) + 1
	}

	if a.CompOp == LTE && b.CompOp == LTE {
		if a.Value.(float64) < b.Value.(float64) {
			return Filter{Field: a.Field, CompOp: LTE, Value: a.Value}, true
		}
		return Filter{Field: a.Field, CompOp: LTE, Value: b.Value}, true
	}
	if a.CompOp == GTE && b.CompOp == GTE {
		if a.Value.(float64) > b.Value.(float64) {
			return Filter{Field: a.Field, CompOp: GTE, Value: a.Value}, true
		}
		return Filter{Field: a.Field, CompOp: GTE, Value: b.Value}, true
	}

	return Filter{}, false
}

func RulesZoom(rs []Rule) ZoomRange {
	z := InvalidZoom
	for _, r := range rs {
		z = ZoomRange(r.Zoom | z)
	}
	return z

}
