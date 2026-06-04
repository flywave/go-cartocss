package mapnik

import (
	"io"
	"os"
)

func (m *Map) Write(w io.Writer) error {
	if m.inner == nil {
		return nil
	}
	xml, err := m.inner.SaveToString()
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, xml)
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
