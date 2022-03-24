package daemons

import (
	"github.com/phantom-d/go-daemons/config"
	"github.com/phantom-d/go-daemons/imports"
	"os"
	"syscall"
)

type Import struct {
	*DaemonData
}

func (imp *Import) SetData(data *DaemonData) {
	imp.DaemonData = data
}

func (imp *Import) Run() (err error) {
	for _, cfg := range imp.Workers {
		if worker := imports.New(cfg, imp.Name, imp.Params); worker != nil {
			wd := worker.Data()
			if config.Cfg().Worker == "" || config.Cfg().Worker == wd.Name {
				var dm *os.Process
				dm, err = wd.Context.Search()
				if err != nil {
					config.Log().Error().Err(err).Msgf("Exec worker '%s'", cfg.Name)
				} else if dm != nil {
					err = dm.Signal(syscall.Signal(0))
					if err == os.ErrProcessDone {
						dm = nil
					}
				}
				if dm == nil {
					if config.Cfg().Worker == wd.Name {
						if err = imports.Run(worker); err != nil {
							config.Log().Error().Err(err).Msgf("Start worker '%s'", cfg.Name)
							err = nil
						}
						break
					} else {
						if err = worker.Run(); err != nil {
							config.Log().Error().Err(err).Msgf("Exec worker '%s'", cfg.Name)
							err = nil
						}
					}
				}
			}
		}
	}
	return
}

func (imp *Import) Terminate(s os.Signal) {
	for _, cfg := range imp.Workers {
		if worker := imports.New(cfg, imp.Name, imp.Params); worker != nil {
			wd := worker.Data()
			dm, err := wd.Context.Search()
			config.Log().Debug().Msgf("Terminate worker process: '%+v'", dm)
			config.Log().Debug().Msgf("Terminate worker Context: '%+v'", wd.Context)
			if err == nil {
				if err := dm.Signal(s); err != nil {
					config.Log().Error().Err(err).Msgf("Terminate worker '%s'", cfg.Name)
				}
			} else {
				config.Log().Error().Err(err).Msgf("Terminate worker '%s'", cfg.Name)
			}
		}
	}
	err := imp.Context.Release()
	if err != nil {
		config.Log().Error().Err(err).Msgf("Worker '%s' terminate", imp.Name)
	}
}
