package imports

import (
	"context"
	"os"
	"time"

	"github.com/phantom-d/go-daemons/config"
)

type Worker struct {
	Name        string        `mapstructure:"Name"`
	MemoryLimit uint64        `mapstructure:"MemoryLimit"`
	Queue       string        `mapstructure:"Queue"`
	Enabled     bool          `mapstructure:"Enabled"`
	Sleep       time.Duration `mapstructure:"Sleep"`
	Params      map[string]interface{}
	Parent      string
	Context     *config.Context
	ctx         context.Context
	signalChan  chan os.Signal
	done        chan struct{}
}

type WorkerInterface interface {
	BeforeRun() (interface{}, error)
	AfterRun() (interface{}, error)
	BeforeIteration(interface{}) (interface{}, error)
	AfterIteration(*ResultProcess) error
	Data() *Worker
	GetEntities() (interface{}, error)
	GetStatus() (bool, error)
	Run() error
	SetData(worker *Worker)
	Processing(interface{}, *ResultProcess) error
	ExtractId(interface{}) ([]string, error)
	Terminate(os.Signal)
}

type ResultProcess struct {
	Queue      string
	Duration   time.Duration
	Total      int
	Memory     uint64
	ErrorItems []interface{}
}

type FactoryStore map[string]func() WorkerInterface

var Factory = make(FactoryStore)

func (factory *FactoryStore) Register(name string, factoryFunc func() WorkerInterface) {
	(*factory)[name] = factoryFunc
}

func (factory *FactoryStore) CreateInstance(name string) (result WorkerInterface) {
	if factoryFunc, ok := (*factory)[name]; ok {
		result = factoryFunc()
	}
	return
}
