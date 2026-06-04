package mapnik

import (
	cartocss "github.com/flywave/go-cartocss"

	"github.com/flywave/go-cartocss/builder"
	"github.com/flywave/go-cartocss/config"
	nik "github.com/flywave/flywave-mapnik"
)

type Map struct {
	inner          *nik.Map
	locator        config.Locator
	scaleFactor    float64
	autoTypeFilter bool
	zoomScales     []int
}

type maker struct{}

func (m maker) Type() string       { return "mapnik" }
func (m maker) FileSuffix() string { return ".xml" }
func (m maker) New(locator config.Locator) builder.MapWriter {
	return New(locator)
}

var Maker3 = maker{}

func New(locator config.Locator) *Map {
	fm := nik.New()
	fm.SetSRS("epsg:3857")
	return &Map{
		inner:       fm,
		locator:     locator,
		scaleFactor: 1.0,
		zoomScales:  webmercZoomScales,
	}
}

func (m *Map) SetAutoTypeFilter(enable bool) {
	m.autoTypeFilter = enable
}

func (m *Map) SetZoomScales(zoomScales []int) {
	m.zoomScales = zoomScales
}

func (m *Map) getNativeMap() *nik.Map {
	return m.inner
}

func (m *Map) AddLayer(l cartocss.Layer, rules []cartocss.Rule) {
	if l.ScaleFactor != 0.0 {
		prevScaleFactor := m.scaleFactor
		defer func() { m.scaleFactor = prevScaleFactor }()
		m.scaleFactor = l.ScaleFactor
	}

	styleMap := m.buildStyles(rules)

	var styleNames []string
	for name, style := range styleMap {
		fs := nik.NewStyle()
		fs.SetFilterMode(1)

		if style.compOp != "" {
			fs.SetCompOp(compOpToInt(style.compOp))
		}
		if style.opacity > 0 {
			fs.SetOpacity(float32(style.opacity))
		}

		for _, r := range style.rules {
			fs.AddRule(r)
		}

		if err := m.inner.InsertStyle(name, fs); err != nil {
			panic("insert style " + name + ": " + err.Error())
		}
		styleNames = append(styleNames, name)
	}

	srs := l.SRS
	layer := nik.NewLayer(l.ID, srs)
	if !l.Active {
		layer.SetSrs(srs)
	}

	z := cartocss.RulesZoom(rules)
	if z != cartocss.AllZoom {
		if first := z.First(); first > 0 && first <= len(m.zoomScales) {
			layer.SetMaxZoom(float64(m.zoomScales[first-1]))
		}
		if last := z.Last(); last < len(m.zoomScales) {
			layer.SetMinZoom(float64(m.zoomScales[last]))
		}
	}

	ds := m.newDatasource(l.Datasource, rules)
	if ds != nil {
		layer.SetDatasource(ds)
	}
	for _, name := range styleNames {
		layer.AddStyle(name)
	}
	m.inner.AddLayer(layer)
}

type styleDesc struct {
	compOp string
	opacity float64
	rules   []*nik.Rule
}

func (m *Map) buildStyles(rules []cartocss.Rule) map[string]*styleDesc {
	styles := make(map[string]*styleDesc)
	var order []string

	for _, r := range rules {
		styleName := r.Layer
		if r.Attachment != "" {
			styleName += "-" + r.Attachment
		}

		sd, exists := styles[styleName]
		if !exists {
			sd = &styleDesc{}
			styles[styleName] = sd
			order = append(order, styleName)
		}
		if sd.compOp == "" {
			if v, ok := r.Properties.GetString("comp-op"); ok {
				sd.compOp = v
			}
		}
		if sd.opacity == 0 {
			if v, ok := r.Properties.GetFloat("opacity"); ok {
				sd.opacity = v
			}
		}

		fsRule := m.buildRule(r)
		if fsRule != nil {
			sd.rules = append(sd.rules, fsRule)
		}
	}
	return styles
}

func (m *Map) buildRule(r cartocss.Rule) *nik.Rule {
	var minScale, maxScale float64
	if l := r.Zoom.First(); l > 0 {
		if l > len(m.zoomScales) {
			l = len(m.zoomScales)
		}
		maxScale = float64(m.zoomScales[l-1])
	}
	if l := r.Zoom.Last(); l < len(m.zoomScales) {
		minScale = float64(m.zoomScales[l])
	}

	fr := nik.NewRule(r.Layer, minScale, maxScale)

	if filter := fmtFilters(r.Filters); filter != "" {
		fr.SetFilter(filter)
	}

	prefixes := cartocss.SortedPrefixes(r.Properties, []string{
		"line-", "line-pattern-", "polygon-", "polygon-pattern-",
		"text-", "shield-", "marker-", "point-", "building-", "dot-", "raster-",
	})

	for _, p := range prefixes {
		r.Properties.SetDefaultInstance(p.Instance)

		var sym *nik.Symbolizer
		switch p.Name {
		case "line-":
			sym = m.addLineSymbolizer(r)
		case "line-pattern-":
			sym = m.addLinePatternSymbolizer(r)
		case "polygon-":
			sym = m.addPolygonSymbolizer(r)
		case "polygon-pattern-":
			sym = m.addPolygonPatternSymbolizer(r)
		case "text-":
			sym = m.addTextSymbolizer(r)
		case "shield-":
			sym = m.addShieldSymbolizer(r)
		case "marker-":
			sym = m.addMarkerSymbolizer(r)
		case "point-":
			sym = m.addPointSymbolizer(r)
		case "building-":
			sym = m.addBuildingSymbolizer(r)
		case "dot-":
			sym = m.addDotSymbolizer(r)
		case "raster-":
			sym = m.addRasterSymbolizer(r)
		}
		if sym != nil {
			fr.Append(sym)
		}
	}
	r.Properties.SetDefaultInstance("")

	if fr != nil {
		return fr
	}
	return nil
}

func compOpToInt(op string) int {
	switch op {
	case "clear":
		return 0
	case "src":
		return 1
	case "dst":
		return 2
	case "src-over":
		return 3
	case "dst-over":
		return 4
	case "src-in":
		return 5
	case "dst-in":
		return 6
	case "src-out":
		return 7
	case "dst-out":
		return 8
	case "src-atop":
		return 9
	case "dst-atop":
		return 10
	case "xor":
		return 11
	case "plus":
		return 12
	case "minus":
		return 13
	case "multiply":
		return 14
	case "screen":
		return 15
	case "overlay":
		return 16
	case "darken":
		return 17
	case "lighten":
		return 18
	case "color-dodge":
		return 19
	case "color-burn":
		return 20
	case "hard-light":
		return 21
	case "soft-light":
		return 22
	case "difference":
		return 23
	case "exclusion":
		return 24
	case "contrast":
		return 25
	case "invert":
		return 26
	case "invert-rgb":
		return 27
	case "grain-merge":
		return 28
	case "grain-extract":
		return 29
	case "hue":
		return 30
	case "saturation":
		return 31
	case "color":
		return 32
	case "value":
		return 33
	default:
		return 3
	}
}

func (m *Map) newDatasource(ds cartocss.Datasource, rules []cartocss.Rule) *nik.Datasource {
	switch ds := ds.(type) {
	case *cartocss.Shapefile:
		fname := m.locator.Shape(ds.Filename)
		d, err := nik.NewFileDatasource("shape", fname)
		if err == nil && d != nil {
			return d
		}
	case *cartocss.GeoJson:
		fname := m.locator.Shape(ds.Filename)
		d, err := nik.NewFileDatasource("geojson", fname)
		if err == nil && d != nil {
			return d
		}
	case *cartocss.SQLite:
		fname := m.locator.SQLite(ds.Filename)
		d, err := nik.NewFileDatasource("sqlite", fname)
		if err == nil && d != nil {
			return d
		}
	case *cartocss.OGR:
		file := ds.Filename
		if !isOgrConnection.MatchString(ds.Filename) {
			file = m.locator.Data(ds.Filename)
		}
		d, err := nik.NewFileDatasource("ogr", file)
		if err == nil && d != nil {
			return d
		}
	case *cartocss.GDAL:
		file := m.locator.Data(ds.Filename)
		d, err := nik.NewFileDatasource("gdal", file)
		if err == nil && d != nil {
			return d
		}
	}
	return nil
}
