package config

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"

	cartocss "github.com/flywave/go-cartocss"
)

type CartoCSS struct {
	Mapnik      Mapnik
	StylesDir   string `toml:"styles_dir"`
	OutDir      string `toml:"out_dir"`
	Datasources Datasource
	PostGIS     PostGIS
	BaseDir     string
}

type Mapnik struct {
	PluginDirs []string `toml:"plugin_dirs"`
	FontDirs   []string `toml:"font_dirs"`
}

type Datasource struct {
	ShapefileDirs []string `toml:"shapefile_dirs"`
	SQLiteDirs    []string `toml:"sqlite_dirs"`
	ImageDirs     []string `toml:"image_dirs"`
	DataDirs      []string `toml:"data_dirs"`
	FontDirs      []string `toml:"font_dirs"`
}

type PostGIS struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
	SRID     string
}

type Locator interface {
	Font(string) string
	SQLite(string) string
	Shape(string) string
	Image(string) string
	Data(string) string
	PostGIS(cartocss.PostGIS) cartocss.PostGIS
	SetBaseDir(string)
	SetOutDir(string)
	UseRelPaths(bool)
	MissingFiles() []string
}

func Load(fileName string) (*CartoCSS, error) {
	config := CartoCSS{}
	config.BaseDir = filepath.Dir(fileName)
	_, err := toml.DecodeFile(fileName, &config)
	if err != nil {
		return &config, err
	}

	if !filepath.IsAbs(config.StylesDir) {
		config.StylesDir = filepath.Join(config.BaseDir, config.StylesDir)
	}
	if !filepath.IsAbs(config.OutDir) {
		config.OutDir = filepath.Join(config.BaseDir, config.OutDir)
	}

	return &config, nil
}

func (m *CartoCSS) Load(fileName string) error {
	_, err := toml.DecodeFile(fileName, &m)
	if err != nil {
		return err
	}
	return nil
}

func (m *CartoCSS) Locator() Locator {
	locator := &LookupLocator{baseDir: m.BaseDir}
	for _, dir := range m.Datasources.SQLiteDirs {
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(m.BaseDir, dir)
		}
		locator.AddSQLiteDir(dir)
	}
	for _, dir := range m.Datasources.ImageDirs {
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(m.BaseDir, dir)
		}
		locator.AddImageDir(dir)
	}
	for _, dir := range m.Datasources.ShapefileDirs {
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(m.BaseDir, dir)
		}
		locator.AddShapeDir(dir)
	}
	for _, dir := range m.Datasources.DataDirs {
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(m.BaseDir, dir)
		}
		locator.AddDataDir(dir)
	}

	fontDirs := append([]string{}, m.Mapnik.FontDirs...)
	fontDirs = append(fontDirs, m.Datasources.FontDirs...)
	for _, dir := range fontDirs {
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(m.BaseDir, dir)
		}
		locator.AddFontDir(dir)
	}
	locator.SetPGConfig(m.PostGIS)
	return locator
}

type LookupLocator struct {
	fontDirs   []string
	sqliteDirs []string
	shapeDirs  []string
	imageDirs  []string
	dataDirs   []string
	pgConfig   *PostGIS
	baseDir    string
	outDir     string
	relative   bool
	missing    map[string]struct{}
}

func (l *LookupLocator) SetBaseDir(dir string) {
	l.baseDir = dir
}

func (l *LookupLocator) SetOutDir(dir string) {
	l.outDir = dir
}

func (l *LookupLocator) UseRelPaths(rel bool) {
	l.relative = rel
}

func (l *LookupLocator) find(basename string, dirs []string) (string, bool) {
	check := func(dir string) string {
		fname := filepath.Join(dir, basename)
		if _, err := os.Stat(fname); err == nil {
			return fname
		}
		return ""
	}

	fname, ok := func() (string, bool) {
		if filepath.IsAbs(basename) {
			if fname := check(""); fname != "" {
				return fname, true
			}
		}

		for _, d := range dirs {
			if fname := check(d); fname != "" {
				return fname, true
			}
		}
		for _, d := range l.dataDirs {
			if fname := check(d); fname != "" {
				return fname, true
			}
		}

		if fname := check(l.baseDir); fname != "" {
			return fname, true
		}

		return "", false
	}()

	if !ok {
		if l.missing == nil {
			l.missing = make(map[string]struct{})
		}
		l.missing[basename] = struct{}{}
		fname = basename
	} else {
		absfname, err := filepath.Abs(fname)
		if err == nil {
			fname = absfname
		}
	}

	if l.relative {
		relfname, err := filepath.Rel(l.outDir, fname)
		if err == nil {
			fname = relfname
		}
	} else {
		if !filepath.IsAbs(fname) {
			fname = filepath.Join(l.outDir, fname)
		}
	}
	return fname, ok

}

func (l *LookupLocator) AddFontDir(dir string) {
	l.fontDirs = append(l.fontDirs, dir)
}
func (l *LookupLocator) AddSQLiteDir(dir string) {
	l.sqliteDirs = append(l.sqliteDirs, dir)
}
func (l *LookupLocator) AddShapeDir(dir string) {
	l.shapeDirs = append(l.shapeDirs, dir)
}
func (l *LookupLocator) AddImageDir(dir string) {
	l.imageDirs = append(l.imageDirs, dir)
}
func (l *LookupLocator) AddDataDir(dir string) {
	l.dataDirs = append(l.dataDirs, dir)
}
func (l *LookupLocator) SetPGConfig(pgConfig PostGIS) {
	l.pgConfig = &pgConfig
}

func (l *LookupLocator) Font(basename string) string {
	for _, variation := range fontVariations(basename, ".ttf") {
		if file, ok := l.find(variation, l.fontDirs); ok {
			return file
		} else {
			delete(l.missing, variation)
		}
	}
	l.missing[basename] = struct{}{}
	return ""
}

func (l *LookupLocator) SQLite(basename string) string {
	fname, _ := l.find(basename, l.sqliteDirs)
	return fname
}
func (l *LookupLocator) Shape(basename string) string {
	fname, _ := l.find(basename, l.shapeDirs)
	return fname
}
func (l *LookupLocator) Image(basename string) string {
	fname, _ := l.find(basename, l.imageDirs)
	return fname
}
func (l *LookupLocator) Data(basename string) string {
	fname, _ := l.find(basename, nil)
	return fname
}

func (l *LookupLocator) PostGIS(ds cartocss.PostGIS) cartocss.PostGIS {
	if l.pgConfig == nil {
		return ds
	}
	c := l.pgConfig
	if c.Host != "" {
		ds.Host = c.Host
	}
	if c.Port != "" {
		ds.Port = c.Port
	}
	if c.Database != "" {
		ds.Database = c.Database
	}
	if c.Username != "" {
		ds.Username = c.Username
	}
	if c.Password != "" {
		ds.Password = c.Password
	}
	if c.SRID != "" {
		ds.SRID = c.SRID
	}

	return ds
}

func (l *LookupLocator) MissingFiles() []string {
	if len(l.missing) == 0 {
		return nil
	}
	files := make([]string, 0, len(l.missing))
	for f := range l.missing {
		files = append(files, f)
	}
	sort.Strings(files)
	return files
}

var _ Locator = &LookupLocator{}

func fontVariations(font, suffix string) []string {
	parts := strings.Split(font, " ")
	var result []string

	result = append(result, strings.Join(parts, "")+suffix)

	for i := 1; i < len(parts); i++ {
		result = append(result,
			strings.Join(parts[:i], "")+"-"+strings.Join(parts[i:], "")+suffix,
		)
	}

	if len(parts) > 1 {
		result = append(result, strings.Join(parts[:len(parts)-1], "")+suffix)
	}

	return result
}
