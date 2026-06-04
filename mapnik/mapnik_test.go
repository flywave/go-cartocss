package mapnik

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	cartocss "github.com/flywave/go-cartocss"
	"github.com/flywave/go-cartocss/color"
	"github.com/flywave/go-cartocss/config"
	"github.com/stretchr/testify/assert"
)

func newTestDecoder(t *testing.T, mss string) *cartocss.Decoder {
	t.Helper()
	d := cartocss.NewDecoder()
	err := d.ParseString(mss)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.Evaluate(); err != nil {
		t.Fatal(err)
	}
	return d
}

func newTestLocator() config.Locator {
	l := &config.LookupLocator{}
	l.SetBaseDir(".")
	l.AddImageDir(".")
	l.AddShapeDir(".")
	l.AddDataDir(".")
	return l
}

func TestNew(t *testing.T) {
	m := New(newTestLocator())
	assert.NotNil(t, m)
	assert.NotNil(t, m.inner)
	assert.NotNil(t, m.XML)
	assert.Equal(t, "epsg:3857", m.XML.SRS)
}

func TestSetBackgroundColor(t *testing.T) {
	m := New(newTestLocator())
	m.SetBackgroundColor(color.MustParse("#ff0000"))
	assert.Equal(t, "#ff0000", *m.XML.BgColor)
}

func TestSetProj4(t *testing.T) {
	m := New(newTestLocator())
	m.SetProj4(true)
	assert.True(t, strings.HasPrefix(m.XML.SRS, "+init="))
}

func TestAddLayerSimpleRule(t *testing.T) {
	mss := `
		#layer {
			line-width: 2;
			line-color: #ff0;
		}
	`
	d := newTestDecoder(t, mss)
	layer := cartocss.Layer{ID: "layer", Type: cartocss.LineString, Active: true}
	rules := d.MSS().LayerZoomRules("layer", cartocss.AllZoom)
	assert.NotEmpty(t, rules)

	m := New(newTestLocator())
	m.AddLayer(layer, rules)

	var buf bytes.Buffer
	err := m.Write(&buf)
	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "layer")
	assert.Contains(t, output, "LineSymbolizer")
	assert.Contains(t, output, `"#ffff00"`)
}

func TestAddLayerPolygon(t *testing.T) {
	mss := `
		#layer {
			polygon-fill: #3366cc;
			polygon-opacity: 0.8;
		}
	`
	d := newTestDecoder(t, mss)
	layer := cartocss.Layer{ID: "layer", Type: cartocss.Polygon, Active: true}
	rules := d.MSS().LayerZoomRules("layer", cartocss.AllZoom)

	m := New(newTestLocator())
	m.AddLayer(layer, rules)

	var buf bytes.Buffer
	m.Write(&buf)
	output := buf.String()
	assert.Contains(t, output, "PolygonSymbolizer")
	assert.Contains(t, output, "#3366cc")
	assert.Contains(t, output, "0.8")
}

func TestAddLayerText(t *testing.T) {
	mss := `
		#layer {
			text-name: "[name]";
			text-size: 12;
			text-fill: #333;
			text-face-name: "DejaVu Sans Book";
		}
	`
	d := newTestDecoder(t, mss)
	layer := cartocss.Layer{ID: "layer", Type: cartocss.LineString, Active: true}
	rules := d.MSS().LayerZoomRules("layer", cartocss.AllZoom)

	m := New(newTestLocator())
	m.AddLayer(layer, rules)

	var buf bytes.Buffer
	m.Write(&buf)
	output := buf.String()
	assert.Contains(t, output, "TextSymbolizer")
	assert.Contains(t, output, "[name]")
	assert.Contains(t, output, `"DejaVu Sans Book"`)
}

func TestAddLayerShield(t *testing.T) {
	mss := `
		#layer[type='motorway'] {
			shield-name: "[ref]";
			shield-file: url(img/shield.svg);
			shield-size: 10;
			shield-fill: #fff;
			shield-placement: line;
		}
	`
	d := newTestDecoder(t, mss)
	layer := cartocss.Layer{ID: "layer", Type: cartocss.LineString, Active: true}
	rules := d.MSS().LayerZoomRules("layer", cartocss.AllZoom)

	m := New(newTestLocator())
	m.AddLayer(layer, rules)

	var buf bytes.Buffer
	m.Write(&buf)
	output := buf.String()
	assert.Contains(t, output, "ShieldSymbolizer")
	assert.Contains(t, output, "shield.svg")
}

func TestAddLayerRaster(t *testing.T) {
	mss := `
		#layer {
			raster-opacity: 0.8;
			raster-scaling: lanczos;
			raster-comp-op: color-dodge;
		}
	`
	d := newTestDecoder(t, mss)
	layer := cartocss.Layer{ID: "layer", Type: cartocss.Raster, Active: true}
	rules := d.MSS().LayerZoomRules("layer", cartocss.AllZoom)

	m := New(newTestLocator())
	m.AddLayer(layer, rules)

	var buf bytes.Buffer
	m.Write(&buf)
	output := buf.String()
	assert.Contains(t, output, "RasterSymbolizer")
	assert.Contains(t, output, "lanczos")
	assert.Contains(t, output, "0.8")
}

func TestAddLayerMarker(t *testing.T) {
	mss := `
		#layer {
			marker-file: url(img/marker.svg);
			marker-width: 20;
			marker-height: 20;
			marker-fill: #f00;
			marker-allow-overlap: true;
		}
	`
	d := newTestDecoder(t, mss)
	layer := cartocss.Layer{ID: "layer", Type: cartocss.Point, Active: true}
	rules := d.MSS().LayerZoomRules("layer", cartocss.AllZoom)

	m := New(newTestLocator())
	m.AddLayer(layer, rules)

	var buf bytes.Buffer
	m.Write(&buf)
	output := buf.String()
	assert.Contains(t, output, "MarkersSymbolizer")
	assert.Contains(t, output, "marker.svg")
}

func TestAddLayerWithZoom(t *testing.T) {
	mss := `
		#layer[zoom>=12] {
			line-width: 2;
			line-color: #ff0;
		}
	`
	d := newTestDecoder(t, mss)
	layer := cartocss.Layer{ID: "layer", Type: cartocss.LineString, Active: true}
	rules := d.MSS().LayerZoomRules("layer", cartocss.ZoomRange(cartocss.NewZoomRange(cartocss.GTE, 12)))

	m := New(newTestLocator())
	m.AddLayer(layer, rules)

	var buf bytes.Buffer
	m.Write(&buf)
	output := buf.String()
	assert.Contains(t, output, "maximum-scale-denominator")
}

func TestAddLayerWithFilter(t *testing.T) {
	mss := `
		#layer[type='motorway'] {
			line-width: 2;
			line-color: #ff0;
		}
	`
	d := newTestDecoder(t, mss)
	layer := cartocss.Layer{ID: "layer", Type: cartocss.LineString, Active: true, Properties: map[string]interface{}{"minzoom": 10}}
	rules := d.MSS().LayerZoomRules("layer", cartocss.NewZoomRange(cartocss.GTE, 10))

	m := New(newTestLocator())
	m.AddLayer(layer, rules)

	var buf bytes.Buffer
	m.Write(&buf)
	output := buf.String()
	assert.Contains(t, output, "[type]")
}

func TestAddLayerInactive(t *testing.T) {
	mss := `
		#layer {
			line-width: 2;
			line-color: #ff0;
		}
	`
	d := newTestDecoder(t, mss)
	layer := cartocss.Layer{ID: "layer", Type: cartocss.LineString, Active: false}
	rules := d.MSS().LayerZoomRules("layer", cartocss.AllZoom)

	m := New(newTestLocator())
	m.AddLayer(layer, rules)

	var buf bytes.Buffer
	m.Write(&buf)
	output := buf.String()
	assert.Contains(t, output, `status="off"`)
}

func TestMultipleLayers(t *testing.T) {
	mss := `
		#roads {
			line-width: 2;
			line-color: #333;
		}
		#buildings {
			polygon-fill: #ccc;
		}
	`
	d := newTestDecoder(t, mss)

	m := New(newTestLocator())
	m.AddLayer(cartocss.Layer{ID: "roads", Type: cartocss.LineString, Active: true},
		d.MSS().LayerZoomRules("roads", cartocss.AllZoom))
	m.AddLayer(cartocss.Layer{ID: "buildings", Type: cartocss.Polygon, Active: true},
		d.MSS().LayerZoomRules("buildings", cartocss.AllZoom))

	var buf bytes.Buffer
	m.Write(&buf)
	output := buf.String()
	assert.Contains(t, output, "roads")
	assert.Contains(t, output, "buildings")
}

func TestAddLayerWithAttachment(t *testing.T) {
	mss := `
		#layer::outline {
			line-width: 1;
			line-color: #000;
		}
		#layer::fill {
			polygon-fill: #fff;
		}
	`
	d := newTestDecoder(t, mss)
	layer := cartocss.Layer{ID: "layer", Type: cartocss.Polygon, Active: true}
	rules := d.MSS().LayerZoomRules("layer", cartocss.AllZoom)

	m := New(newTestLocator())
	m.AddLayer(layer, rules)

	var buf bytes.Buffer
	m.Write(&buf)
	output := buf.String()
	assert.Contains(t, output, "layer-outline")
	assert.Contains(t, output, "layer-fill")
}

func TestWriteFiles(t *testing.T) {
	dir := t.TempDir()
	mss := `
		#layer {
			line-width: 1;
			line-color: #000;
		}
	`
	d := newTestDecoder(t, mss)
	layer := cartocss.Layer{ID: "layer", Type: cartocss.LineString, Active: true}

	m := New(newTestLocator())
	m.AddLayer(layer, d.MSS().LayerZoomRules("layer", cartocss.AllZoom))

	basename := filepath.Join(dir, "test")
	err := m.WriteFiles(basename)
	assert.NoError(t, err)
	assert.FileExists(t, basename)
}

func TestMaker3(t *testing.T) {
	l := newTestLocator()
	mw := Maker3.New(l)
	assert.NotNil(t, mw)
	assert.Equal(t, "mapnik", Maker3.Type())
	assert.Equal(t, ".xml", Maker3.FileSuffix())
}

func TestMaker3Proj4(t *testing.T) {
	l := newTestLocator()
	mw := Maker3Proj4.New(l)
	assert.NotNil(t, mw)
	assert.Equal(t, "mapnik", Maker3Proj4.Type())
	assert.Equal(t, ".xml", Maker3Proj4.FileSuffix())
}

func TestParseFullFixture(t *testing.T) {
	mss := `
		Map {
			background-color: #f8f8f8;
		}
		#water {
			polygon-fill: #aaddff;
			polygon-opacity: 0.8;
		}
		#roads {
			line-width: 2;
			line-color: #333;
			line-opacity: 0.9;
		}
		#labels {
			text-name: "[name]";
			text-size: 10;
			text-fill: #222;
			text-face-name: "DejaVu Sans";
			text-halo-fill: #fff;
			text-halo-radius: 1;
			text-placement: point;
		}
	`
	d := newTestDecoder(t, mss)

	m := New(newTestLocator())
	m.AddLayer(cartocss.Layer{ID: "water", Type: cartocss.Polygon, Active: true},
		d.MSS().LayerZoomRules("water", cartocss.AllZoom))
	m.AddLayer(cartocss.Layer{ID: "roads", Type: cartocss.LineString, Active: true},
		d.MSS().LayerZoomRules("roads", cartocss.AllZoom))
	m.AddLayer(cartocss.Layer{ID: "labels", Type: cartocss.LineString, Active: true},
		d.MSS().LayerZoomRules("labels", cartocss.AllZoom))

	var buf bytes.Buffer
	err := m.Write(&buf)
	assert.NoError(t, err)
	output := buf.String()

	assert.Contains(t, output, "name=\"water\"")
	assert.Contains(t, output, "name=\"roads\"")
	assert.Contains(t, output, "name=\"labels\"")
	assert.Contains(t, output, "PolygonSymbolizer")
	assert.Contains(t, output, "LineSymbolizer")
	assert.Contains(t, output, "TextSymbolizer")
}

func TestFilterString(t *testing.T) {
	rules := []cartocss.Rule{
		{Filters: []cartocss.Filter{
			{Field: "type", CompOp: cartocss.EQ, Value: "motorway"},
		}},
	}
	result := FilterString(rules)
	assert.Equal(t, `("type" IN ('motorway'))`, result)
}

func TestWrapWhere(t *testing.T) {
	result := WrapWhere("roads", `"type" IN ('motorway')`)
	assert.Equal(t, "(SELECT * FROM roads WHERE \"type\" IN ('motorway')) as filtered", result)

	result2 := WrapWhere("roads", "")
	assert.Equal(t, "roads", result2)
}

func TestHelpers(t *testing.T) {
	c := color.MustParse("#ff0000")
	s := fmtColor(c, true)
	assert.NotNil(t, s)
	assert.Equal(t, "#ff0000", *s)

	s2 := fmtColor(color.Color{}, false)
	assert.Nil(t, s2)

	f := fmtFloat(3.14, true)
	assert.NotNil(t, f)
	assert.Equal(t, "3.14", *f)

	f2 := fmtFloat(0, false)
	assert.Nil(t, f2)

	str := fmtString("hello", true)
	assert.NotNil(t, str)
	assert.Equal(t, "hello", *str)

	b := fmtBool(true, true)
	assert.NotNil(t, b)
	assert.Equal(t, "true", *b)

	b2 := fmtBool(false, true)
	assert.Equal(t, "false", *b2)
}

func TestFmtFilters(t *testing.T) {
	filters := []cartocss.Filter{
		{Field: "type", CompOp: cartocss.EQ, Value: "motorway"},
	}
	result := fmtFilters(filters)
	assert.Equal(t, "([type] = 'motorway')", result)

	filters2 := []cartocss.Filter{
		{Field: "type", CompOp: cartocss.GTE, Value: float64(5)},
	}
	result2 := fmtFilters(filters2)
	assert.Equal(t, "([type] >= 5)", result2)
}

func TestFmtPattern(t *testing.T) {
	v := []float64{10, 20, 30}
	s := fmtPattern(v, 1.0, true)
	assert.NotNil(t, s)
	assert.Equal(t, "10, 20, 30", *s)

	s2 := fmtPattern(v, 2.0, true)
	assert.Equal(t, "20, 40, 60", *s2)
}

func TestFmtField(t *testing.T) {
	f := fmtField([]interface{}{cartocss.Field("[name]")}, true)
	assert.NotNil(t, f)
	assert.Equal(t, "[name]", *f)

	f2 := fmtField([]interface{}{"hello"}, true)
	assert.Equal(t, "'hello'", *f2)
}

func TestColorToRGBA(t *testing.T) {
	c := color.MustParse("#ff0000")
	r, g, b, a := colorToRGBA(c)
	assert.Equal(t, uint8(255), r)
	assert.Equal(t, uint8(0), g)
	assert.Equal(t, uint8(0), b)
	assert.Equal(t, uint8(255), a)

	c2 := color.MustParse("transparent")
	r2, g2, b2, a2 := colorToRGBA(c2)
	assert.Equal(t, uint8(0), r2)
	assert.Equal(t, uint8(0), g2)
	assert.Equal(t, uint8(0), b2)
	assert.Equal(t, uint8(0), a2)
}

func TestWebmercZoomScales(t *testing.T) {
	assert.Equal(t, 500000000, webmercZoomScales[0])
	assert.Equal(t, 100, webmercZoomScales[len(webmercZoomScales)-1])
	assert.Equal(t, 23, len(webmercZoomScales))
}

func BenchmarkAddLayer(b *testing.B) {
	mss := `
		Map { background-color: #f8f8f8; }
		#water { polygon-fill: #aaddff; polygon-opacity: 0.8; }
		#roads { line-width: 2; line-color: #333; line-opacity: 0.9; }
		#labels {
			text-name: "[name]"; text-size: 10; text-fill: #222;
			text-face-name: "DejaVu Sans"; text-placement: point;
		}
		#parks { polygon-fill: #cfc; polygon-opacity: 0.5; }
		#buildings { polygon-fill: #ccc; }
	`
	d := cartocss.NewDecoder()
	if err := d.ParseString(mss); err != nil {
		b.Fatal(err)
	}
	if err := d.Evaluate(); err != nil {
		b.Fatal(err)
	}

	layers := []cartocss.Layer{
		{ID: "water", Type: cartocss.Polygon, Active: true},
		{ID: "roads", Type: cartocss.LineString, Active: true},
		{ID: "labels", Type: cartocss.LineString, Active: true},
		{ID: "parks", Type: cartocss.Polygon, Active: true},
		{ID: "buildings", Type: cartocss.Polygon, Active: true},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := New(newTestLocator())
		for _, l := range layers {
			m.AddLayer(l, d.MSS().LayerZoomRules(l.ID, cartocss.AllZoom))
		}
	}
}

func TestCompOpToInt(t *testing.T) {
	assert.Equal(t, 0, compOpToInt("clear"))
	assert.Equal(t, 3, compOpToInt("src-over"))
	assert.Equal(t, 15, compOpToInt("screen"))
	assert.Equal(t, 32, compOpToInt("color"))
	assert.Equal(t, 33, compOpToInt("value"))
	assert.Equal(t, 3, compOpToInt("unknown"))
}
