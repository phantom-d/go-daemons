package config

import (
	goflag "flag"
	"github.com/rs/zerolog"
	flag "github.com/spf13/pflag"
	"time"
)

type Config struct {
	PidDir  string
	LogFile string
	Daemon  string
	Worker  string
	Debug   bool
	Daemons map[string]Daemon
	Signal  string
}

type Daemon struct {
	Name        string                 `yaml:"name" mapstructure:"Name"`
	Enabled     bool                   `yaml:"enabled" mapstructure:"Enabled"`
	MemoryLimit uint64                 `yaml:"memory-limit" mapstructure:"MemoryLimit"`
	Sleep       time.Duration          `yaml:"sleep" mapstructure:"Sleep"`
	Workers     []Worker               `yaml:"workers" mapstructure:"Workers"`
	Params      map[string]interface{} `yaml:"params" mapstructure:"Params"`
}

type Worker struct {
	Name        string        `yaml:"name" mapstructure:"Name"`
	MemoryLimit uint64        `yaml:"memory-limit" mapstructure:"MemoryLimit"`
	Queue       string        `yaml:"queue" mapstructure:"Queue"`
	Enabled     bool          `yaml:"enabled" mapstructure:"Enabled"`
	Sleep       time.Duration `yaml:"sleep" mapstructure:"Sleep"`
}

var (
	application *Config = &Config{}
	logger      *zerolog.Logger
)

func Init() {
	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.StringVarP(&application.PidDir, "pid-dir", "p", "pids", "Path to a save pid files")
	flag.StringVarP(&application.Daemon, "daemon", "d", "watcher", "Daemon name to starting")
	flag.StringVarP(&application.Worker, "worker", "w", "", "Warker name to starting")
	flag.BoolVar(&application.Debug, "debug", false, "Enable debug mode")
}
