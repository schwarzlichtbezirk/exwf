package exwf

import (
	"errors"
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	yaml "gopkg.in/yaml.v3"
)

// Entry - HTTP request description.
type Entry struct {
	URL    string `yaml:"url"`
	Method string `yaml:"method,omitempty"`
	Data   string `yaml:"data,omitempty"`
	Token  string `yaml:"token,omitempty"`

	DelayMin time.Duration `yaml:"delay-min,omitempty"`
	DelayMax time.Duration `yaml:"delay-max,omitempty"`
	WaitRpl  bool          `yaml:"wait-reply,omitempty"`
}

// Chain - is the consistent chain of requests entries.
type Chain struct {
	Entries []*Entry `yaml:"entries"`
	Repeats int      `yaml:"repeats,omitempty"`
}

var (
	// ReqCount - request counter.
	ReqCount int64
)

const (
	cfgenv  = "EXWFCONFIGPATH"
	cfgfile = "exwf.yaml"
	cfgbase = "."
	srcpath = "src/github.com/schwarzlichtbezirk/exwf"
)

// Path given from command line parameter.
var cmdpath = flag.String("c", "", "configuration path")

// ConfigPath determines configuration path, depended on what directory is exist.
var ConfigPath string

// ErrNoCongig is "no configuration path was found" error message.
var ErrNoCongig = errors.New("no configuration path was found")

// DetectConfigPath finds configuration path with existing configuration file at least.
func DetectConfigPath() (cfgpath string, err error) {
	var path string
	var exepath = filepath.Dir(os.Args[0])

	// try to get from environment setting
	if path = envfmt(os.Getenv(cfgenv)); path != "" {
		// try to get access to full path
		if ok, _ := pathexists(filepath.Join(path, cfgfile)); ok {
			cfgpath = path
			return
		}
		// try to find relative from executable path
		path = filepath.Join(exepath, path)
		if ok, _ := pathexists(filepath.Join(path, cfgfile)); ok {
			cfgpath = exepath
			return
		}
		log.Printf("no access to pointed configuration path '%s'\n", path)
	}

	// try to get from command path arguments
	if path = *cmdpath; path != "" {
		// try to get access to full path
		if ok, _ := pathexists(filepath.Join(path, cfgfile)); ok {
			cfgpath = path
			return
		}
		// try to find relative from executable path
		path = filepath.Join(exepath, path)
		if ok, _ := pathexists(filepath.Join(path, cfgfile)); ok {
			cfgpath = path
			return
		}
	}

	// try to get from config subdirectory on executable path
	path = filepath.Join(exepath, cfgbase)
	if ok, _ := pathexists(filepath.Join(path, cfgfile)); ok {
		cfgpath = path
		return
	}
	// try to find in executable path
	if ok, _ := pathexists(filepath.Join(exepath, cfgfile)); ok {
		cfgpath = exepath
		return
	}
	// try to find in current path
	if ok, _ := pathexists(cfgfile); ok {
		cfgpath = "."
		return
	}

	// if GOPATH is present
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		// try to get from go bin config
		path = filepath.Join(gopath, "bin", cfgbase)
		if ok, _ := pathexists(filepath.Join(path, cfgfile)); ok {
			cfgpath = path
			return
		}
		// try to get from go bin root
		path = filepath.Join(gopath, "bin")
		if ok, _ := pathexists(filepath.Join(path, cfgfile)); ok {
			cfgpath = path
			return
		}
		// try to get from source code
		path = filepath.Join(gopath, srcpath)
		if ok, _ := pathexists(filepath.Join(path, cfgfile)); ok {
			cfgpath = path
			return
		}
	}

	// no config was found
	err = ErrNoCongig
	return
}

// ReadYaml reads "data" object from YAML-file with given file path.
func ReadYaml(fname string, data interface{}) (err error) {
	var body []byte
	if body, err = os.ReadFile(filepath.Join(ConfigPath, fname)); err != nil {
		return
	}
	if err = yaml.Unmarshal(body, data); err != nil {
		return
	}
	return
}

// The End.
