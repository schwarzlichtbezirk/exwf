package exwf

import (
	"log"
	"net/http"
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

	req *http.Request
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

// ReadConfig reads all config YAML-files given at command line.
func ReadConfig() (err error) {
	var fpath string
	// try to read files given at command line
	for _, fpath = range os.Args[1:] {
		var thr []*Chain
		if thr, err = ReadYaml(fpath); err != nil {
			return
		}
		Threads = append(Threads, thr...)
	}
	if len(Threads) > 0 {
		return
	}
	// try to read config current directory
	fpath = "exwf.yaml"
	if ok, _ := pathexists(fpath); ok {
		Threads, err = ReadYaml(fpath)
		return
	}
	// try to read config at binary location
	fpath = filepath.Join(filepath.Dir(os.Args[0]), "exwf.yaml")
	if ok, _ := pathexists(fpath); ok {
		Threads, err = ReadYaml(fpath)
		return
	}
	// if GOPATH is present
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		// try to read config at GOPATH binary location
		fpath = envfmt("${GOPATH}/bin/exwf.yaml")
		if ok, _ := pathexists(fpath); ok {
			Threads, err = ReadYaml(fpath)
			return
		}
		// try to read config from source code
		fpath = envfmt("${GOPATH}/src/github.com/schwarzlichtbezirk/exwf/exwf.yaml")
		if ok, _ := pathexists(fpath); ok {
			Threads, err = ReadYaml(fpath)
			return
		}
	}
	return
}

// The End.
