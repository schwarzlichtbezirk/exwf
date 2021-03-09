package exwf

import (
	"os"
	"path/filepath"
	"regexp"
)

var efre = regexp.MustCompile(`\$\{\w+\}`)

func envfmt(p string) string {
	return filepath.ToSlash(efre.ReplaceAllStringFunc(p, func(name string) string {
		return os.Getenv(name[2 : len(name)-1]) // strip ${...} and replace by env value
	}))
}

func pathexists(path string) (bool, error) {
	var err error
	if _, err = os.Stat(path); err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// The End.
