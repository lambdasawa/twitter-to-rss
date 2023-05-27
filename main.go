package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/guregu/dynamo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Payload struct {
	Text      string `json:"text"`
	Link      string `json:"link"`
	CreatedAt string `json:"createdAt"`
}

func main() {
	db := newDB()

	e := echo.New()

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}

		c.Logger().Error(err)

		errorPage := fmt.Sprintf("%d.html", code)
		if err := c.File(errorPage); err != nil {
			c.Logger().Error(err)
		}
	}

	e.Use(middleware.Logger())

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	e.POST("/:id", func(c echo.Context) error {
		ctx := c.Request().Context()

		id := c.Param("id")

		var payload Payload

		if err := c.Bind(&payload); err != nil {
			return fmt.Errorf("bind: %w", err)
		}

		feed, err := load(ctx, db, id)
		if err != nil && !errors.Is(err, dynamo.ErrNotFound) {
			return fmt.Errorf("load: %w", err)
		}

		if feed == nil {
			feed = NewFeed(id, payload)
		}

		feed.PrependItem(id, payload)

		if err := save(ctx, db, feed); err != nil {
			return fmt.Errorf("save: %w", err)
		}

		//5c.Response().Header().Set("Content-Type", "application/atom+xml; charset=utf-8")

		return c.NoContent(http.StatusCreated)
	})

	e.GET("/:id", func(c echo.Context) error {
		ctx := c.Request().Context()

		id := c.Param("id")

		feed, err := load((ctx), db, id)
		if err != nil && errors.Is(err, dynamo.ErrNotFound) {
			return fmt.Errorf("load: %w", err)
		}

		xml, err := feed.Encode()
		if err != nil {
			return fmt.Errorf("encode: %w", err)
		}

		c.Response().Header().Set("Content-Type", "application/atom+xml; charset=utf-8")

		return c.String(http.StatusOK, xml)
	})

	e.Logger.Fatal(e.Start(":8000"))
}
