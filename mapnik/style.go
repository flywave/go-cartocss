package mapnik

import (
	"strconv"
	"strings"

	cartocss "github.com/flywave/go-cartocss"
	"github.com/flywave/go-cartocss/color"

	nik "github.com/flywave/flywave-mapnik"
)

func setSymbolizerColor(sym *nik.Symbolizer, key nik.SymbolizerKey, c color.Color) {
	cr, cg, cb, ca := colorToRGBA(c)
	sym.SetColor(key, cr, cg, cb, ca)
}

func (m *Map) addLineSymbolizer(r cartocss.Rule) *nik.Symbolizer {
	if width, ok := r.Properties.GetFloat("line-width"); ok && width != 0.0 {
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
			fs.SetString(nik.KeyStrokeDasharray, formatDashArray(v, m.scaleFactor))
		}
		if v, ok := r.Properties.GetFloatList("line-dash-offset"); ok {
			fs.SetString(nik.KeyStrokeDashoffset, formatDashArray(v, m.scaleFactor))
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
		return fs
	}
	return nil
}

func (m *Map) addLinePatternSymbolizer(r cartocss.Rule) *nik.Symbolizer {
	if patFile, ok := r.Properties.GetString("line-pattern-file"); ok {
		fname := m.locator.Image(patFile)
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
		return fs
	}
	return nil
}

func (m *Map) addPolygonSymbolizer(r cartocss.Rule) *nik.Symbolizer {
	if fill, ok := r.Properties.GetColor("polygon-fill"); ok {
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
		return fs
	}
	return nil
}

func (m *Map) addPolygonPatternSymbolizer(r cartocss.Rule) *nik.Symbolizer {
	if patFile, ok := r.Properties.GetString("polygon-pattern-file"); ok {
		fname := m.locator.Image(patFile)
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
		return fs
	}
	return nil
}

func (m *Map) addTextSymbolizer(r cartocss.Rule) *nik.Symbolizer {
	if size, ok := r.Properties.GetFloat("text-size"); ok {
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
		return fs
	}
	return nil
}

func (m *Map) addShieldSymbolizer(r cartocss.Rule) *nik.Symbolizer {
	if shieldFile, ok := r.Properties.GetString("shield-file"); ok {
		fname := m.locator.Image(shieldFile)
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
		return fs
	}
	return nil
}

func (m *Map) addMarkerSymbolizer(r cartocss.Rule) *nik.Symbolizer {
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
	return fs
}

func (m *Map) addPointSymbolizer(r cartocss.Rule) *nik.Symbolizer {
	if pointFile, ok := r.Properties.GetString("point-file"); ok {
		fname := m.locator.Image(pointFile)
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
		return fs
	}
	return nil
}

func (m *Map) addBuildingSymbolizer(r cartocss.Rule) *nik.Symbolizer {
	if fill, ok := r.Properties.GetColor("building-fill"); ok {
		fs := nik.NewSymbolizer(nik.SymbolizerBuilding)
		setSymbolizerColor(fs, nik.KeyFill, fill)
		if v, ok := r.Properties.GetFloat("building-fill-opacity"); ok {
			fs.SetDouble(nik.KeyFillOpacity, v)
		}
		if v, ok := r.Properties.GetFloat("building-height"); ok {
			fs.SetDouble(nik.KeyHeight, v*m.scaleFactor)
		}
		return fs
	}
	return nil
}

func (m *Map) addDotSymbolizer(r cartocss.Rule) *nik.Symbolizer {
	if fill, ok := r.Properties.GetColor("dot-fill"); ok {
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
		return fs
	}
	return nil
}

func (m *Map) addRasterSymbolizer(r cartocss.Rule) *nik.Symbolizer {
	opacity, ok := r.Properties.GetFloat("raster-opacity")
	if ok && opacity == 0.0 {
		return nil
	}
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

	hasColorizer := false
	if v, ok := r.Properties.GetString("raster-colorizer-default-mode"); ok {
		fs.RasterColorizerSetMode(v)
		hasColorizer = true
	}
	if c, ok := r.Properties.GetColor("raster-colorizer-default-color"); ok {
		cr, cg, cb, ca := colorToRGBA(c)
		fs.RasterColorizerSetDefaultColor(cr, cg, cb, ca)
		hasColorizer = true
	}
	if v, ok := r.Properties.GetFloat("raster-colorizer-epsilon"); ok {
		fs.RasterColorizerSetEpsilon(v)
		hasColorizer = true
	}
	if stops, ok := r.Properties.GetStopList("raster-colorizer-stops"); ok {
		for _, stop := range stops {
			cr, cg, cb, ca := colorToRGBA(stop.Color)
			fs.RasterColorizerAddStop(float64(stop.Value), cr, cg, cb, ca)
		}
		hasColorizer = true
	}
	if hasColorizer {
		fs.RasterColorizerInit()
	}
	return fs
}

func formatDashArray(v []float64, scale float64) string {
	parts := make([]string, 0, len(v))
	for i := range v {
		parts = append(parts, strconv.FormatFloat(v[i]*scale, 'f', -1, 64))
	}
	return strings.Join(parts, ", ")
}
