package cartocss

type GeometryType string

const (
	Unknown    GeometryType = "Unknown"
	LineString GeometryType = "LineString"
	Polygon    GeometryType = "Polygon"
	Point      GeometryType = "Point"
	Raster     GeometryType = "Raster"
)

type Layer struct {
	ID              string
	CssIds          []string
	Classes         []string
	SRS             *string
	Datasource      interface{}
	Type            GeometryType
	Active          bool
	GroupBy         string
	ClearLabelCache bool
	PostLabelCache  bool
	CacheFeatures   bool
	ScaleFactor     float64
	Maxzoom         uint32
	Minzoom         uint32
}
