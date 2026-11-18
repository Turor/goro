package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gogpu/ui/plugin"
	"github.com/kivutar/goro/ui/rotheme"
)

// maxUIFontFileSize bounds a [text] font_file; UI fonts are well under 2 MiB.
const maxUIFontFileSize = 16 << 20

// uiFontState remembers the last font file registered, so a reload that does
// not change it costs one stat.
type uiFontState struct {
	path    string
	family  string
	modTime time.Time
	size    int64
}

var loadedUIFont uiFontState

// loadUIFont registers [text] font_file under [text] font_family. With no
// font_file (the default) it does nothing. A font that cannot be read or
// parsed is reported and the UI keeps drawing with the built-in font, since
// an unknown family falls back to Inter.
func loadUIFont(layoutPath string) error {
	file := rotheme.FontFile
	if file == "" {
		return nil
	}
	if !filepath.IsAbs(file) {
		file = filepath.Join(filepath.Dir(layoutPath), file)
	}
	info, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("font_file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("font_file %s is not a regular file", file)
	}
	if info.Size() > maxUIFontFileSize {
		return fmt.Errorf("font_file %s is %d bytes; the limit is %d", file, info.Size(), maxUIFontFileSize)
	}
	family := rotheme.Default.Typography.FontFamily
	next := uiFontState{path: file, family: family, modTime: info.ModTime(), size: info.Size()}
	if next == loadedUIFont {
		return nil
	}
	data, err := readUIFontFile(file)
	if err != nil {
		return fmt.Errorf("font_file: %w", err)
	}
	if len(data) > maxUIFontFileSize {
		return fmt.Errorf("font_file %s is over the %d byte limit", file, maxUIFontFileSize)
	}
	if err := plugin.NewDefaultPluginContext().Assets.LoadFont(family, data); err != nil {
		return fmt.Errorf("font_file %s: %w", file, err)
	}
	loadedUIFont = next
	return nil
}

// readUIFontFile reads at most maxUIFontFileSize+1 bytes, the same way the
// layout file is read, so a file replaced or grown after the Stat cannot make
// it allocate more. A file that is no longer regular once opened (swapped for
// a FIFO or device) is rejected before any read.
func readUIFontFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", path)
	}
	return io.ReadAll(io.LimitReader(f, maxUIFontFileSize+1))
}
