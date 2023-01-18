package domain_grp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/penthious/catchall/business/adapters"
	"github.com/penthious/catchall/business/models"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	e := echo.New()
	db := adapters.NewMemoryRepo()
	handler := Handlers{
		DB: db,
	}
	req := httptest.NewRequest(http.MethodGet, "/domain/:domain_name", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	t.Run("catchall", func(t *testing.T) {
		for i := 0; i < 1_000; i++ {
			req := httptest.NewRequest(http.MethodPut, "/events/test/delivered", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			setEchoPath(c, "/events/:domain_name/delivered", "domain_name", "test")

			if err := handler.PutDelivered(c); err != nil {
				t.Fatal(err)
			}
		}

		want := "catch-all"
		setEchoPath(c, "/events/:domain_name/delivered", "domain_name", "test")
		if err := handler.Get(c); err != nil {
			t.Fatal(err)
		}
		assert.Contains(t, rec.Body.String(), want)
	})
	t.Run("not catchall", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/events/test/bounced", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		setEchoPath(c, "/events/:domain_name/bounced", "domain_name", "test")

		if err := handler.PutBounced(c); err != nil {
			t.Fatal(err)
		}

		want := "not catchall"
		setEchoPath(c, "/events/:domain_name/delivered", "domain_name", "test")
		if err := handler.Get(c); err != nil {
			t.Fatal(err)
		}
		assert.NotEqual(t, want, rec.Body.String())
	})
	t.Run("unknown", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/events/blah/bounced", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		setEchoPath(c, "/events/:domain_name/bounced", "domain_name", "blah")

		if err := handler.PutBounced(c); err != nil {
			t.Fatal(err)
		}

		want := "unknown"
		setEchoPath(c, "/events/:domain_name/delivered", "domain_name", "test")
		if err := handler.Get(c); err != nil {
			t.Fatal(err)
		}
		assert.NotEqual(t, want, rec.Body.String())
	})
}

func TestPutBounced(t *testing.T) {
	e := echo.New()
	db := adapters.NewMemoryRepo()
	handler := Handlers{
		DB: db,
	}
	req := httptest.NewRequest(http.MethodPut, "/events/test/bounced", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setEchoPath(c, "/events/:domain_name/bounced", "domain_name", "test")

	if err := handler.PutBounced(c); err != nil {
		t.Fatal(err)
	}

	want := map[string]models.Domain{
		"test": {
			Bounced: 1,
		},
	}

	assert.Equal(t, want, db.Storage)

}

func TestPutDelivered(t *testing.T) {
	e := echo.New()
	db := adapters.NewMemoryRepo()
	handler := Handlers{
		DB: db,
	}
	req := httptest.NewRequest(http.MethodPut, "/events/test/delivered", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setEchoPath(c, "/events/:domain_name/delivered", "domain_name", "test")

	if err := handler.PutDelivered(c); err != nil {
		t.Fatal(err)
	}

	want := map[string]models.Domain{
		"test": {
			Delivered: 1,
		},
	}

	assert.Equal(t, want, db.Storage)
}

// this happens by default in the echo framework, but we need to do it manually for testing
func setEchoPath(c echo.Context, path string, name string, value string) {
	c.SetPath(path)
	c.SetParamNames(name)
	c.SetParamValues(value)
}
