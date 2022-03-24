package config

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// Default file permissions for log and pid files.
const FilePerm = os.FileMode(0640)

// A Context describes daemon context.
type Context struct {
	Name string
	Type string
	// If PidFileName is non-empty, parent process will try to create and lock
	// pid file with given name. Child process writes process id to file.
	PidFileName string
	// Permissions for new pid file.
	PidFilePerm os.FileMode

	// If WorkDir is non-empty, the child changes into the directory before
	// creating the process.
	WorkDir string

	// If Env is non-nil, it gives the environment variables for the
	// daemon-process in the form returned by os.Environ.
	// If it is nil, the result of os.Environ will be used.
	Env []string
	// If Args is non-nil, it gives the command-line args for the
	// daemon-process. If it is nil, the result of os.Args will be used.
	Args []string

	// Credential holds user and group identities to be assumed by a daemon-process.
	Credential *syscall.Credential
	// If Umask is non-zero, the daemon-process call Umask() func with given value.
	Umask int

	// Struct contains only serializable public fields (!!!)
	pidFile *LockFile
}

// Search searches daemons process by given in context pid file name.
// If success returns pointer on daemons os.Process structure,
// else returns error. Returns nil if filename is empty.
func (d *Context) Search() (daemon *os.Process, err error) {
	if len(d.PidFileName) > 0 {
		var pid int
		if _, err = os.Stat(d.PidFileName); err == nil {
			if pid, err = ReadPidFile(d.PidFileName); err != nil {
				return
			}
			Log().Debug().Msgf("Search %s '%s': %v", d.Type, d.PidFileName, pid)
			daemon, err = os.FindProcess(pid)
		} else if errors.Is(err, fs.ErrNotExist) {
			err = nil
		}
	}
	return
}

// Release provides correct pid-file release in daemon.
func (d *Context) Release() (err error) {
	if d.pidFile != nil {
		fd := d.pidFile.Fd()
		Log().Debug().Msgf("Pid `%s` descriptor: %v", d.PidFileName, fd)
		err = d.pidFile.Remove()
	}
	return
}

func (d *Context) Run() (child *os.Process, err error) {
	if err = d.prepareEnv(); err != nil {
		return
	}

	defer d.closeFiles()

	cmd := &exec.Cmd{
		Path:   d.Args[0],
		Args:   d.Args,
		Dir:    d.WorkDir,
		Env:    d.Env,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		SysProcAttr: &syscall.SysProcAttr{
			//Chroot:     d.Chroot,
			Credential: d.Credential,
			Setsid:     true,
		},
	}
	defer cmd.Wait()

	if err = cmd.Start(); err != nil {
		if d.pidFile != nil {
			_ = d.pidFile.Remove()
		}
		return
	}
	return
}

func (d *Context) CreatePidFile() (err error) {
	if len(d.PidFileName) > 0 {
		if d.PidFilePerm == 0 {
			d.PidFilePerm = FilePerm
		}
		if d.PidFileName, err = filepath.Abs(d.PidFileName); err != nil {
			return
		}
		if d.pidFile, err = CreatePidFile(d.PidFileName, d.PidFilePerm); err != nil {
			return
		}
	}
	return
}

func (d *Context) GetStatus() (result bool, err error) {
	var dm *os.Process
	dm, err = d.Search()
	if err != nil {
		Log().Error().Err(err).Msgf("Status %s '%s'", d.Type, d.Name)
	} else if dm != nil {
		err = dm.Signal(syscall.Signal(0))
		if err == os.ErrProcessDone {
			dm = nil
		}
	}

	if dm != nil {
		result = true
	}
	return
}

func (d *Context) closeFiles() (err error) {
	if d.pidFile != nil {
		_ = d.pidFile.Close()
		d.pidFile = nil
	}
	return
}

func (d *Context) prepareEnv() (err error) {
	if len(d.Args) == 0 {
		d.Args = os.Args
	}

	if len(d.Env) == 0 {
		d.Env = os.Environ()
	}
	return
}

func (d *Context) files() (f []*os.File) {
	f = []*os.File{os.Stdin, os.Stdout, os.Stderr}
	if d.pidFile != nil {
		f = append(f, d.pidFile.File)
	}
	return
}
