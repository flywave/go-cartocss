package mapnik

import (
	"fmt"
	"strings"

	imagecolor "image/color"

	cartocss "github.com/flywave/go-cartocss"

	nik "github.com/flywave/flywave-mapnik"
	"github.com/flywave/go-cartocss/builder"
	"github.com/flywave/go-cartocss/color"
	"github.com/flywave/go-cartocss/config"
)

type Map struct {
	inner          *nik.Map
	XML            *XMLMap
	fontSets       map[string]string
	locator        config.Locator
	scaleFactor    float64
	autoTypeFilter bool
	zoomScales     []int
	proj4          bool
}

type maker struct {
	proj4 bool
}

func (m maker) Type() string       { return "mapnik" }
func (m maker) FileSuffix() string { return ".xml" }
func (m maker) New(locator config.Locator) builder.MapWriter {
	mm := New(locator)
	mm.SetProj4(m.proj4)
	return mm
}

var Maker3 = maker{}
var Maker3Proj4 = maker{proj4: true}

func New(locator config.Locator) *Map {
	fm := nik.New()
	fm.SetSRS("epsg:3857")
	return &Map{
		inner:       fm,
		XML:         &XMLMap{SRS: "epsg:3857"},
		fontSets:    make(map[string]string),
		locator:     locator,
		scaleFactor: 1.0,
		zoomScales:  webmercZoomScales,
	}
}

func (m *Map) SetAutoTypeFilter(enable bool) {
	m.autoTypeFilter = enable
}

func (m *Map) SetBackgroundColor(c color.Color) {
	r, g, b := c.ToRgb()
	nr := uint8(r*255 + 0.5)
	ng := uint8(g*255 + 0.5)
	nb := uint8(b*255 + 0.5)
	na := uint8(c.A*255 + 0.5)
	m.inner.SetBackgroundColor(imagecolor.NRGBA{R: nr, G: ng, B: nb, A: na})

	cs := fmt.Sprintf("#%02x%02x%02x", nr, ng, nb)
	m.XML.BgColor = &cs
}

func (m *Map) SetZoomScales(zoomScales []int) {
	m.zoomScales = zoomScales
}

func (m *Map) SetProj4(enable bool) {
	if enable {
		m.XML.SRS = "+init=epsg:3857"
		m.proj4 = true
	} else {
		m.XML.SRS = "epsg:3857"
	}
	if m.inner != nil {
		m.inner.SetSRS(m.XML.SRS)
	}
}

func (m *Map) AddLayer(l cartocss.Layer, rules []cartocss.Rule) {
	if l.ScaleFactor != 0.0 {
		prevScaleFactor := m.scaleFactor
		defer func() { m.scaleFactor = prevScaleFactor }()
		m.scaleFactor = l.ScaleFactor
	}

	styles := m.newStyles(rules)

	var styleNames []string
	for _, s := range styles {
		fs := nik.NewStyle()
		fs.SetFilterMode(1)

		if s.CompOp != nil {
			fs.SetCompOp(compOpToInt(*s.CompOp))
		}
		if s.Opacity != nil {
			fs.SetOpacity(float32(*s.Opacity))
		}

		for _, sr := range s.Rules {
			if sr.fsRule != nil {
				fs.AddRule(sr.fsRule)
			}
		}

		if err := m.inner.InsertStyle(s.Name, fs); err != nil {
			panic(fmt.Sprintf("insert style %s: %v", s.Name, err))
		}
		styleNames = append(styleNames, s.Name)
	}

	srs := l.SRS
	if m.proj4 && strings.HasPrefix(strings.ToLower(srs), "epsg:") {
		srs = "+init=" + srs
	}
	layer := nik.NewLayer(l.ID, srs)
	if !l.Active {
		layer.SetSrs(srs)
	}
	if l.GroupBy != "" {
		layer.SetName(l.ID)
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

	m.xmlAddLayer(l, rules, &styles)
}

func (m *Map) newStyles(rules []cartocss.Rule) []Style {
	styles := []Style{}
	style := Style{FilterMode: "first"}

	for _, r := range rules {
		mr := m.newRule(r)

		styleName := r.Layer
		if r.Attachment != "" {
			styleName += "-" + r.Attachment
		}

		if style.Name != styleName {
			if len(style.Rules) > 0 {
				styles = append(styles, style)
			}
			style = Style{Name: styleName, FilterMode: "first"}
			for _, rr := range rules {
				if r.Attachment == rr.Attachment {
					if v, ok := r.Properties.GetString("comp-op"); ok {
						style.CompOp = &v
					}
					if v, ok := r.Properties.GetFloat("opacity"); ok {
						style.Opacity = &v
					}
				}
			}
		}
		style.Rules = append(style.Rules, *mr)
	}
	if len(style.Rules) > 0 {
		styles = append(styles, style)
	}
	return styles
}

func (m *Map) newRule(r cartocss.Rule) *Rule {
	result := &Rule{}

	if r.Zoom != cartocss.AllZoom {
		result.Zoom = r.Zoom.String()
	}
	var minScale, maxScale float64
	if l := r.Zoom.First(); l > 0 {
		if l > len(m.zoomScales) {
			l = len(m.zoomScales)
		}
		result.MaxScaleDenom = m.zoomScales[l-1]
		maxScale = float64(result.MaxScaleDenom)
	}
	if l := r.Zoom.Last(); l < len(m.zoomScales) {
		result.MinScaleDenom = m.zoomScales[l]
		minScale = float64(result.MinScaleDenom)
	}

	result.Filter = fmtFilters(r.Filters)
	result.fsRule = nik.NewRule(r.Layer, minScale, maxScale)
	if result.Filter != "" {
		result.fsRule.SetFilter(result.Filter)
	}

	prefixes := cartocss.SortedPrefixes(r.Properties, []string{"line-", "polygon-", "polygon-pattern-", "text-", "shield-", "marker-", "point-", "building-", "raster-"})

	for _, p := range prefixes {
		r.Properties.SetDefaultInstance(p.Instance)
		switch p.Name {
		case "line-":
			m.addLineSymbolizer(result, r)
		case "line-pattern-":
			m.addLinePatternSymbolizer(result, r)
		case "polygon-":
			m.addPolygonSymbolizer(result, r)
		case "polygon-pattern-":
			m.addPolygonPatternSymbolizer(result, r)
		case "text-":
			m.addTextSymbolizer(result, r)
		case "shield-":
			m.addShieldSymbolizer(result, r)
		case "marker-":
			m.addMarkerSymbolizer(result, r)
		case "point-":
			m.addPointSymbolizer(result, r)
		case "building-":
			m.addBuildingSymbolizer(result, r)
		case "dot-":
			m.addDotSymbolizer(result, r)
		case "raster-":
			m.addRasterSymbolizer(result, r)
		}
	}
	r.Properties.SetDefaultInstance("")
	return result
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

func (m *Map) xmlAddLayer(l cartocss.Layer, rules []cartocss.Rule, styles *[]Style) {
	xmlStyles := *styles
	m.XML.Styles = append(m.XML.Styles, xmlStyles...)

	layer := Layer{}
	layer.SRS = &l.SRS
	if m.proj4 && strings.HasPrefix(strings.ToLower(*layer.SRS), "epsg:") {
		srs := "+init=" + *layer.SRS
		layer.SRS = &srs
	}
	layer.Name = l.ID
	if !l.Active {
		layer.Status = "off"
	}
	if l.GroupBy != "" {
		layer.GroupBy = l.GroupBy
	}
	if l.ClearLabelCache {
		layer.ClearLabelCache = "on"
	}
	if l.CacheFeatures {
		layer.CacheFeatures = "true"
	}

	z := cartocss.RulesZoom(rules)
	if z != cartocss.AllZoom {
		if first := z.First(); first > 0 {
			if first > len(m.zoomScales) {
				first = len(m.zoomScales)
			}
			layer.MaxScaleDenom = m.zoomScales[first-1]
		}
		if last := z.Last(); last < len(m.zoomScales) {
			layer.MinScaleDenom = m.zoomScales[last]
		}
	}
	params := m.newXMLDatasource(l.Datasource, rules)
	if params != nil {
		layer.Datasource = &params
	}
	for _, s := range xmlStyles {
		layer.StyleNames = append(layer.StyleNames, s.Name)
	}
	m.XML.Layers = append(m.XML.Layers, layer)
}

func (m *Map) newXMLDatasource(ds cartocss.Datasource, rules []cartocss.Rule) []Parameter {
	var params []Parameter
	switch ds := ds.(type) {
	case *cartocss.PostGIS:
		params = []Parameter{
			{Name: "host", Value: ds.Host},
			{Name: "port", Value: ds.Port},
			{Name: "geometry_field", Value: ds.GeometryField},
			{Name: "dbname", Value: ds.Database},
			{Name: "user", Value: ds.Username},
			{Name: "password", Value: ds.Password},
			{Name: "extent", Value: ds.Extent},
			{Name: "table", Value: pqSelectString(ds.Query, rules, m.autoTypeFilter)},
			{Name: "srid", Value: ds.SRID},
			{Name: "type", Value: "postgis"},
		}
		if ds.SimplifyGeometries == "true" {
			params = append(params, Parameter{Name: "simplify_geometries", Value: "true"})
		}
	case *cartocss.Shapefile:
		fname := m.locator.Shape(ds.Filename)
		params = []Parameter{
			{Name: "file", Value: fname},
			{Name: "type", Value: "shape"},
		}
	case *cartocss.SQLite:
		fname := m.locator.SQLite(ds.Filename)
		params = []Parameter{
			{Name: "file", Value: fname},
			{Name: "srid", Value: ds.SRID},
			{Name: "extent", Value: ds.Extent},
			{Name: "geometry_field", Value: ds.GeometryField},
			{Name: "table", Value: ds.Query},
			{Name: "type", Value: "sqlite"},
		}
	case *cartocss.OGR:
		file := ds.Filename
		if !isOgrConnection.MatchString(ds.Filename) {
			file = m.locator.Data(ds.Filename)
		}
		params = []Parameter{
			{Name: "file", Value: file},
			{Name: "srid", Value: ds.SRID},
			{Name: "extent", Value: ds.Extent},
			{Name: "layer", Value: ds.Layer},
			{Name: "layer_by_sql", Value: ds.Query},
			{Name: "type", Value: "ogr"},
		}
	case *cartocss.GDAL:
		params = []Parameter{
			{Name: "file", Value: m.locator.Data(ds.Filename)},
			{Name: "srid", Value: ds.SRID},
			{Name: "extent", Value: ds.Extent},
			{Name: "band", Value: ds.Band},
			{Name: "type", Value: "gdal"},
		}
	case *cartocss.GeoJson:
		fname := m.locator.Shape(ds.Filename)
		params = []Parameter{
			{Name: "file", Value: fname},
			{Name: "type", Value: "geojson"},
		}
	}

	var result []Parameter
	for _, p := range params {
		if p.Value != "" {
			result = append(result, p)
		}
	}
	return result
}
