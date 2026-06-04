package mapnik

import (
	"fmt"
	"strings"

	cartocss "github.com/flywave/go-cartocss"

	nik "github.com/flywave/flywave-mapnik"
	"github.com/flywave/go-cartocss/color"
)

func colorToRGBA(c color.Color) (uint8, uint8, uint8, uint8) {
	r, g, b := c.ToRgb()
	return uint8(r*255 + 0.5), uint8(g*255 + 0.5), uint8(b*255 + 0.5), uint8(c.A*255 + 0.5)
}

func setSymbolizerColor(sym *nik.Symbolizer, key nik.SymbolizerKey, c color.Color) {
	cr, cg, cb, ca := colorToRGBA(c)
	sym.SetColor(key, cr, cg, cb, ca)
}

func setSymbolizerColorByKey(sym *nik.Symbolizer, key string, c color.Color) {
	cr, cg, cb, ca := colorToRGBA(c)
	sym.SetColorByKey(key, cr, cg, cb, ca)
}

func setSymbolizerDoubleByKey(sym *nik.Symbolizer, key string, v float64) {
	sym.SetDoubleByKey(key, v)
}

func setSymbolizerStringByKey(sym *nik.Symbolizer, key, v string) {
	sym.SetStringByKey(key, v)
}

func setSymbolizerBoolByKey(sym *nik.Symbolizer, key string, v bool) {
	sym.SetBoolByKey(key, v)
}

func (m *Map) addLineSymbolizer(result *Rule, r cartocss.Rule) {
	if width, ok := r.Properties.GetFloat("line-width"); ok && width != 0.0 {
		sym := &Symbolizer{}

		ls := &LineSymbolizer{}
		ls.Width = fmtFloat(width*m.scaleFactor, true)
		ls.Clip = fmtBool(r.Properties.GetBool("line-clip"))
		ls.Color = fmtColor(r.Properties.GetColor("line-color"))
		if v, ok := r.Properties.GetFloatList("line-dasharray"); ok {
			ls.Dasharray = fmtPattern(v, m.scaleFactor, true)
		}
		if v, ok := r.Properties.GetFloatList("line-dash-offset"); ok {
			ls.DashOffset = fmtPattern(v, m.scaleFactor, true)
		}
		ls.Gamma = fmtFloat(r.Properties.GetFloat("line-gamma"))
		ls.GammaMethod = fmtString(r.Properties.GetString("line-gamma-method"))
		ls.Linecap = fmtString(r.Properties.GetString("line-cap"))
		ls.Miterlimit = fmtFloatProp(r.Properties, "line-miterlimit", m.scaleFactor)
		ls.Linejoin = fmtString(r.Properties.GetString("line-join"))
		ls.Offset = fmtFloatProp(r.Properties, "line-offset", m.scaleFactor)
		ls.Opacity = fmtFloat(r.Properties.GetFloat("line-opacity"))
		ls.Rasterizer = fmtString(r.Properties.GetString("line-rasterizer"))
		ls.Simplify = fmtFloat(r.Properties.GetFloat("line-simplify"))
		ls.SimplifyAlgorithm = fmtString(r.Properties.GetString("line-simplify-algorithm"))
		ls.Smooth = fmtFloat(r.Properties.GetFloat("line-smooth"))
		ls.CompOp = fmtString(r.Properties.GetString("line-comp-op"))
		ls.GeometryTransform = fmtString(r.Properties.GetString("line-geometry-transform"))
		sym.LineSymbolizer = ls
		result.Symbolizers = append(result.Symbolizers, sym)

		// native API
		fs := nik.NewSymbolizer(nik.SymbolizerLine)
		fs.SetDouble(nik.KeyWidth, width*m.scaleFactor)
		if c, ok := r.Properties.GetColor("line-color"); ok {
			cr, cg, cb, ca := colorToRGBA(c)
			fs.SetColor(nik.KeyStroke, cr, cg, cb, ca)
		}
		if v, ok := r.Properties.GetFloat("line-opacity"); ok {
			fs.SetDouble(nik.KeyStrokeOpacity, v)
		}
		if v, ok := r.Properties.GetBool("line-clip"); ok {
			fs.SetBool(nik.KeyClip, v)
		}
		if v, ok := r.Properties.GetFloat("line-gamma"); ok {
			fs.SetDouble(nik.KeyStrokeGamma, v)
		}
		if v, ok := r.Properties.GetString("line-gamma-method"); ok {
			fs.SetString(nik.KeyStrokeGammaMethod, v)
		}
		if v, ok := r.Properties.GetString("line-cap"); ok {
			fs.SetString(nik.KeyStrokeLinecap, v)
		}
		if v, ok := r.Properties.GetString("line-join"); ok {
			fs.SetString(nik.KeyStrokeLinejoin, v)
		}
		if v, ok := r.Properties.GetFloat("line-miterlimit"); ok {
			fs.SetDouble(nik.KeyStrokeMiterlimit, v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("line-offset"); ok {
			fs.SetDouble(nik.KeyOffset, v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("line-smooth"); ok {
			fs.SetDouble(nik.KeySmooth, v)
		}
		if v, ok := r.Properties.GetString("line-rasterizer"); ok {
			fs.SetString(nik.KeyLineRasterizer, v)
		}
		if v, ok := r.Properties.GetString("line-comp-op"); ok {
			fs.SetString(nik.KeyCompOp, v)
		}
		if v, ok := r.Properties.GetFloatList("line-dasharray"); ok {
			fs.SetString(nik.KeyStrokeDasharray, *fmtPatternStr(v, m.scaleFactor))
		}
		if v, ok := r.Properties.GetFloatList("line-dash-offset"); ok {
			fs.SetString(nik.KeyStrokeDashoffset, *fmtPatternStr(v, m.scaleFactor))
		}
		if v, ok := r.Properties.GetFloat("line-simplify"); ok {
			fs.SetDouble(nik.KeySimplifyTolerance, v)
		}
		if v, ok := r.Properties.GetString("line-simplify-algorithm"); ok {
			fs.SetString(nik.KeySimplifyAlgorithm, v)
		}
		if v, ok := r.Properties.GetString("line-geometry-transform"); ok {
			fs.SetString(nik.KeyGeometryTransform, v)
		}
		m.appendSymbolizer(result.fsRule, fs)
	}
}

func (m *Map) addLinePatternSymbolizer(result *Rule, r cartocss.Rule) {
	if patFile, ok := r.Properties.GetString("line-pattern-file"); ok {
		sym := &Symbolizer{}
		lps := &LinePatternSymbolizer{}
		fname := m.locator.Image(patFile)
		lps.File = &fname
		lps.Offset = fmtFloatProp(r.Properties, "line-pattern-offset", m.scaleFactor)
		lps.Clip = fmtBool(r.Properties.GetBool("line-pattern-clip"))
		lps.Simplify = fmtFloat(r.Properties.GetFloat("line-pattern-simplify"))
		lps.SimplifyAlgorithm = fmtString(r.Properties.GetString("line-pattern-simplify-algorithm"))
		lps.Smooth = fmtFloat(r.Properties.GetFloat("line-pattern-smooth"))
		lps.GeometryTransform = fmtString(r.Properties.GetString("line-pattern-geometry-transform"))
		lps.CompOp = fmtString(r.Properties.GetString("line-pattern-comp-op"))
		lps.Opacity = fmtFloat(r.Properties.GetFloat("line-pattern-opacity"))
		sym.LinePatternSymbolizer = lps
		result.Symbolizers = append(result.Symbolizers, sym)

		fs := nik.NewSymbolizer(nik.SymbolizerLinePattern)
		fs.SetString(nik.KeyFile, fname)
		if v, ok := r.Properties.GetFloat("line-pattern-offset"); ok {
			fs.SetDouble(nik.KeyOffset, v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetBool("line-pattern-clip"); ok {
			fs.SetBool(nik.KeyClip, v)
		}
		if v, ok := r.Properties.GetFloat("line-pattern-smooth"); ok {
			fs.SetDouble(nik.KeySmooth, v)
		}
		if v, ok := r.Properties.GetString("line-pattern-comp-op"); ok {
			fs.SetString(nik.KeyCompOp, v)
		}
		if v, ok := r.Properties.GetFloat("line-pattern-opacity"); ok {
			fs.SetDouble(nik.KeyOpacity, v)
		}
		if v, ok := r.Properties.GetFloat("line-pattern-simplify"); ok {
			fs.SetDouble(nik.KeySimplifyTolerance, v)
		}
		if v, ok := r.Properties.GetString("line-pattern-simplify-algorithm"); ok {
			fs.SetString(nik.KeySimplifyAlgorithm, v)
		}
		if v, ok := r.Properties.GetString("line-pattern-geometry-transform"); ok {
			fs.SetString(nik.KeyGeometryTransform, v)
		}
		m.appendSymbolizer(result.fsRule, fs)
	}
}

func (m *Map) addPolygonSymbolizer(result *Rule, r cartocss.Rule) {
	if fill, ok := r.Properties.GetColor("polygon-fill"); ok {
		sym := &Symbolizer{}
		ps := &PolygonSymbolizer{}
		ps.Color = fmtColor(fill, true)
		ps.Opacity = fmtFloat(r.Properties.GetFloat("polygon-opacity"))
		ps.Gamma = fmtFloat(r.Properties.GetFloat("polygon-gamma"))
		ps.GammaMethod = fmtString(r.Properties.GetString("polygon-gamma-method"))
		ps.Clip = fmtBool(r.Properties.GetBool("polygon-clip"))
		ps.Simplify = fmtFloat(r.Properties.GetFloat("polygon-simplify"))
		ps.SimplifyAlgorithm = fmtString(r.Properties.GetString("polygon-simplify-algorithm"))
		ps.Smooth = fmtFloat(r.Properties.GetFloat("polygon-smooth"))
		ps.GeometryTransform = fmtString(r.Properties.GetString("polygon-geometry-transform"))
		ps.CompOp = fmtString(r.Properties.GetString("polygon-comp-op"))
		sym.PolygonSymbolizer = ps
		result.Symbolizers = append(result.Symbolizers, sym)

		fs := nik.NewSymbolizer(nik.SymbolizerPolygon)
		setSymbolizerColor(fs, nik.KeyFill, fill)
		if v, ok := r.Properties.GetFloat("polygon-opacity"); ok {
			fs.SetDouble(nik.KeyFillOpacity, v)
		}
		if v, ok := r.Properties.GetFloat("polygon-gamma"); ok {
			fs.SetDouble(nik.KeyGamma, v)
		}
		if v, ok := r.Properties.GetString("polygon-gamma-method"); ok {
			fs.SetString(nik.KeyGammaMethod, v)
		}
		if v, ok := r.Properties.GetBool("polygon-clip"); ok {
			fs.SetBool(nik.KeyClip, v)
		}
		if v, ok := r.Properties.GetFloat("polygon-smooth"); ok {
			fs.SetDouble(nik.KeySmooth, v)
		}
		if v, ok := r.Properties.GetString("polygon-comp-op"); ok {
			fs.SetString(nik.KeyCompOp, v)
		}
		if v, ok := r.Properties.GetFloat("polygon-simplify"); ok {
			fs.SetDouble(nik.KeySimplifyTolerance, v)
		}
		if v, ok := r.Properties.GetString("polygon-simplify-algorithm"); ok {
			fs.SetString(nik.KeySimplifyAlgorithm, v)
		}
		if v, ok := r.Properties.GetString("polygon-geometry-transform"); ok {
			fs.SetString(nik.KeyGeometryTransform, v)
		}
		m.appendSymbolizer(result.fsRule, fs)
	}
}

func (m *Map) addPolygonPatternSymbolizer(result *Rule, r cartocss.Rule) {
	if patFile, ok := r.Properties.GetString("polygon-pattern-file"); ok {
		sym := &Symbolizer{}
		pps := &PolygonPatternSymbolizer{}
		fname := m.locator.Image(patFile)
		pps.File = &fname
		pps.Alignment = fmtString(r.Properties.GetString("polygon-pattern-alignment"))
		pps.Gamma = fmtFloat(r.Properties.GetFloat("polygon-pattern-gamma"))
		pps.Opacity = fmtFloat(r.Properties.GetFloat("polygon-pattern-opacity"))
		pps.Clip = fmtBool(r.Properties.GetBool("polygon-pattern-clip"))
		pps.Simplify = fmtFloat(r.Properties.GetFloat("polygon-pattern-simplify"))
		pps.SimplifyAlgorithm = fmtString(r.Properties.GetString("polygon-pattern-simplify-algorithm"))
		pps.Smooth = fmtFloat(r.Properties.GetFloat("polygon-pattern-smooth"))
		pps.GeometryTransform = fmtString(r.Properties.GetString("polygon-pattern-geometry-transform"))
		pps.CompOp = fmtString(r.Properties.GetString("polygon-pattern-comp-op"))
		sym.PolygonPatternSymbolizer = pps
		result.Symbolizers = append(result.Symbolizers, sym)

		fs := nik.NewSymbolizer(nik.SymbolizerPolygonPattern)
		fs.SetString(nik.KeyFile, fname)
		if v, ok := r.Properties.GetString("polygon-pattern-alignment"); ok {
			fs.SetString(nik.KeyAlignment, v)
		}
		if v, ok := r.Properties.GetFloat("polygon-pattern-opacity"); ok {
			fs.SetDouble(nik.KeyOpacity, v)
		}
		if v, ok := r.Properties.GetBool("polygon-pattern-clip"); ok {
			fs.SetBool(nik.KeyClip, v)
		}
		if v, ok := r.Properties.GetFloat("polygon-pattern-smooth"); ok {
			fs.SetDouble(nik.KeySmooth, v)
		}
		if v, ok := r.Properties.GetString("polygon-pattern-comp-op"); ok {
			fs.SetString(nik.KeyCompOp, v)
		}
		if v, ok := r.Properties.GetFloat("polygon-pattern-gamma"); ok {
			fs.SetDouble(nik.KeyGamma, v)
		}
		if v, ok := r.Properties.GetFloat("polygon-pattern-simplify"); ok {
			fs.SetDouble(nik.KeySimplifyTolerance, v)
		}
		if v, ok := r.Properties.GetString("polygon-pattern-simplify-algorithm"); ok {
			fs.SetString(nik.KeySimplifyAlgorithm, v)
		}
		if v, ok := r.Properties.GetString("polygon-pattern-geometry-transform"); ok {
			fs.SetString(nik.KeyGeometryTransform, v)
		}
		m.appendSymbolizer(result.fsRule, fs)
	}
}

func (m *Map) addTextSymbolizer(result *Rule, r cartocss.Rule) {
	if size, ok := r.Properties.GetFloat("text-size"); ok {
		sym := &Symbolizer{}
		ts := &TextSymbolizer{}
		ts.Size = fmtFloat(size*m.scaleFactor, true)
		ts.Fill = fmtColor(r.Properties.GetColor("text-fill"))
		ts.Name = fmtField(r.Properties.GetFieldList("text-name"))
		ts.AvoidEdges = fmtBool(r.Properties.GetBool("text-avoid-edges"))
		ts.HaloFill = fmtColor(r.Properties.GetColor("text-halo-fill"))
		ts.HaloRadius = fmtFloatProp(r.Properties, "text-halo-radius", m.scaleFactor)
		ts.HaloRasterizer = fmtString(r.Properties.GetString("text-halo-rasterizer"))
		ts.Opacity = fmtFloat(r.Properties.GetFloat("text-opacity"))
		ts.WrapCharacter = fmtString(r.Properties.GetString("text-wrap-character"))
		ts.WrapBefore = fmtString(r.Properties.GetString("text-wrap-before"))
		ts.WrapWidth = fmtFloatProp(r.Properties, "text-wrap-width", m.scaleFactor)
		ts.Ratio = fmtFloat(r.Properties.GetFloat("text-ratio"))
		ts.MaxCharAngleDelta = fmtFloat(r.Properties.GetFloat("text-max-char-angle-delta"))
		ts.Placement = fmtString(r.Properties.GetString("text-placement"))
		ts.PlacementType = fmtString(r.Properties.GetString("text-placement-type"))
		ts.Placements = fmtString(r.Properties.GetString("text-placements"))
		ts.LabelPositionTolerance = fmtFloatProp(r.Properties, "text-label-position-tolerance", m.scaleFactor)
		ts.VerticalAlign = fmtString(r.Properties.GetString("text-vertical-alignment"))
		ts.HorizontalAlign = fmtString(r.Properties.GetString("text-horizontal-alignment"))
		ts.JustifyAlign = fmtString(r.Properties.GetString("text-justify-alignment"))
		ts.CompOp = fmtString(r.Properties.GetString("text-comp-op"))
		ts.Dx = fmtFloatProp(r.Properties, "text-dx", m.scaleFactor)
		ts.Dy = fmtFloatProp(r.Properties, "text-dy", m.scaleFactor)
		if v, ok := r.Properties.GetFloat("text-orientation"); ok {
			ts.Orientation = fmtFloat(v, true)
		} else if v, ok := r.Properties.GetFieldList("text-orientation"); ok {
			ts.Orientation = fmtField(v, true)
		}
		ts.CharacterSpacing = fmtFloatProp(r.Properties, "text-character-spacing", m.scaleFactor)
		ts.LineSpacing = fmtFloatProp(r.Properties, "text-line-spacing", m.scaleFactor)
		ts.AllowOverlap = fmtBool(r.Properties.GetBool("text-allow-overlap"))
		ts.Spacing = fmtFloatProp(r.Properties, "text-spacing", m.scaleFactor)
		ts.MinimumDistance = fmtFloatProp(r.Properties, "text-min-distance", m.scaleFactor)
		ts.MinimumPadding = fmtFloatProp(r.Properties, "text-min-padding", m.scaleFactor)
		ts.MinPathLength = fmtFloatProp(r.Properties, "text-min-path-length", m.scaleFactor)
		ts.Clip = fmtBool(r.Properties.GetBool("text-clip"))
		ts.TextTransform = fmtString(r.Properties.GetString("text-transform"))
		if faceNames, ok := r.Properties.GetStringList("text-face-name"); ok {
			ts.FontsetName = m.fontSetName(faceNames)
		}
		ts.HaloOpacity = fmtFloat(r.Properties.GetFloat("text-halo-opacity"))
		ts.HaloTransform = fmtString(r.Properties.GetString("text-halo-transform"))
		ts.HaloCompOp = fmtString(r.Properties.GetString("text-halo-comp-op"))
		ts.RepeatWrapCharacter = fmtBool(r.Properties.GetBool("text-repeat-wrap-characater"))
		ts.Margin = fmtFloatProp(r.Properties, "text-margin", m.scaleFactor)
		ts.Simplify = fmtFloat(r.Properties.GetFloat("text-simplify"))
		ts.SimplifyAlgorithm = fmtString(r.Properties.GetString("text-simplify-algorithm"))
		ts.Smooth = fmtFloat(r.Properties.GetFloat("text-smooth"))
		ts.RotateDisplacement = fmtBool(r.Properties.GetBool("text-rotate-displacement"))
		ts.Upright = fmtString(r.Properties.GetString("text-upgright"))
		ts.FontFeatureSettings = fmtString(r.Properties.GetString("font-feature-settings"))
		ts.LargestBboxOnly = fmtBool(r.Properties.GetBool("text-largest-bbox-only"))
		ts.RepeatDistance = fmtFloatProp(r.Properties, "text-repeat-distance", m.scaleFactor)
		sym.TextSymbolizer = ts
		if ts.Name != nil && *ts.Name != "" {
			result.Symbolizers = append(result.Symbolizers, sym)
		}

		fs := nik.NewSymbolizer(nik.SymbolizerText)

		if v, ok := r.Properties.GetString("text-name"); ok {
			fs.TextSetName(v)
		} else if v, ok := r.Properties.GetFieldList("text-name"); ok {
			fs.TextSetName(*fmtField(v, true))
		}

		fs.SetDouble(nik.KeyWidth, size*m.scaleFactor)

		if v, ok := r.Properties.GetColor("text-fill"); ok {
			cr, cg, cb, ca := colorToRGBA(v)
			fs.TextSetColor("fill", cr, cg, cb, ca)
		}
		if v, ok := r.Properties.GetColor("text-halo-fill"); ok {
			cr, cg, cb, ca := colorToRGBA(v)
			fs.TextSetColor("halo-fill", cr, cg, cb, ca)
		}
		if v, ok := r.Properties.GetFloat("text-opacity"); ok {
			fs.TextSetDouble("text-opacity", v)
		}
		if v, ok := r.Properties.GetBool("text-avoid-edges"); ok {
			fs.SetBool(nik.KeyAvoidEdges, v)
		}

		if v, ok := r.Properties.GetFloat("text-halo-radius"); ok {
			fs.TextSetDouble("halo-radius", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("text-halo-opacity"); ok {
			fs.TextSetDouble("halo-opacity", v)
		}
		if v, ok := r.Properties.GetString("text-halo-rasterizer"); ok {
			fs.SetString(nik.KeyHaloRasterizer, v)
		}
		if v, ok := r.Properties.GetString("text-halo-transform"); ok {
			fs.SetString(nik.KeyHaloTransform, v)
		}
		if v, ok := r.Properties.GetString("text-halo-comp-op"); ok {
			fs.SetString(nik.KeyHaloCompOp, v)
		}

		if v, ok := r.Properties.GetString("text-placement"); ok {
			fs.SetString(nik.KeyLabelPlacement, v)
		}
		if v, ok := r.Properties.GetString("text-placement-type"); ok {
			fs.SetString(nik.KeyPointPlacement, v)
		}
		if v, ok := r.Properties.GetString("text-placements"); ok {
			fs.SetString(nik.KeyTextPlacements, v)
		}
		if v, ok := r.Properties.GetString("text-transform"); ok {
			fs.TextSetString("text-transform", v)
		}
		if v, ok := r.Properties.GetString("text-horizontal-alignment"); ok {
			fs.SetString(nik.KeyHorizontalAlignment, v)
		}
		if v, ok := r.Properties.GetString("text-vertical-alignment"); ok {
			fs.SetString(nik.KeyVerticalAlignment, v)
		}
		if v, ok := r.Properties.GetString("text-justify-alignment"); ok {
			fs.SetString(nik.KeyJustifyAlignment, v)
		}
		if v, ok := r.Properties.GetString("text-comp-op"); ok {
			fs.SetString(nik.KeyCompOp, v)
		}

		if v, ok := r.Properties.GetBool("text-allow-overlap"); ok {
			fs.SetBool(nik.KeyAllowOverlap, v)
		}
		if v, ok := r.Properties.GetFloat("text-spacing"); ok {
			fs.TextSetDouble("spacing", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("text-min-distance"); ok {
			fs.TextSetDouble("minimum-distance", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("text-min-padding"); ok {
			fs.TextSetDouble("minimum-padding", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetBool("text-clip"); ok {
			fs.SetBool(nik.KeyClip, v)
		}
		if v, ok := r.Properties.GetFloat("text-smooth"); ok {
			fs.SetDouble(nik.KeySmooth, v)
		}
		if v, ok := r.Properties.GetString("text-upright"); ok {
			fs.SetString(nik.KeyUpright, v)
		}
		if v, ok := r.Properties.GetString("font-feature-settings"); ok {
			fs.SetString(nik.KeyFFSettings, v)
		}
		if v, ok := r.Properties.GetBool("text-largest-bbox-only"); ok {
			fs.SetBool(nik.KeyLargestBoxOnly, v)
		}
		if v, ok := r.Properties.GetFloat("text-min-path-length"); ok {
			fs.SetDouble(nik.KeyMinimumPathLength, v)
		}
		if v, ok := r.Properties.GetFloat("text-simplify"); ok {
			fs.SetDouble(nik.KeySimplifyTolerance, v)
		}
		if v, ok := r.Properties.GetString("text-simplify-algorithm"); ok {
			fs.SetString(nik.KeySimplifyAlgorithm, v)
		}

		if v, ok := r.Properties.GetString("text-face-name"); ok {
			fs.TextSetFaceName(v)
		} else if v, ok := r.Properties.GetStringList("text-face-name"); ok {
			fs.TextSetFaceName(strings.Join(v, ","))
		}
		if v, ok := r.Properties.GetFloat("text-dx"); ok {
			fs.TextSetDouble("dx", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("text-dy"); ok {
			fs.TextSetDouble("dy", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("text-wrap-width"); ok {
			fs.TextSetDouble("wrap-width", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetString("text-wrap-character"); ok {
			fs.TextSetString("wrap-character", v)
		}
		if v, ok := r.Properties.GetFloat("text-character-spacing"); ok {
			fs.TextSetDouble("character-spacing", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("text-line-spacing"); ok {
			fs.TextSetDouble("line-spacing", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("text-ratio"); ok {
			fs.TextSetDouble("text-ratio", v)
		}
		if v, ok := r.Properties.GetFloat("text-margin"); ok {
			fs.TextSetDouble("margin", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("text-repeat-distance"); ok {
			fs.TextSetDouble("repeat-distance", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("text-label-position-tolerance"); ok {
			fs.TextSetDouble("label-position-tolerance", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("text-max-char-angle-delta"); ok {
			fs.TextSetDouble("max-char-angle-delta", v)
		}
		if v, ok := r.Properties.GetFloat("text-orientation"); ok {
			fs.TextSetDouble("orientation", v)
		} else if v, ok := r.Properties.GetFieldList("text-orientation"); ok {
			fs.TextSetString("orientation", *fmtField(v, true))
		}
		m.appendSymbolizer(result.fsRule, fs)
	}
}

func (m *Map) addShieldSymbolizer(result *Rule, r cartocss.Rule) {
	if shieldFile, ok := r.Properties.GetString("shield-file"); ok {
		sym := &Symbolizer{}
		ss := &ShieldSymbolizer{}
		fname := m.locator.Image(shieldFile)
		ss.File = &fname
		ss.Size = fmtFloatProp(r.Properties, "shield-size", m.scaleFactor)
		ss.Fill = fmtColor(r.Properties.GetColor("shield-fill"))
		ss.Name = fmtField(r.Properties.GetFieldList("shield-name"))
		ss.TextOpacity = fmtFloat(r.Properties.GetFloat("shield-text-opacity"))
		ss.Opacity = fmtFloat(r.Properties.GetFloat("shield-opacity"))
		ss.Transform = fmtString(r.Properties.GetString("shield-transform"))
		ss.CompOp = fmtString(r.Properties.GetString("shield-comp-op"))
		ss.Placement = fmtString(r.Properties.GetString("shield-placement"))
		ss.PlacementType = fmtString(r.Properties.GetString("shield-placement-type"))
		ss.Placements = fmtString(r.Properties.GetString("shield-placements"))
		ss.UnlockImage = fmtBool(r.Properties.GetBool("shield-unlock-image"))
		ss.HorizontalAlign = fmtString(r.Properties.GetString("shield-horizontal-alignment"))
		ss.VerticalAlign = fmtString(r.Properties.GetString("shield-vertical-alignment"))
		ss.JustifyAlign = fmtString(r.Properties.GetString("shield-justify-alignment"))
		ss.Clip = fmtBool(r.Properties.GetBool("shield-clip"))
		ss.AllowOverlap = fmtBool(r.Properties.GetBool("shield-allow-overlap"))
		ss.AvoidEdges = fmtBool(r.Properties.GetBool("shield-avoid-edges"))
		ss.HaloFill = fmtColor(r.Properties.GetColor("shield-halo-fill"))
		ss.HaloRadius = fmtFloatProp(r.Properties, "shield-halo-radius", m.scaleFactor)
		ss.HaloRasterizer = fmtString(r.Properties.GetString("shield-halo-rasterizer"))
		ss.CharacterSpacing = fmtFloatProp(r.Properties, "shield-character-spacing", m.scaleFactor)
		ss.WrapCharacter = fmtString(r.Properties.GetString("shield-wrap-character"))
		ss.WrapBefore = fmtBool(r.Properties.GetBool("shield-wrap-before"))
		ss.WrapWidth = fmtFloatProp(r.Properties, "shield-wrap-width", m.scaleFactor)
		ss.LineSpacing = fmtFloatProp(r.Properties, "shield-line-spacing", m.scaleFactor)
		ss.HaloTransform = fmtString(r.Properties.GetString("shield-halo-transform"))
		ss.HaloCompOp = fmtString(r.Properties.GetString("shield-halo-comp-op"))
		ss.HaloOpacity = fmtFloat(r.Properties.GetFloat("shield-halo-opacity"))
		ss.Dx = fmtFloatProp(r.Properties, "shield-dx", m.scaleFactor)
		ss.Dy = fmtFloatProp(r.Properties, "shield-dy", m.scaleFactor)
		ss.TextDx = fmtFloatProp(r.Properties, "shield-text-dx", m.scaleFactor)
		ss.TextDy = fmtFloatProp(r.Properties, "shield-text-dy", m.scaleFactor)
		ss.TextTransform = fmtString(r.Properties.GetString("shield-text-transform"))
		ss.Spacing = fmtFloatProp(r.Properties, "shield-spacing", m.scaleFactor)
		ss.MinimumDistance = fmtFloatProp(r.Properties, "shield-min-distance", m.scaleFactor)
		ss.MinimumPadding = fmtFloatProp(r.Properties, "shield-min-padding", m.scaleFactor)
		ss.LabelPositionTolerance = fmtFloatProp(r.Properties, "shield-label-position-tolerance", m.scaleFactor)
		ss.Margin = fmtFloatProp(r.Properties, "shield-margin", m.scaleFactor)
		ss.RepeatDistance = fmtFloatProp(r.Properties, "shield-repeat-distance", m.scaleFactor)
		ss.Simplify = fmtFloat(r.Properties.GetFloat("shield-simplify"))
		ss.SimplifyAlgorithm = fmtString(r.Properties.GetString("shield-simplify-algorithm"))
		ss.Smooth = fmtFloat(r.Properties.GetFloat("shield-smooth"))
		if faceNames, ok := r.Properties.GetStringList("shield-face-name"); ok {
			ss.FontsetName = m.fontSetName(faceNames)
		}
		sym.ShieldSymbolizer = ss
		result.Symbolizers = append(result.Symbolizers, sym)

		fs := nik.NewSymbolizer(nik.SymbolizerShield)

		fs.SetString(nik.KeyFile, fname)

		if v, ok := r.Properties.GetString("shield-name"); ok {
			fs.TextSetName(v)
		} else if v, ok := r.Properties.GetFieldList("shield-name"); ok {
			fs.TextSetName(*fmtField(v, true))
		}

		if v, ok := r.Properties.GetFloat("shield-size"); ok {
			fs.SetDouble(nik.KeyWidth, v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetColor("shield-fill"); ok {
			setSymbolizerColor(fs, nik.KeyFill, v)
		}
		if v, ok := r.Properties.GetColor("shield-halo-fill"); ok {
			cr, cg, cb, ca := colorToRGBA(v)
			fs.TextSetColor("halo-fill", cr, cg, cb, ca)
		}
		if v, ok := r.Properties.GetFloat("shield-opacity"); ok {
			fs.SetDouble(nik.KeyOpacity, v)
		}
		if v, ok := r.Properties.GetString("shield-comp-op"); ok {
			fs.SetString(nik.KeyCompOp, v)
		}
		if v, ok := r.Properties.GetString("shield-placement"); ok {
			fs.SetString(nik.KeyLabelPlacement, v)
		}
		if v, ok := r.Properties.GetBool("shield-unlock-image"); ok {
			fs.SetBool(nik.KeyUnlockImage, v)
		}
		if v, ok := r.Properties.GetBool("shield-allow-overlap"); ok {
			fs.SetBool(nik.KeyAllowOverlap, v)
		}
		if v, ok := r.Properties.GetBool("shield-avoid-edges"); ok {
			fs.SetBool(nik.KeyAvoidEdges, v)
		}
		if v, ok := r.Properties.GetBool("shield-clip"); ok {
			fs.SetBool(nik.KeyClip, v)
		}

		if v, ok := r.Properties.GetString("shield-halo-rasterizer"); ok {
			fs.SetString(nik.KeyHaloRasterizer, v)
		}
		if v, ok := r.Properties.GetString("shield-halo-transform"); ok {
			fs.SetString(nik.KeyHaloTransform, v)
		}
		if v, ok := r.Properties.GetString("shield-halo-comp-op"); ok {
			fs.SetString(nik.KeyHaloCompOp, v)
		}

		if v, ok := r.Properties.GetFloat("shield-dx"); ok {
			fs.SetDouble(nik.KeyShieldDx, v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("shield-dy"); ok {
			fs.SetDouble(nik.KeyShieldDy, v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("shield-text-dx"); ok {
			fs.TextSetDouble("dx", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("shield-text-dy"); ok {
			fs.TextSetDouble("dy", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetString("shield-text-transform"); ok {
			fs.SetString(nik.KeyTextTransform, v)
		}
		if v, ok := r.Properties.GetFloat("shield-smooth"); ok {
			fs.SetDouble(nik.KeySmooth, v)
		}
		if v, ok := r.Properties.GetFloat("shield-simplify"); ok {
			fs.SetDouble(nik.KeySimplifyTolerance, v)
		}
		if v, ok := r.Properties.GetString("shield-simplify-algorithm"); ok {
			fs.SetString(nik.KeySimplifyAlgorithm, v)
		}
		if v, ok := r.Properties.GetString("shield-horizontal-alignment"); ok {
			fs.SetString(nik.KeyHorizontalAlignment, v)
		}
		if v, ok := r.Properties.GetString("shield-vertical-alignment"); ok {
			fs.SetString(nik.KeyVerticalAlignment, v)
		}
		if v, ok := r.Properties.GetString("shield-justify-alignment"); ok {
			fs.SetString(nik.KeyJustifyAlignment, v)
		}

		if v, ok := r.Properties.GetString("shield-face-name"); ok {
			fs.TextSetFaceName(v)
		} else if v, ok := r.Properties.GetStringList("shield-face-name"); ok {
			fs.TextSetFaceName(strings.Join(v, ","))
		}
		if v, ok := r.Properties.GetFloat("shield-halo-radius"); ok {
			fs.TextSetDouble("halo-radius", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("shield-halo-opacity"); ok {
			fs.TextSetDouble("halo-opacity", v)
		}
		if v, ok := r.Properties.GetFloat("shield-spacing"); ok {
			fs.TextSetDouble("spacing", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("shield-min-distance"); ok {
			fs.TextSetDouble("minimum-distance", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("shield-min-padding"); ok {
			fs.TextSetDouble("minimum-padding", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("shield-wrap-width"); ok {
			fs.TextSetDouble("wrap-width", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetString("shield-wrap-character"); ok {
			fs.TextSetString("wrap-character", v)
		}
		if v, ok := r.Properties.GetFloat("shield-character-spacing"); ok {
			fs.TextSetDouble("character-spacing", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("shield-line-spacing"); ok {
			fs.TextSetDouble("line-spacing", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("shield-margin"); ok {
			fs.TextSetDouble("margin", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("shield-repeat-distance"); ok {
			fs.TextSetDouble("repeat-distance", v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("shield-label-position-tolerance"); ok {
			fs.TextSetDouble("label-position-tolerance", v*m.scaleFactor)
		}
		m.appendSymbolizer(result.fsRule, fs)
	}
}

func (m *Map) addMarkerSymbolizer(result *Rule, r cartocss.Rule) {
	sym := &Symbolizer{}
	ms := &MarkersSymbolizer{}
	ms.Width = fmtFloatProp(r.Properties, "marker-width", m.scaleFactor)
	ms.Height = fmtFloatProp(r.Properties, "marker-height", m.scaleFactor)
	ms.Fill = fmtColor(r.Properties.GetColor("marker-fill"))
	ms.FillOpacity = fmtFloat(r.Properties.GetFloat("marker-fill-opacity"))
	ms.Opacity = fmtFloat(r.Properties.GetFloat("marker-opacity"))
	ms.Placement = fmtString(r.Properties.GetString("marker-placement"))
	ms.Transform = fmtString(r.Properties.GetString("marker-transform"))
	ms.GeometryTransform = fmtString(r.Properties.GetString("marker-geometry-transform"))
	ms.Spacing = fmtFloatProp(r.Properties, "marker-spacing", m.scaleFactor)
	ms.Stroke = fmtColor(r.Properties.GetColor("marker-line-color"))
	ms.StrokeOpacity = fmtFloat(r.Properties.GetFloat("marker-line-opacity"))
	ms.StrokeWidth = fmtFloatProp(r.Properties, "marker-line-width", m.scaleFactor)
	ms.AllowOverlap = fmtBool(r.Properties.GetBool("marker-allow-overlap"))
	ms.MultiPolicy = fmtString(r.Properties.GetString("marker-multi-policy"))
	ms.IgnorePlacement = fmtBool(r.Properties.GetBool("marker-ignore-placement"))
	ms.MaxError = fmtFloatProp(r.Properties, "marker-max-error", m.scaleFactor)
	ms.Clip = fmtBool(r.Properties.GetBool("marker-clip"))
	ms.Smooth = fmtFloat(r.Properties.GetFloat("marker-smooth"))
	ms.CompOp = fmtString(r.Properties.GetString("marker-comp-op"))
	ms.AvoidEdges = fmtBool(r.Properties.GetBool("marker-avoid-edges"))
	ms.Simplify = fmtFloat(r.Properties.GetFloat("marker-simplify"))
	ms.SimplifyAlgorithm = fmtString(r.Properties.GetString("marker-simplify-algorithm"))
	ms.Offset = fmtFloatProp(r.Properties, "marker-offset", m.scaleFactor)
	ms.Direction = fmtString(r.Properties.GetString("marker-direction"))
	if markerFile, ok := r.Properties.GetString("marker-file"); ok {
		fname := m.locator.Image(markerFile)
		ms.File = &fname
	} else {
		markerType, ok := r.Properties.GetString("marker-type")
		if !ok {
			markerType = "ellipse"
			if ms.Fill == nil && ms.Stroke == nil && ms.StrokeWidth == nil {
				return
			}
		}
		ms.MarkerType = &markerType
	}
	sym.MarkersSymbolizer = ms
	result.Symbolizers = append(result.Symbolizers, sym)

	fs := nik.NewSymbolizer(nik.SymbolizerMarkers)
	if v, ok := r.Properties.GetFloat("marker-width"); ok {
		fs.SetDouble(nik.KeyWidth, v*m.scaleFactor)
	}
	if v, ok := r.Properties.GetFloat("marker-height"); ok {
		fs.SetDouble(nik.KeyHeight, v*m.scaleFactor)
	}
	if c, ok := r.Properties.GetColor("marker-fill"); ok {
		setSymbolizerColor(fs, nik.KeyFill, c)
	}
	if v, ok := r.Properties.GetFloat("marker-fill-opacity"); ok {
		fs.SetDouble(nik.KeyFillOpacity, v)
	}
	if v, ok := r.Properties.GetFloat("marker-opacity"); ok {
		fs.SetDouble(nik.KeyOpacity, v)
	}
	if v, ok := r.Properties.GetString("marker-placement"); ok {
		fs.SetString(nik.KeyMarkersPlacement, v)
	}
	if v, ok := r.Properties.GetFloat("marker-spacing"); ok {
		fs.SetDouble(nik.KeySpacing, v*m.scaleFactor)
	}
	if v, ok := r.Properties.GetString("marker-geometry-transform"); ok {
		fs.SetString(nik.KeyGeometryTransform, v)
	}
	if c, ok := r.Properties.GetColor("marker-line-color"); ok {
		setSymbolizerColor(fs, nik.KeyStroke, c)
	}
	if v, ok := r.Properties.GetFloat("marker-line-opacity"); ok {
		fs.SetDouble(nik.KeyStrokeOpacity, v)
	}
	if v, ok := r.Properties.GetFloat("marker-line-width"); ok {
		fs.SetDouble(nik.KeyStrokeWidth, v*m.scaleFactor)
	}
	if v, ok := r.Properties.GetBool("marker-allow-overlap"); ok {
		fs.SetBool(nik.KeyAllowOverlap, v)
	}
	if v, ok := r.Properties.GetString("marker-multi-policy"); ok {
		fs.SetString(nik.KeyMarkersMultipolicy, v)
	}
	if v, ok := r.Properties.GetBool("marker-ignore-placement"); ok {
		fs.SetBool(nik.KeyIgnorePlacement, v)
	}
	if v, ok := r.Properties.GetFloat("marker-max-error"); ok {
		fs.SetDouble(nik.KeyMaxError, v*m.scaleFactor)
	}
	if v, ok := r.Properties.GetBool("marker-clip"); ok {
		fs.SetBool(nik.KeyClip, v)
	}
	if v, ok := r.Properties.GetFloat("marker-smooth"); ok {
		fs.SetDouble(nik.KeySmooth, v)
	}
	if v, ok := r.Properties.GetString("marker-comp-op"); ok {
		fs.SetString(nik.KeyCompOp, v)
	}
	if v, ok := r.Properties.GetBool("marker-avoid-edges"); ok {
		fs.SetBool(nik.KeyAvoidEdges, v)
	}
	if v, ok := r.Properties.GetFloat("marker-offset"); ok {
		fs.SetDouble(nik.KeyOffset, v*m.scaleFactor)
	}
	if v, ok := r.Properties.GetString("marker-direction"); ok {
		fs.SetString(nik.KeyDirection, v)
	}
	if markerFile, ok := r.Properties.GetString("marker-file"); ok {
		fs.SetString(nik.KeyFile, m.locator.Image(markerFile))
	}
	m.appendSymbolizer(result.fsRule, fs)
}

func (m *Map) addPointSymbolizer(result *Rule, r cartocss.Rule) {
	if pointFile, ok := r.Properties.GetString("point-file"); ok {
		sym := &Symbolizer{}
		ps := &PointSymbolizer{}
		fname := m.locator.Image(pointFile)
		ps.File = &fname
		ps.AllowOverlap = fmtBool(r.Properties.GetBool("point-allow-overlap"))
		ps.Opacity = fmtFloat(r.Properties.GetFloat("point-opacity"))
		ps.Transform = fmtString(r.Properties.GetString("point-transform"))
		ps.IgnorePlacement = fmtBool(r.Properties.GetBool("point-ignore-placement"))
		ps.Placement = fmtString(r.Properties.GetString("point-placement"))
		ps.CompOp = fmtString(r.Properties.GetString("point-comp-op"))
		sym.PointSymbolizer = ps
		result.Symbolizers = append(result.Symbolizers, sym)

		fs := nik.NewSymbolizer(nik.SymbolizerPoint)
		fs.SetString(nik.KeyFile, fname)
		if v, ok := r.Properties.GetBool("point-allow-overlap"); ok {
			fs.SetBool(nik.KeyAllowOverlap, v)
		}
		if v, ok := r.Properties.GetFloat("point-opacity"); ok {
			fs.SetDouble(nik.KeyOpacity, v)
		}
		if v, ok := r.Properties.GetString("point-transform"); ok {
			fs.SetString(nik.KeyImageTransform, v)
		}
		if v, ok := r.Properties.GetBool("point-ignore-placement"); ok {
			fs.SetBool(nik.KeyIgnorePlacement, v)
		}
		if v, ok := r.Properties.GetString("point-placement"); ok {
			fs.SetString(nik.KeyPointPlacement, v)
		}
		if v, ok := r.Properties.GetString("point-comp-op"); ok {
			fs.SetString(nik.KeyCompOp, v)
		}
		m.appendSymbolizer(result.fsRule, fs)
	}
}

func (m *Map) addBuildingSymbolizer(result *Rule, r cartocss.Rule) {
	if fill, ok := r.Properties.GetColor("building-fill"); ok {
		sym := &Symbolizer{}
		bs := &BuildingSymbolizer{}
		bs.Fill = fmtColor(fill, true)
		bs.FillOpacity = fmtFloat(r.Properties.GetFloat("building-fill-opacity"))
		bs.Height = fmtFloatProp(r.Properties, "building-height", m.scaleFactor)
		sym.BuildingSymbolizer = bs
		result.Symbolizers = append(result.Symbolizers, sym)

		fs := nik.NewSymbolizer(nik.SymbolizerBuilding)
		setSymbolizerColor(fs, nik.KeyFill, fill)
		if v, ok := r.Properties.GetFloat("building-fill-opacity"); ok {
			fs.SetDouble(nik.KeyFillOpacity, v)
		}
		if v, ok := r.Properties.GetFloat("building-height"); ok {
			fs.SetDouble(nik.KeyHeight, v*m.scaleFactor)
		}
		m.appendSymbolizer(result.fsRule, fs)
	}
}

func (m *Map) addDotSymbolizer(result *Rule, r cartocss.Rule) {
	if fill, ok := r.Properties.GetColor("dot-fill"); ok {
		sym := &Symbolizer{}
		ds := &DotSymbolizer{}
		ds.Fill = fmtColor(fill, true)
		ds.Opacity = fmtFloat(r.Properties.GetFloat("dot-opacity"))
		ds.Width = fmtFloatProp(r.Properties, "dot-width", m.scaleFactor)
		ds.Height = fmtFloatProp(r.Properties, "dot-height", m.scaleFactor)
		ds.CompOp = fmtString(r.Properties.GetString("dot-comp-op"))
		sym.DotSymbolizer = ds
		result.Symbolizers = append(result.Symbolizers, sym)

		fs := nik.NewSymbolizer(nik.SymbolizerDot)
		setSymbolizerColor(fs, nik.KeyFill, fill)
		if v, ok := r.Properties.GetFloat("dot-opacity"); ok {
			fs.SetDouble(nik.KeyOpacity, v)
		}
		if v, ok := r.Properties.GetFloat("dot-width"); ok {
			fs.SetDouble(nik.KeyWidth, v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetFloat("dot-height"); ok {
			fs.SetDouble(nik.KeyHeight, v*m.scaleFactor)
		}
		if v, ok := r.Properties.GetString("dot-comp-op"); ok {
			fs.SetString(nik.KeyCompOp, v)
		}
		m.appendSymbolizer(result.fsRule, fs)
	}
}

func (m *Map) addRasterSymbolizer(result *Rule, r cartocss.Rule) {
	opacity, ok := r.Properties.GetFloat("raster-opacity")
	if ok && opacity == 0.0 {
		return
	}
	sym := &Symbolizer{}
	rs := &RasterSymbolizer{}
	if stops, ok := r.Properties.GetStopList("raster-colorizer-stops"); ok {
		for _, stop := range stops {
			rs.Stops = append(rs.Stops,
				Stop{
					Value: fmt.Sprintf("%d", stop.Value),
					Color: *fmtColor(stop.Color, true),
				},
			)
		}
	}
	rs.Opacity = fmtFloat(r.Properties.GetFloat("raster-opacity"))
	rs.Epsilon = fmtFloat(r.Properties.GetFloat("raster-colorizer-epsilon"))
	rs.MeshSize = fmtFloat(r.Properties.GetFloat("raster-mesh-size"))
	rs.FilterFactor = fmtFloat(r.Properties.GetFloat("raster-filter-factor"))
	rs.CompOp = fmtString(r.Properties.GetString("raster-comp-op"))
	rs.Scaling = fmtString(r.Properties.GetString("raster-scaling"))
	rs.DefaultMode = fmtString(r.Properties.GetString("raster-colorizer-default-mode"))
	rs.DefaultColor = fmtColor(r.Properties.GetColor("raster-colorizer-default-color"))
	sym.RasterSymbolizer = rs
	result.Symbolizers = append(result.Symbolizers, sym)

	fs := nik.NewSymbolizer(nik.SymbolizerRaster)
	if v, ok := r.Properties.GetFloat("raster-opacity"); ok {
		fs.SetDouble(nik.KeyOpacity, v)
	}
	if v, ok := r.Properties.GetFloat("raster-mesh-size"); ok {
		fs.SetDouble(nik.KeyMeshSize, v)
	}
	if v, ok := r.Properties.GetFloat("raster-filter-factor"); ok {
		fs.SetDouble(nik.KeyFilterFactor, v)
	}
	if v, ok := r.Properties.GetString("raster-comp-op"); ok {
		fs.SetString(nik.KeyCompOp, v)
	}
	if v, ok := r.Properties.GetString("raster-scaling"); ok {
		fs.SetString(nik.KeyScaling, v)
	}

	if v, ok := r.Properties.GetString("raster-colorizer-default-mode"); ok {
		fs.RasterColorizerSetMode(v)
	}
	if c, ok := r.Properties.GetColor("raster-colorizer-default-color"); ok {
		cr, cg, cb, ca := colorToRGBA(c)
		fs.RasterColorizerSetDefaultColor(cr, cg, cb, ca)
	}
	if v, ok := r.Properties.GetFloat("raster-colorizer-epsilon"); ok {
		fs.RasterColorizerSetEpsilon(v)
	}
	if stops, ok := r.Properties.GetStopList("raster-colorizer-stops"); ok {
		for _, stop := range stops {
			cr, cg, cb, ca := colorToRGBA(stop.Color)
			fs.RasterColorizerAddStop(float64(stop.Value), cr, cg, cb, ca)
		}
	}
	m.appendSymbolizer(result.fsRule, fs)
}

func (m *Map) appendSymbolizer(rule *nik.Rule, sym *nik.Symbolizer) {
	if rule != nil && sym != nil {
		rule.Append(sym)
	}
}

func (m *Map) fontSetName(fontFaces []string) *string {
	if len(fontFaces) == 0 {
		return nil
	}
	var str string
	for i, f := range fontFaces {
		if i > 0 {
			str += ", "
		}
		str += f
	}
	if fontSetName, ok := m.fontSets[str]; ok {
		return &fontSetName
	}
	fontSet := FontSet{}
	fontSet.Name = fmt.Sprintf("fontset-%d", len(m.fontSets)+1)
	m.fontSets[str] = fontSet.Name
	for _, f := range fontFaces {
		fontSet.Fonts = append(fontSet.Fonts, Font{FaceName: f})
	}
	m.XML.FontSets = append(m.XML.FontSets, fontSet)
	return &fontSet.Name
}
