package daemons

import (
	"context"
	"github.com/phantom-d/go-daemons/config"
	"os"
	"time"
)

type DaemonInterface interface {
	Run() error
	Data() *DaemonData
	SetData(*DaemonData)
	Terminate(os.Signal)
}

type DaemonData struct {
	Name        string                 `mapstructure:"Name"`
	MemoryLimit uint64                 `mapstructure:"MemoryLimit"`
	Workers     []config.Worker        `mapstructure:"Workers"`
	Params      map[string]interface{} `mapstructure:"Params"`
	Sleep       time.Duration          `mapstructure:"Sleep"`
	Context     *config.Context
	ctx         context.Context
	signalChan  chan os.Signal
	done        chan struct{}
}

type DaemonStatus struct {
	Count struct {
		Current int8 `json:"current"`
		Total   int8 `json:"total"`
	} `json:"count"`
}

type Factory map[string]func() DaemonInterface

var factory = make(Factory)

func init() {
	factory.Register("watcher", func() DaemonInterface { return &Watcher{} })
	factory.Register("import", func() DaemonInterface { return &Import{} })
}

func (factory *Factory) Register(name string, factoryFunc func() DaemonInterface) {
	(*factory)[name] = factoryFunc
}

func (factory *Factory) CreateInstance(name string) (result DaemonInterface) {
	if factoryFunc, ok := (*factory)[name]; ok {
		result = factoryFunc()
	}
	return
}
