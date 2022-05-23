package main

import (
	"crypto/subtle"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.IPExtractor = echo.ExtractIPFromXFFHeader()
	e.Use(middleware.Logger())

	app := &rcox{
		cfg:  configure(),
		echo: e,
	}

	authMiddleware := middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		return subtle.ConstantTimeCompare([]byte(username), []byte(app.cfg.username)) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte(app.cfg.password)) == 1, nil
	})

	e.GET("/events", app.handleGetEvents)
	e.PUT("/preliminary/http-notify/v1/notify",
		app.cfg.whitelist.middleware(
			authMiddleware(app.handleNotification),
		),
	)

	e.Start(app.cfg.listen)
}
