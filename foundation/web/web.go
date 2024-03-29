// Package web contains a small web framework extension.
package web

import (
	"errors"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// App is the entrypoint into our application and what configures our context
// object for each of our http handlers. Feel free to add any configuration
// data/logic on this App struct.
type App struct {
	Mux      *echo.Echo
	shutdown chan os.Signal
	mw       []echo.MiddlewareFunc
	log      *zerolog.Logger
}

// NewApp creates an App value that handle a set of routes for the application.
func NewApp(serviceName string, shutdown chan os.Signal, log *zerolog.Logger, mw ...echo.MiddlewareFunc) *App {
	e := echo.New()

	return &App{
		Mux:      e,
		shutdown: shutdown,
		mw:       mw,
		log:      log,
	}
}

// SignalShutdown is used to gracefully shut down the app when an integrity
// issue is identified.
func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// ServeHTTP implements the http.Handler interface. It's the entry point for
// all http traffic and allows the opentelemetry mux to run first to handle
// tracing. The opentelemetry mux then calls the application mux to handle
// application traffic. This was set up on line 44 in the NewApp function.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.Mux.ServeHTTP(w, r)
}

// Handle sets a handler function for a given HTTP method and path pair
// to the application server mux.
func (a *App) Handle(method string, group string, path string, handler echo.HandlerFunc, mw ...echo.MiddlewareFunc) {

	// First wrap handler specific middleware around this handler.
	handler = wrapMiddleware(mw, handler)

	// Add the application's general middleware to the handler chain.
	handler = wrapMiddleware(a.mw, handler)

	// The function to execute for each request.
	h := func(ctx echo.Context) error {

		ctx.Set(ctxKey, &Values{Now: time.Now().UTC(), TraceID: uuid.New(), Log: a.log})

		// Call the wrapped handler functions.
		if err := handler(ctx); err != nil {

			// Since there was an errors, validate the condition of this
			// errors and determine if we need to actually shutdown or not.
			if validateShutdown(err) {
				a.SignalShutdown()
				return nil
			}
		}
		return nil
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}
	switch method {
	case http.MethodPatch:
		a.Mux.PATCH(finalPath, h)
	case http.MethodGet:
		a.Mux.GET(finalPath, h)
	case http.MethodPost:
		a.Mux.POST(finalPath, h)
	case http.MethodOptions:
		a.Mux.OPTIONS(finalPath, h)
	case http.MethodPut:
		a.Mux.PUT(finalPath, h)
	case http.MethodConnect:
		a.Mux.CONNECT(finalPath, h)
	case http.MethodTrace:
		a.Mux.TRACE(finalPath, h)
	case http.MethodHead:
		a.Mux.HEAD(finalPath, h)
	case http.MethodDelete:
		a.Mux.DELETE(finalPath, h)

	}
	// a.mux.HandleFunc(finalPath, h).Methods(method)
}

// validateShutdown validates the errors for special conditions that do not
// warrant an actual shutdown by the system.
func validateShutdown(err error) bool {

	// Ignore syscall.EPIPE and syscall.ECONNRESET errors which occurs
	// when a write operation happens on the http.ResponseWriter that
	// has simultaneously been disconnected by the client (TCP
	// connections is broken). For instance, when large amounts of
	// data are being written or streamed to the client.
	// https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
	// https://gosamples.dev/broken-pipe/
	// https://gosamples.dev/connection-reset-by-peer/

	switch {
	case errors.Is(err, syscall.EPIPE):

		// Usually, you get the broken pipe errors when you write to the connection after the
		// RST (TCP RST Flag) is sent.
		// The broken pipe is a TCP/IP errors occurring when you write to a stream where the
		// other end (the peer) has closed the underlying connection. The first write to the
		// closed connection causes the peer to reply with an RST packet indicating that the
		// connection should be terminated immediately. The second write to the socket that
		// has already received the RST causes the broken pipe errors.
		return false

	case errors.Is(err, syscall.ECONNRESET):

		// Usually, you get connection reset by peer errors when you read from the
		// connection after the RST (TCP RST Flag) is sent.
		// The connection reset by peer is a TCP/IP errors that occurs when the other end (peer)
		// has unexpectedly closed the connection. It happens when you send a packet from your
		// end, but the other end crashes and forcibly closes the connection with the RST
		// packet instead of the TCP FIN, which is used to close a connection under normal
		// circumstances.
		return false
	}

	return true
}
