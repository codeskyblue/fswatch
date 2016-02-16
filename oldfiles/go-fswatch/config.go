package fswatch

import (
	"os"
	"regexp"
	"time"
	"io/ioutil"

	"github.com/go-fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

// Config of fswatch
type Config struct {
	Paths     []string `yaml:"paths"`
	Depth     int      `yaml:"depth"`
	Exclude   []string `yaml:"exclude"`
	reExclude []*regexp.Regexp
	Include   []string `yaml:"include"`
	reInclude []*regexp.Regexp
	bufdur    time.Duration `yaml:"-"`
	Command   interface{}   `yaml:"command"` // can be string or []string
	cmd       []string
	Env       map[string]string `yaml:"env"`

	AutoRestart     bool          `yaml:"autorestart"`
	Delay time.Duration `yaml:"delay"`
	Signal      string        `yaml:"signal"`

	w       *fsnotify.Watcher
	modtime map[string]time.Time
	sig     chan string
	sigOS   chan os.Signal
}

// WriteFile Save config
func (c *Config) WriteFile(filename string) error {
	data, err := yaml.Marshal(c)
	if err!= nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}