package mapnik

import (
	"encoding/xml"
	"io"
	"os"
)

func (m *Map) Write(w io.Writer) error {
	e := xml.NewEncoder(w)
	e.Indent("", "  ")
	err := e.Encode(m.XML)
	return err
}

func (m *Map) WriteFiles(basename string) error {
	f, err := os.Create(basename)
	if err != nil {
		return err
	}
	defer f.Close()
	return m.Write(f)
}
