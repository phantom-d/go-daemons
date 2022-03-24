package daemons

import (
	"github.com/phantom-d/go-daemons/config"
	"os"
	"syscall"
)

type Watcher struct {
	*DaemonData
}

func (watcher *Watcher) SetData(data *DaemonData) {
	watcher.DaemonData = data
}

func (watcher *Watcher) Run() (err error) {
	for _, cfg := range watcher.Workers {
		if daemon := New(cfg.Name); daemon != nil {
			var dm *os.Process
			dm, err = daemon.Data().Context.Search()
			if err != nil {
				config.Log().Error().Err(err).Msgf("Exec daemon '%s'", cfg.Name)
			} else if dm != nil {
				err = dm.Signal(syscall.Signal(0))
				if err == os.ErrProcessDone {
					dm = nil
				}
			}
			if dm == nil {
				if err = Exec(daemon); err != nil {
					config.Log().Error().Err(err).Msgf("Exec daemon '%s'", cfg.Name)
					err = nil
				}
			}
		}
	}
	return
}
