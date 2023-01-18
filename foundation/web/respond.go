package web

import (
	"net/http"
	"reflect"

	"github.com/labstack/echo/v4"
)

// Respond converts a Go value to JSON and sends it to the client.
func Respond(ctx echo.Context, statusCode int, data interface{}) error {
	// Set the status code for the request logger middleware.
	SetStatusCode(ctx, statusCode)

	if statusCode == http.StatusNoContent {
		return ctx.NoContent(http.StatusNoContent)
	}

	if statusCode == http.StatusCreated && data == nil {
		return ctx.NoContent(http.StatusCreated)
	}

	if statusCode == http.StatusNotFound {
		return ctx.NoContent(http.StatusNotFound)
	}

	// Check if data is passed in with a nil value: ie `var stuff []thing`, set data to be empty `[]`
	switch reflect.TypeOf(data).Kind() {
	case reflect.Ptr, reflect.Array, reflect.Slice:
		if reflect.ValueOf(data).IsNil() {
			data = []interface{}{}
		}
	case reflect.Map:
		if reflect.ValueOf(data).IsNil() {
			data = make(map[interface{}]interface{})
		}
	}

	// Send the result back to the client.
	if err := ctx.JSON(statusCode, data); err != nil {
		return err
	}

	return nil
}
