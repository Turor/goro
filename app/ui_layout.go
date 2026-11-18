package app

import (
	"fmt"
	"io"
	"os"

	"github.com/kivutar/goro/config"
	"github.com/kivutar/goro/glog"
	"github.com/kivutar/goro/ui/rotheme"
)

// loadUILayout applies the optional ui-layout.ini override once at startup.
// A missing file costs one failed open and keeps the built-in layout; an
// invalid file is logged and also keeps the built-in layout. It never fails
// startup.
func loadUILayout(cfg config.Config) {
	path := cfg.UILayoutPath()
	applied, warnings, err := rotheme.LoadLayoutFile(path)
	// ParseLayout already caps warnings at MaxLayoutWarnings plus a total.
	for _, warning := range warnings {
		glog.Warnf("ui layout: %s", warning)
	}
	switch {
	case err != nil:
		glog.Errorf("ui layout rejected, using built-in layout: %v", err)
	case applied:
		glog.Infof("ui layout loaded path=%s", path)
	}
}

// DumpUILayout writes every layout key with its built-in default to path, or
// to stdout when path is "-".
func DumpUILayout(path string) error {
	if path == "-" {
		return rotheme.WriteDefaultLayout(os.Stdout)
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("dump ui layout: %w", err)
	}
	if err := writeAndClose(file, rotheme.WriteDefaultLayout); err != nil {
		return fmt.Errorf("dump ui layout: %w", err)
	}
	return nil
}

func writeAndClose(file *os.File, write func(io.Writer) error) error {
	if err := write(file); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}
