package middleware

import (
	"github.com/penthious/catchall/foundation/web"
	webErr "github.com/penthious/catchall/foundation/web/errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Errors handles errors coming out of the call chain. It detects normal
// application errors which are used to respond to the client in a uniform way.
// Unexpected errors (status >= 500) are logged.
func Errors() echo.MiddlewareFunc {

	// This is the actual middleware function to be executed.
	m := func(handler echo.HandlerFunc) echo.HandlerFunc {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx echo.Context) error {

			v, err := web.GetValues(ctx)
			if err != nil {
				return webErr.NewShutdownError("web value missing from context")
			}

			// Run the next handler and catch any propagated errors.
			if err := handler(ctx); err != nil {
				// Build out the errors response.
				var er webErr.ErrorResponse
				var status int
				switch {
				case webErr.IsRequestError(err):
					reqErr := webErr.GetRequestError(err)
					er = webErr.ErrorResponse{
						Error: reqErr.Error(),
					}
					status = reqErr.Status

				default:
					er = webErr.ErrorResponse{
						Error: http.StatusText(http.StatusInternalServerError),
					}
					status = http.StatusInternalServerError
				}

				// Log out the errors
				logger, _ := web.GetLogger(ctx)
				logger.
					Error().
					Interface("uuid", v.TraceID).
					Str("httpMethod", ctx.Request().Method).
					Str("route", ctx.Request().URL.Path).
					Interface("er", er).
					Err(err).
					Msg(err.Error())

				// Respond with the errors back to the client.
				if err := web.Respond(ctx, status, er); err != nil {
					return err
				}

				// If we receive the shutdown err we need to return it
				// back to the base handler to shut down the service.
				if ok := webErr.IsShutdown(err); ok {
					return err
				}
			}

			// The errors has been handled, so we can stop propagating it.
			return nil
		}

		return h
	}

	return m
}
