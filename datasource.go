package cartocss

type PostGIS struct {
	Id            string
	Host          string
	Port          string
	Database      string
	Username      string
	Password      string
	Query         string
	SRID          string
	GeometryField string
	Extent        string
}

type Shapefile struct {
	Id       string
	Filename string
	SRID     string
}

type SQLite struct {
	Id            string
	Filename      string
	SRID          string
	Query         string
	GeometryField string
	Extent        string
}

type OGR struct {
	Id       string
	Filename string
	SRID     string
	Layer    string
	Query    string
	Extent   string
}

type GDAL struct {
	Id         string
	Filename   string
	SRID       string
	Extent     string
	Band       string
	Processing []string
}

type GeoJson struct {
	Id       string
	Filename string
}

const (
	DATASET        = "dataset"
	DATASET_RASTER = "dataset_raster"
)

type Dataset struct {
	Id   string
	Name string
	Type string
}

func (d *Dataset) GetId() string {
	return d.Id
}
func (d *Dataset) GetType() string {
	return d.Type
}
func (d *Dataset) GetName() string {
	return d.Name
}

type DatasetRaster struct {
	Dataset
	Multi      bool
	Lox        float64
	Loy        float64
	Hix        float64
	Hiy        float64
	Tilesize   uint32
	TileStride uint32
	Zoom       uint32
}

type Datasource interface {
	GetId() string
	GetType() string
	GetName() string
}
