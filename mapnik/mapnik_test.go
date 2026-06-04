package mapnik

import (
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
}

func TestAddLayerSimpleRule(t *testing.T) {
	mss := `#layer { line-width: 2; line-color: #ff0; }`
	d := newTestDecoder(t, mss)
	layer := cartocss.Layer{ID: "layer", Type: cartocss.LineString, Active: true}
	rules := d.MSS().LayerZoomRules("layer", cartocss.AllZoom)
	assert.NotEmpty(t, rules)

	m := New(newTestLocator())
	assert.NotPanics(t, func() { m.AddLayer(layer, rules) })
}

func TestAddLayerPolygon(t *testing.T) {
	mss := `#layer { polygon-fill: #3366cc; polygon-opacity: 0.8; }`
	d := newTestDecoder(t, mss)
	layer := cartocss.Layer{ID: "layer", Type: cartocss.Polygon, Active: true}

	m := New(newTestLocator())
	assert.NotPanics(t, func() {
		m.AddLayer(layer, d.MSS().LayerZoomRules("layer", cartocss.AllZoom))
	})
}

func TestAddLayerText(t *testing.T) {
	mss := `#layer { text-name: "[name]"; text-size: 12; text-fill: #333; text-face-name: "DejaVu Sans Book"; }`
	d := newTestDecoder(t, mss)
	layer := cartocss.Layer{ID: "layer", Type: cartocss.LineString, Active: true}

	m := New(newTestLocator())
	assert.NotPanics(t, func() {
		m.AddLayer(layer, d.MSS().LayerZoomRules("layer", cartocss.AllZoom))
	})
}

func TestAddLayerShield(t *testing.T) {
	mss := `#layer[type='motorway'] { shield-name: "[ref]"; shield-file: url(img/shield.svg); shield-size: 10; shield-fill: #fff; shield-placement: line; }`
	d := newTestDecoder(t, mss)
	layer := cartocss.Layer{ID: "layer", Type: cartocss.LineString, Active: true}

	m := New(newTestLocator())
	assert.NotPanics(t, func() {
		m.AddLayer(layer, d.MSS().LayerZoomRules("layer", cartocss.AllZoom))
	})
}

func TestAddLayerRaster(t *testing.T) {
	mss := `#layer { raster-opacity: 0.8; raster-scaling: lanczos; raster-comp-op: color-dodge; }`
	d := newTestDecoder(t, mss)
	layer := cartocss.Layer{ID: "layer", Type: cartocss.Raster, Active: true}

	m := New(newTestLocator())
	assert.NotPanics(t, func() {
		m.AddLayer(layer, d.MSS().LayerZoomRules("layer", cartocss.AllZoom))
	})
}

func TestAddLayerMarker(t *testing.T) {
	mss := `#layer { marker-file: url(img/marker.svg); marker-width: 20; marker-height: 20; marker-fill: #f00; marker-allow-overlap: true; }`
	d := newTestDecoder(t, mss)
	layer := cartocss.Layer{ID: "layer", Type: cartocss.Point, Active: true}

	m := New(newTestLocator())
	assert.NotPanics(t, func() {
		m.AddLayer(layer, d.MSS().LayerZoomRules("layer", cartocss.AllZoom))
	})
}

func TestAddLayerInactive(t *testing.T) {
	mss := `#layer { line-width: 2; line-color: #ff0; }`
	d := newTestDecoder(t, mss)
	layer := cartocss.Layer{ID: "layer", Type: cartocss.LineString, Active: false}

	m := New(newTestLocator())
	assert.NotPanics(t, func() {
		m.AddLayer(layer, d.MSS().LayerZoomRules("layer", cartocss.AllZoom))
	})
}

func TestMultipleLayers(t *testing.T) {
	mss := `#roads { line-width: 2; line-color: #333; } #buildings { polygon-fill: #ccc; }`
	d := newTestDecoder(t, mss)

	m := New(newTestLocator())
	assert.NotPanics(t, func() {
		m.AddLayer(cartocss.Layer{ID: "roads", Type: cartocss.LineString, Active: true},
			d.MSS().LayerZoomRules("roads", cartocss.AllZoom))
		m.AddLayer(cartocss.Layer{ID: "buildings", Type: cartocss.Polygon, Active: true},
			d.MSS().LayerZoomRules("buildings", cartocss.AllZoom))
	})
}

func TestWriteXML(t *testing.T) {
	mss := `#layer { line-width: 2; line-color: #ff0; }`
	d := newTestDecoder(t, mss)
	layer := cartocss.Layer{ID: "layer", Type: cartocss.LineString, Active: true}

	m := New(newTestLocator())
	m.AddLayer(layer, d.MSS().LayerZoomRules("layer", cartocss.AllZoom))

	var buf strings.Builder
	err := m.Write(&buf)
	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "<?xml")
	assert.Contains(t, output, "Map")
	assert.Contains(t, output, "layer")
}

func TestWriteFiles(t *testing.T) {
	mss := `#layer { line-width: 2; line-color: #ff0; }`
	d := newTestDecoder(t, mss)
	layer := cartocss.Layer{ID: "layer", Type: cartocss.LineString, Active: true}

	m := New(newTestLocator())
	m.AddLayer(layer, d.MSS().LayerZoomRules("layer", cartocss.AllZoom))

	basename := filepath.Join(t.TempDir(), "test.xml")
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

func TestCompOpToInt(t *testing.T) {
	assert.Equal(t, 0, compOpToInt("clear"))
	assert.Equal(t, 3, compOpToInt("src-over"))
	assert.Equal(t, 15, compOpToInt("screen"))
	assert.Equal(t, 32, compOpToInt("color"))
	assert.Equal(t, 33, compOpToInt("value"))
	assert.Equal(t, 3, compOpToInt("unknown"))
}

func BenchmarkAddLayer(b *testing.B) {
	mss := `
		#water { polygon-fill: #aaddff; polygon-opacity: 0.8; }
		#roads { line-width: 2; line-color: #333; line-opacity: 0.9; }
		#labels { text-name: "[name]"; text-size: 10; text-fill: #222; text-face-name: "DejaVu Sans"; }
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
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := New(newTestLocator())
		for _, l := range layers {
			m.AddLayer(l, d.MSS().LayerZoomRules(l.ID, cartocss.AllZoom))
		}
	}
}
