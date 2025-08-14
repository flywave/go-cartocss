package builder

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	cartocss "github.com/flywave/go-cartocss"

	"github.com/flywave/go-cartocss/color"
	"github.com/flywave/go-cartocss/config"
)

// Builder builds map styles from MML and MSS files.
type Builder struct {
	dstMap          Map
	mss             []string
	mml             string
	locator         config.Locator
	dumpRules       io.Writer
	includeInactive bool
}

// New returns a Builder
func New(mw Map) *Builder {
	return &Builder{dstMap: mw, includeInactive: true}
}

// AddMSS adds another mss file to this builder.
func (b *Builder) AddMSS(mss string) {
	b.mss = append(b.mss, mss)
}

// SetMML sets/overwirtes the mml file of this builder.
func (b *Builder) SetMML(mml string) {
	b.mml = mml
}

// SetDumpRulesDest enables internal debuging output.
func (b *Builder) SetDumpRulesDest(w io.Writer) {
	b.dumpRules = w
}

// SetIncludeInactive set whether status=off layers should be included in output.
func (b *Builder) SetIncludeInactive(includeInactive bool) {
	b.includeInactive = includeInactive
}

// Build parses MML, MSS files, builds all rules and adds them to the Map.
func (b *Builder) Build() error {
	layerIDs := []string{}
	layers := []cartocss.Layer{}

	var mmlObj *cartocss.MML
	if b.mml != "" {
		r, err := os.Open(b.mml)
		if err != nil {
			return err
		}
		defer r.Close()
		mmlObj, err = cartocss.Parse(r)
		if err != nil {
			return err
		}
		if len(b.mss) == 0 {
			for _, s := range mmlObj.Stylesheets {
				b.mss = append(b.mss, filepath.Join(filepath.Dir(b.mml), s))
			}
		}

		for _, l := range mmlObj.Layers {
			layers = append(layers, l)
			layerIDs = append(layerIDs, l.ID)
		}
	}

	carto := cartocss.NewDecoder()

	for _, mss := range b.mss {
		err := carto.ParseFile(mss)
		if err != nil {
			return err
		}
	}

	if err := carto.Evaluate(); err != nil {
		return err
	}

	if m, ok := b.dstMap.(MapZoomScaleSetter); ok {
		if mmlObj != nil && mmlObj.Map.ZoomScales != nil {
			m.SetZoomScales(mmlObj.Map.ZoomScales)
		}
	}

	if b.mml == "" {
		layerIDs = carto.MSS().Layers()
		for _, layerID := range layerIDs {
			layers = append(layers,
				// XXX assume we only have LineStrings for -mss only export
				cartocss.Layer{ID: layerID, Type: cartocss.LineString},
			)
		}
	}

	for _, l := range layers {
		zoom := layerZoomRange(l)
		rules := carto.MSS().LayerZoomRules(l.ID, zoom, l.Classes...)

		if b.dumpRules != nil {
			for _, r := range rules {
				fmt.Fprintln(b.dumpRules, r.String())
			}
		}
		if len(rules) > 0 && (l.Active || b.includeInactive) {
			b.dstMap.AddLayer(l, rules)
		}
	}

	if m, ok := b.dstMap.(MapOptionsSetter); ok {
		if bgColor, ok := carto.MSS().Map().GetColor("background-color"); ok {
			m.SetBackgroundColor(bgColor)
		}
	}
	return nil
}

func layerZoomRange(l cartocss.Layer) cartocss.ZoomRange {
	zoom := cartocss.InvalidZoom
	minZoom, minOk := l.Properties["minzoom"].(int)
	maxZoom, maxOk := l.Properties["maxzoom"].(int)
	if minOk {
		zoom = cartocss.NewZoomRange(cartocss.GTE, int64(minZoom))
		if maxOk {
			zoom = zoom & cartocss.NewZoomRange(cartocss.LTE, int64(maxZoom))
		}
	} else if maxOk {
		zoom = cartocss.NewZoomRange(cartocss.LTE, int64(maxZoom))
	}
	return zoom
}

type MapOptionsSetter interface {
	SetBackgroundColor(color.Color)
}

type MapZoomScaleSetter interface {
	SetZoomScales([]int)
}

type Writer interface {
	Write(io.Writer) error
	WriteFiles(basename string) error
}

type Map interface {
	AddLayer(cartocss.Layer, []cartocss.Rule)
}

type MapWriter interface {
	Writer
	Map
}

// BuildMapFromString parses the style from a string and adds all
// cartocss.Layers to the map.
func BuildMapFromString(m Map, mml *cartocss.MML, style string) error {
	carto := cartocss.NewDecoder()

	err := carto.ParseString(style)
	if err != nil {
		return err
	}
	if err := carto.Evaluate(); err != nil {
		return err
	}

	if m, ok := m.(MapZoomScaleSetter); ok {
		if mml.Map.ZoomScales != nil {
			m.SetZoomScales(mml.Map.ZoomScales)
		}
	}

	for _, l := range mml.Layers {
		zoom := layerZoomRange(l)
		rules := carto.MSS().LayerZoomRules(l.ID, zoom, l.Classes...)

		if len(rules) > 0 {
			m.AddLayer(l, rules)
		}
	}

	if m, ok := m.(MapOptionsSetter); ok {
		if bgColor, ok := carto.MSS().Map().GetColor("background-color"); ok {
			m.SetBackgroundColor(bgColor)
		}
	}
	return nil
}
