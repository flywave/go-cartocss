package cartocss

type Map struct {
	SRS   string `yaml:"srs"`
	BBOX  []int  `yaml:"bounds"`
	Scale int    `yaml:"scale"`
}
