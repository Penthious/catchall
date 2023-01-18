package middleware

import (
	"github.com/penthious/catchall/foundation/web"
	webErr "github.com/penthious/catchall/foundation/web/errors"
	"time"

	"github.com/labstack/echo/v4"
)

func LogRequest() echo.MiddlewareFunc {
	m := func(handler echo.HandlerFunc) echo.HandlerFunc {
		h := func(ctx echo.Context) error {
			// Before the request is made
			v, err := web.GetValues(ctx)
			if err != nil {
				return webErr.NewShutdownError("web value missing from context")
			}
			routePath := ctx.Request().URL.Path
			v.Log.Info().
				Str("method", ctx.Request().Method).
				Str("route", routePath).
				Interface("uuid", v.TraceID).
				Msg("request started")

			// Run the next handler and catch any propagated errors.
			err = handler(ctx)

			// After the request is made
			dur := time.Since(v.Now).Microseconds()

			v.Log.Info().
				Int64("duration", dur).
				Str("method", ctx.Request().Method).
				Str("route", routePath).
				Interface("uuid", v.TraceID).
				Msg("request finished")
			return err
		}
		return h
	}
	return m
}
