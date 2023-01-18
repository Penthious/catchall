package web

import (
	"errors"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

const ctxKey = "1"

// Values represent state for each request.
type Values struct {
	TraceID    uuid.UUID
	Log        *zerolog.Logger
	Now        time.Time
	StatusCode int
}

// GetValues returns the values from the context.
func GetValues(ctx echo.Context) (*Values, error) {
	v, ok := ctx.Get(ctxKey).(*Values)
	if !ok {
		return nil, errors.New("web value missing from context")
	}
	return v, nil
}

// GetNow returns start time from the context.
func GetNow(ctx echo.Context) time.Time {
	v, ok := ctx.Get(ctxKey).(*Values)
	if !ok {
		return time.Now()
	}
	return v.Now
}

// GetTraceID returns the trace id from the context.
func GetTraceID(ctx echo.Context) uuid.UUID {
	v, ok := ctx.Get(ctxKey).(*Values)
	if !ok {
		return uuid.New()
	}
	return v.TraceID
}

func GetLogger(ctx echo.Context) (*zerolog.Logger, error) {
	v, ok := ctx.Get(ctxKey).(*Values)
	if !ok {
		log := zerolog.New(os.Stdout)
		return &log, errors.New("web ab value missing from context")
	}

	return v.Log, nil
}

// SetStatusCode sets the status code back into the context.
func SetStatusCode(ctx echo.Context, statusCode int) error {
	v, ok := ctx.Get(ctxKey).(*Values)
	if !ok {
		return errors.New("web status value missing from context")
	}
	v.StatusCode = statusCode
	return nil
}
