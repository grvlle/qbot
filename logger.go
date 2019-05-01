package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitializeLogger initializes zero logger with pretty STDERR output
// TODO: Include better db.logger output
func InitializeLogger() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	logger := log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	log.Logger = logger
}

//db.SetLogger(gorm.Logger{revel.TRACE}) db.SetLogger(log.New(os.Stdout, "\r\n", 0))
