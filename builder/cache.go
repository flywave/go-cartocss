package builder

import (
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	cartocss "github.com/flywave/go-cartocss"
	"github.com/flywave/go-cartocss/config"
)

type MapMaker interface {
	New(config.Locator) MapWriter
	Type() string
	FileSuffix() string
}

type locatorCreator func() config.Locator

type style struct {
	mapMaker   MapMaker
	mml        string
	mss        []string
	file       string
	lastUpdate time.Time
}

func styleHash(mapType string, mml string, mss []string) uint32 {
	f := fnv.New32()
	f.Write([]byte(mapType))
	f.Write([]byte(mml))
	for i := range mss {
		f.Write([]byte(mss[i]))
	}
	return f.Sum32()
}

func isNewer(file string, timestamp time.Time) bool {
	info, err := os.Stat(file)
	if err != nil {
		return true
	}
	return info.ModTime().After(timestamp)
}

func (s *style) isStale() (bool, error) {
	if s.file == "" {
		return true, nil
	}

	info, err := os.Stat(s.file)
	if err != nil {
		return true, err
	}
	timestamp := info.ModTime()

	if isNewer(s.mml, timestamp) {
		return true, nil
	}
	for _, mss := range s.mss {
		if isNewer(mss, timestamp) {
			return true, nil
		}
	}
	return false, nil
}

const stylePrefix = "carto-style-"

type Cache struct {
	mu         sync.Mutex
	newLocator locatorCreator
	styles     map[uint32]*style
	destDir    string
}

func NewCache(newLocator locatorCreator) *Cache {
	return &Cache{
		newLocator: newLocator,
		styles:     make(map[uint32]*style),
	}
}

func (c *Cache) SetDestination(dest string) {
	c.destDir = dest
}

func (c *Cache) ClearAll() {
	c.ClearTill(time.Now())
}

func (c *Cache) ClearTill(till time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.destDir != "" {
		files, err := filepath.Glob(filepath.Join(c.destDir, stylePrefix+"*"))
		if err != nil {
			log.Println("cleanup error: ", err)
			return
		}
		for _, f := range files {
			if fi, err := os.Stat(f); err == nil && fi.ModTime().Before(till) {
				if err := os.Remove(f); err != nil {
					log.Println("cleanup error: ", err)
				}
			}
		}
	} else {
		for _, style := range c.styles {
			if fi, err := os.Stat(style.file); err == nil && fi.ModTime().Before(till) {
				if err := os.RemoveAll(filepath.Dir(style.file)); err != nil {
					log.Println(err)
				}
			}
		}
	}

	for hash := range c.styles {
		delete(c.styles, hash)
	}
}

type Update struct {
	Err        error
	Time       time.Time
	UpdatedMML bool
}

func (c *Cache) StyleFile(mm MapMaker, mml string, mss []string) (string, error) {
	style, err := c.style(mm, mml, mss)
	if err != nil {
		return "", err
	}
	return style.file, nil
}

func (c *Cache) style(mm MapMaker, mml string, mss []string) (*style, error) {
	hash := styleHash(mm.Type(), mml, mss)
	c.mu.Lock()
	defer c.mu.Unlock()
	if s, ok := c.styles[hash]; ok {
		stale, err := s.isStale()
		if err != nil {
			return nil, err
		}
		if stale {
			if len(mss) == 0 {
				var err error
				s.mss, err = mssFilesFromMML(mml)
				if err != nil {
					return nil, err
				}
			}
			if err := c.build(s); err != nil {
				return nil, err
			}
		}
		return s, nil
	} else {
		if len(mss) == 0 {
			var err error
			mss, err = mssFilesFromMML(mml)
			if err != nil {
				return nil, err
			}
		}
		s = &style{
			mapMaker: mm,
			mml:      mml,
			mss:      mss,
		}
		if err := c.build(s); err != nil {
			return nil, err
		}
		c.styles[hash] = s
		return s, nil
	}
}

type FilesMissingError struct {
	Files []string
}

func (e *FilesMissingError) Error() string {
	return fmt.Sprintf("missing files: %v", e.Files)
}

func (c *Cache) build(style *style) error {
	l := c.newLocator()
	l.SetBaseDir(filepath.Dir(style.mml))
	l.SetOutDir(c.destDir)
	l.UseRelPaths(false)

	m := style.mapMaker.New(l)
	builder := NewBuilder(m)
	builder.SetIncludeInactive(false)

	builder.SetMML(style.mml)
	for _, mss := range style.mss {
		builder.AddMSS(mss)
	}

	if err := builder.Build(); err != nil {
		return err
	}

	if files := l.MissingFiles(); len(files) > 0 {
		return &FilesMissingError{files}
	}

	var styleFile string
	if c.destDir != "" {
		hash := styleHash(style.mapMaker.Type(), style.mml, style.mss)
		styleFile = filepath.Join(c.destDir, fmt.Sprintf("carto-style-%d%s", hash, style.mapMaker.FileSuffix()))
		if err := m.WriteFiles(styleFile); err != nil {
			return err
		}
	} else {
		tmp, err := ioutil.TempDir("", "carto-style")
		if err != nil {
			return err
		}
		styleFile = filepath.Join(tmp, "style"+style.mapMaker.FileSuffix())
		if err := m.WriteFiles(styleFile); err != nil {
			os.RemoveAll(tmp)
			return err
		}
	}
	log.Printf("rebuild style %s as %s with %v\n", style.mml, styleFile, style.mss)
	style.lastUpdate = time.Now()
	style.file = styleFile
	return nil
}

func mssFilesFromMML(mmlFile string) ([]string, error) {
	r, err := os.Open(mmlFile)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	mml, err := cartocss.Parse(r)
	if err != nil {
		return nil, err
	}
	mssFiles := []string{}
	for _, s := range mml.Stylesheets {
		mssFiles = append(mssFiles, filepath.Join(filepath.Dir(mmlFile), s))
	}
	return mssFiles, nil
}
