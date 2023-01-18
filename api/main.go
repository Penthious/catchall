package main

import (
	"context"
	"fmt"
	"github.com/penthious/catchall/api/handlers"
	"github.com/penthious/catchall/business/adapters"
	"github.com/penthious/catchall/business/ports"
	"github.com/penthious/catchall/foundation/database"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"go.uber.org/automaxprocs/maxprocs"
)

// Entry point to the server
func main() {
	appName := "catchall"
	log := initLogger(loggerConf{})

	log.Info().Msg("Application starting")

	if err := server(appName, log); err != nil {
		log.Error().Err(err).Msg("startup")
	}
}

// server creates the http.Server and runs it
func server(appName string, log *zerolog.Logger) error {
	// Get maxprocs
	opt := maxprocs.Logger(log.Printf)

	// Set the correct number of threads for the service
	// based on what is available either by the machine or quotas.
	if _, err := maxprocs.Set(opt); err != nil {
		return fmt.Errorf("maxprocs: %w", err)
	}

	log.Info().Interface("GOMAXPROCS", runtime.GOMAXPROCS(0)).Msg("startup")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// @TODO: I would use viper or something similar to manage config
	// change this value to `memory` to use the in-memory database.
	adapter := "postgres"

	// used to stop the execution of connecting to the databases if the time limit is reached
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var db ports.DB
	switch adapter {
	case "postgres":
		psql, err := database.Open(database.Config{
			User:         "postgres",
			Password:     "example",
			Host:         "localhost:5434",
			Name:         "postgres",
			DisableTLS:   true,
			MaxOpenConns: 50,
			MaxIdleConns: 2,
			MaxIdleTime:  time.Minute,
		})
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}

		if err := database.StatusCheck(ctx, psql); err != nil {
			return fmt.Errorf("database not ready: %w", err)
		}

		db = adapters.NewPostgresRepo(psql)
	case "mongo":
		// this is where I would add mongo or any other database
	case "memory":
		db = adapters.NewMemoryRepo()
	default:
		return fmt.Errorf("unknown adapter: %s", adapter)
	}

	apiMux := handlers.APIMux(handlers.APIMuxConfig{
		Log:         log,
		ServiceName: appName,
		Shutdown:    shutdown,
		DB:          db,
	})

	// TODO: We should create a background service hosted on a different port to expose debug endpoints
	// Construct a server to service the requests against the mux.
	api := http.Server{
		Addr:         "localhost:7000",
		Handler:      apiMux,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 10,
		IdleTimeout:  time.Minute * 2,
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this errors.
	serverErrors := make(chan error, 1)

	// Start the service listening for api requests.
	go func() {
		log.Info().
			Str("host", api.Addr).
			Msg("Application starting")
		serverErrors <- api.ListenAndServe()
	}()

	// =========================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	if err := waitForSignalShutdown(shutdown, serverErrors, &api, log); err != nil {
		log.Error().
			Err(err).
			Msg("server shutdown")

		return err
	}

	return nil
}

type loggerConf struct {
	App     string
	Build   string
	Now     string
	Version string
	Env     string
	Debug   bool
}

// initLogger initializes the logger for the application.
func initLogger(c loggerConf) *zerolog.Logger {
	log := zerolog.New(os.Stdout).
		With().
		Dict("ctx",
			zerolog.Dict().
				Str("APP", c.App).
				Str("BUILD", c.Build).
				Str("VERSION", c.Version).
				Str("TIME", c.Now),
		).
		Stack().
		Timestamp().
		Logger()

	return &log
}

// waitForSignalShutdown *** THIS IS A BLOCKING CALL *** run a select statement that listens for either server errors
// or shutdown signals, it'll terminate the running http.Server
func waitForSignalShutdown(shutdown chan os.Signal, serverErrors chan error, api *http.Server, logger *zerolog.Logger) error {

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server errors: %w", err)

	case sig := <-shutdown:
		logger.Info().Interface("signal", sig).Msg("shutdown started")
		defer logger.Info().Interface("signal", sig).Msg("shutdown complete")

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		defer cancel()

		// Asking listener to shut down and shed load.
		if err := api.Shutdown(ctx); err != nil {
			_ = api.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
