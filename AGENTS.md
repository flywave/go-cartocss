# go-cartocss

CartoCSS parser and Mapnik XML serializer in Go. Module `github.com/flywave/go-cartocss`.

## Packages

| Path | Purpose |
|------|---------|
| `.` (root) | Scanner, parser (`.mss`), rule engine, property types, CartoCSS spec validation |
| `color/` | HSL-based color model, parse/manipulate CSS colors |
| `config/` | TOML config loader, resource locator with lookup dirs |
| `builder/` | Builds map styles from MML + MSS files, optional caching layer |
| `mapnik/` | Builds styles directly on `flywave-mapnik` Map (native C API), `builder.Map` + `MapWriter` impl |

## Commands

- **Test all**: `go test ./...`
- **Test single package**: `go test ./config/`
- **Test single function**: `go test -run TestDecodeFiles ./...`
- Benchmarks: `go test -bench=. ./...`
- No lint/typecheck config in repo; standard `go vet ./...` works.

## Test conventions

- All tests use `github.com/stretchr/testify/assert`.
- Root package tests load MSS fixtures from `tests/*.mss` (`TestDecodeFiles`).
- Tests call `NewDecoder()` → `ParseFile()`/`ParseString()` → `Evaluate()` → query `MSS()`.
- No integration tests; no external services needed.
- `config` and `color` packages have their own test files.

## Key API flow

```go
d := cartocss.NewDecoder()
d.ParseFile("style.mss")
d.Evaluate()
d.MSS().LayerZoomRules("layer-name", zoom, classes...)
```

The builder package wraps this for MML+MSS pipelines:
```go
b := builder.New(mapnikMap)
b.SetMML("project.mml")
b.AddMSS("style.mss")
b.Build()
```

## Important notes

- `Evaluate()` must be called after all `ParseFile`/`ParseString` calls to resolve variables, expressions, and validate properties.
- MSS selectors use CartoCSS syntax: `#layer::attachment[filter][zoom>=12]`.
- Zoom levels are bitmask-based (`ZoomRange int32`); `AllZoom` = all 31 levels, `InvalidZoom` = 0.
- `spec.go` defines ~150 valid properties with type validators.
- `Datasource` types: PostGIS, Shapefile, SQLite, OGR, GDAL, GeoJson.
- `builder/cache.go` watches file mtimes for automatic rebuilds.

## What's missing / not obvious

- README is a stub (single title line). Don't rely on it for documentation.
- No CI workflows or Makefile (except `tests/Makefile` which compares against a binary not in this repo).
- Commits use generic "update" messages — review code diffs, not commit messages.
- The `mapnik` package is the only concrete `builder.Map` implementation in this repo.
- `mapnik/` uses `github.com/flywave/flywave-mapnik` with local `replace` directive (`../flywave-mapnik`); builds native mapnik `Style`/`Rule`/`Symbolizer` objects via C API instead of XML.
- No XML serialization code exists anymore. `Write`/`WriteFiles` are no-ops kept only for `builder.MapWriter` interface compat.
- `go build ./...` succeeds without C libraries if flywave-mapnik is precompiled (its CGo headers/libs must be present at build time).

### Native API coverage by symbolizer type

| Type | Native coverage | Limitation |
|------|-----------------|------------|
| Line, Polygon, Marker, Point, Building, Dot | **100%** | All CartoCSS properties set via `keys` enum |
| Line, Polygon, Marker, Point, Building, Dot | **100%** | All CartoCSS properties set via `keys` enum |
| PolygonPattern, LinePattern | **100%** | All properties work natively |
| Text, Shield | **100%** | All properties work natively via `text_placements_` C API |
| Raster | **100%** | Opacity, scaling, comp-op, mesh-size, filter-factor via `keys` enum. Colorizer stops, epsilon, default-color/mode via `raster_colorizer` C API |

Text/shield properties are stored in a `text_placements_` composite object. The C API exposes `mapnik_symbolizer_text_init`/`set_name`/`set_face_name`/`set_double`/`set_string`/`set_color` functions that create and configure the placements object directly (no XML needed).
