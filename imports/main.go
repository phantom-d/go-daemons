package imports

import (
	"github.com/phantom-d/go-daemons/config"

	"github.com/mitchellh/mapstructure"

	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"syscall"
	"time"
)

func New(cfg config.Worker, parent string, params map[string]interface{}) WorkerInterface {
	if cfg.Enabled {
		w := Factory.CreateInstance(cfg.Name)
		wd := &Worker{Params: params, Parent: parent}
		err := mapstructure.Decode(cfg, &wd)
		if err != nil {
			config.Log().Info().Msg("Worker load config")
			return nil
		}
		pidFileName, err := filepath.Abs(fmt.Sprintf("%s/%s_%s.pid", config.Cfg().PidDir, parent, cfg.Name))
		if err != nil {
			config.Log().Fatal().Err(err).Msgf("Init daemon '%s'", cfg.Name)
		}
		var args []string
		notExists := true
		daemonArg := "--daemon=" + parent

		for _, arg := range os.Args {
			if matched, _ := regexp.MatchString(`--migrate`, arg); matched {
				continue
			}
			if matched, _ := regexp.MatchString(`--daemon=`, arg); matched {
				arg = daemonArg
				notExists = false
			}
			args = append(args, arg)
		}
		if notExists {
			args = append(args, daemonArg)
		}
		args = append(args, "--worker="+cfg.Name)
		wd.Context = &config.Context{
			Name:        cfg.Name,
			Type:        `worker`,
			PidFileName: pidFileName,
			PidFilePerm: 0644,
			WorkDir:     "./",
			Args:        args,
		}
		w.SetData(wd)
		return w
	} else {
		config.Log().Info().Msgf("Worker '%s' is disabled!", cfg.Name)
	}
	return nil
}

func Run(w WorkerInterface) (err error) {
	var cancel context.CancelFunc
	wd := w.Data()
	err = wd.Context.CreatePidFile()
	if err != nil {
		config.Log().Fatal().Err(err).Msgf("Worker '%s' Process", wd.Name)
	}
	wd.ctx, cancel = context.WithCancel(context.Background())
	wd.signalChan = make(chan os.Signal, 1)
	signal.Notify(wd.signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	defer func() {
		signal.Stop(wd.signalChan)
		cancel()
	}()

	go func() {
		for {
			select {
			case s := <-wd.signalChan:
				switch s {
				case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
					config.Log().Info().Msgf("worker '%s' terminate", wd.Name)
					cancel()
					wd.Terminate(s)
					err := wd.Context.Release()
					if err != nil {
						config.Log().Error().Err(err).Msgf("Worker '%s' terminate", wd.Name)
					}
					os.Exit(1)
				}
			case <-wd.ctx.Done():
				config.Log().Info().Msgf("worker '%s' is done", wd.Name)
				os.Exit(1)
			}
		}
	}()

	config.Log().Info().Msgf("Start worker '%s'!", wd.Name)
	for {
		select {
		case <-wd.ctx.Done():
			return
		case <-time.Tick(wd.Sleep):
			var result ResultProcess
			runtime.GC()
			memStats := &runtime.MemStats{}
			runtime.ReadMemStats(memStats)
			_, err = wd.BeforeRun()
			if err != nil {
				config.Log().Error().Err(err).Msgf("Worker '%s' processing BeforeRun", wd.Name)
			}
			if memStats.Alloc > wd.MemoryLimit {
				break
			}
			timeStart := time.Now()
			data, errorData := w.GetEntities()
			runtime.ReadMemStats(memStats)
			for errorData == nil && data != nil {
				result := ResultProcess{Queue: wd.Queue}
				if memStats.Alloc > wd.MemoryLimit {
					break
				}
				err := w.BeforeProcessing(&data)
				if err != nil {
					config.Log().Error().Err(err).Msgf("Worker '%s' processing BeforeProcessing", wd.Name)
				}
				if data != nil {
					err = w.Processing(data, &result)
					if err != nil {
						config.Log().Error().Err(err).Msgf("Worker '%s' processing", wd.Name)
					}
				}
				err = w.AfterProcessing(result.ErrorItems)
				if err != nil {
					config.Log().Error().Err(err).Msgf("Worker '%s' processing AfterProcessing", wd.Name)
				}
				timeStart = time.Now()
				runtime.GC()
				runtime.ReadMemStats(memStats)
				data, errorData = w.GetEntities()
			}

			if errorData != nil || data == nil {
				result := ResultProcess{Queue: wd.Queue}
				result.Duration = time.Now().Sub(timeStart)
				runtime.ReadMemStats(memStats)
				result.Memory = memStats.Alloc
			}
			runtime.GC()
			runtime.ReadMemStats(memStats)
			err = wd.AfterRun(&result)
			if err != nil {
				config.Log().Error().Err(err).Msgf("Worker '%s' processing AfterRun", wd.Name)
			}
		}
	}
}

// Execute daemon as a new system process
func (w *Worker) Run() (err error) {
	_, err = w.Context.Run()
	return
}

// Return worker settings
func (w *Worker) Data() *Worker {
	return w
}

func (w *Worker) GetStatus() (result bool, err error) {
	return w.Context.GetStatus()
}

func (w *Worker) Terminate(s os.Signal) {
}

func (w *Worker) BeforeRun() (result interface{}, err error) {
	return
}

func (w *Worker) AfterRun(result *ResultProcess) (err error) {
	return
}

func (w *Worker) BeforeProcessing(data interface{}) (result interface{}, err error) {
	return
}

func (w *Worker) AfterProcessing(errorItems interface{}) (result interface{}, err error) {
	return
}
