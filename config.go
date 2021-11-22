package exwf

import (
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

// ReadYaml reads thread object from YAML-file with given file path.
func ReadYaml(fpath string) (thr []*Chain, err error) {
	var body []byte
	if body, err = os.ReadFile(fpath); err != nil {
		return
	}
	if err = yaml.Unmarshal(body, &thr); err != nil {
		return
	}
	for _, chain := range thr {
		for _, ent := range chain.Entries {
			if ent.Method == "" {
				if ent.Data == "" {
					ent.Method = "GET"
				} else {
					ent.Method = "POST"
				}
			}
		}
		if chain.Repeats == 0 {
			chain.Repeats = -1
		}
	}
	log.Printf("readed file: '%s', threads: %d\n", fpath, len(thr))
	return
}

const (
	cfgenv  = "EXWFCONFIGPATH"
	cfgfile = "exwf.yaml"
	srcpath = "src/github.com/schwarzlichtbezirk/exwf"
)

// ReadConfig reads all config YAML-files given at command line.
func ReadConfig() (err error) {
	var fpath string
	var exepath = filepath.Dir(os.Args[0])

	// try to get from environment setting
	if path := envfmt(os.Getenv(cfgenv)); path != "" {
		// try to get access to full path
		fpath = filepath.Join(path, cfgfile)
		if ok, _ := pathexists(fpath); ok {
			Threads, err = ReadYaml(fpath)
			return
		}
		// try to find relative from executable path
		fpath = filepath.Join(exepath, path, cfgfile)
		if ok, _ := pathexists(fpath); ok {
			Threads, err = ReadYaml(fpath)
			return
		}
		log.Printf("no access to pointed configuration path '%s'\n", path)
	}

	// try to get from command path arguments
	for _, path := range os.Args[1:] {
		var thr []*Chain
		if thr, err = ReadYaml(filepath.Join(path, cfgfile)); err != nil {
			return
		}
		Threads = append(Threads, thr...)
	}
	if len(Threads) > 0 {
		return
	}

	// try to find in executable path
	fpath = filepath.Join(exepath, cfgfile)
	if ok, _ := pathexists(fpath); ok {
		Threads, err = ReadYaml(fpath)
		return
	}
	// try to find in current path
	fpath = filepath.Join(".", cfgfile)
	if ok, _ := pathexists(fpath); ok {
		Threads, err = ReadYaml(fpath)
		return
	}
	// if GOPATH is present
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		// try to get from go bin config
		fpath = filepath.Join(gopath, "bin", cfgfile)
		if ok, _ := pathexists(fpath); ok {
			Threads, err = ReadYaml(fpath)
			return
		}
		// try to get from source code
		fpath = filepath.Join(gopath, srcpath, cfgfile)
		if ok, _ := pathexists(fpath); ok {
			Threads, err = ReadYaml(fpath)
			return
		}
	}
	return
}

// The End.
