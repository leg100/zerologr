package main

import (
	"errors"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/zerologr"
	"github.com/rs/zerolog"
)

func main() {
	// Setup logger
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	logger := zerolog.New(consoleWriter).Level(zerolog.DebugLevel).With().Timestamp().Logger()

	log := zerologr.NewLogger(&logger)

	example(log.WithValues("module", "example"))
}

// If this were in another package, all it would depend on in logr, not zapr.
func example(log logr.Logger) {
	log.Info("hello", "val1", 1, "val2", map[string]int{"k": 1})
	log.V(1).Info("you should see this")
	log.V(1).V(1).Info("you should NOT see this")
	log.Error(nil, "uh oh", "trouble", true, "reasons", []float64{0.1, 0.11, 3.14})
	log.Error(errors.New("an error occurred"), "goodbye", "code", -1)
}
