package main

import (
	"net"

	"github.com/labstack/echo/v4"
)

type networks []*net.IPNet

func (n *networks) Contains(ip net.IP) bool {
	for _, r := range *n {
		if r.Contains(ip) {
			return true
		}
	}
	return false
}

func (n *networks) middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		ip := net.ParseIP(c.RealIP())
		if n.Contains(ip) {
			return next(c)
		}

		return echo.ErrForbidden
	}
}
