package cartocss

type PostGIS struct {
	Id                 string
	Host               string
	Port               string
	Database           string
	Username           string
	Password           string
	Query              string
	SRID               string
	GeometryField      string
	Extent             string
	SimplifyGeometries string
}

func (p *PostGIS) GetId() string {
	return p.Id
}

func (p *PostGIS) GetType() string {
	return "postgis"
}

func (p *PostGIS) GetName() string {
	return p.Id
}

type Shapefile struct {
	Id       string
	Filename string
	SRID     string
}

func (s *Shapefile) GetId() string {
	return s.Id
}

func (s *Shapefile) GetType() string {
	return "shapefile"
}

func (s *Shapefile) GetName() string {
	return s.Id
}

type SQLite struct {
	Id            string
	Filename      string
	SRID          string
	Query         string
	GeometryField string
	Extent        string
}

func (s *SQLite) GetId() string {
	return s.Id
}

func (s *SQLite) GetType() string {
	return "sqlite"
}

func (s *SQLite) GetName() string {
	return s.Id
}

type OGR struct {
	Id       string
	Filename string
	SRID     string
	Layer    string
	Query    string
	Extent   string
}

func (o *OGR) GetId() string {
	return o.Id
}

func (o *OGR) GetType() string {
	return "ogr"
}

func (o *OGR) GetName() string {
	return o.Id
}

type GDAL struct {
	Id         string
	Filename   string
	SRID       string
	Extent     string
	Band       string
	Processing []string
}

func (g *GDAL) GetId() string {
	return g.Id
}

func (g *GDAL) GetType() string {
	return "gdal"
}

func (g *GDAL) GetName() string {
	return g.Id
}

type GeoJson struct {
	Id       string
	Filename string
}

func (g *GeoJson) GetId() string {
	return g.Id
}

func (g *GeoJson) GetType() string {
	return "geojson"
}

func (g *GeoJson) GetName() string {
	return g.Id
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
