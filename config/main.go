package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
	"os"
)

func Cfg() *Config {
	return application
}

func Log() *zerolog.Logger {
	if logger == nil {
		GetLogger()
	}
	return logger
}

func SetConfig(cfg interface{}) *Config {
	err := mapstructure.Decode(cfg, &application)
	if err != nil {
		Log().Info().Msg("Worker load config")
		return nil
	}
	return application
}

func SetLogger(log *zerolog.Logger) *zerolog.Logger {
	logger = log
	return logger
}

func GetLogger() *zerolog.Logger {
	logLevel := zerolog.InfoLevel
	if application.Debug {
		logLevel = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(logLevel)
	zerolog.TimestampFieldName = "timestamp"
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()
	return SetLogger(&log)
}
