package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type rcox struct {
	cfg config

	echo   *echo.Echo
	sseMap sync.Map
}

func (r *rcox) verifySessionWithRoundcube(req *http.Request) (string, error) {
	rcReq, err := http.NewRequestWithContext(
		req.Context(),
		http.MethodGet,
		strings.TrimSuffix(r.cfg.roundcube, "/")+"/?_task=mail&_mbox=INBOX&_action=plugin.sse",
		nil,
	)

	if err != nil {
		return "", err
	}

	// Borrow the cookie provided by the request to subscribe to the event bus
	cookie := req.Header.Get("Cookie")
	if cookie == "" {
		return "", echo.ErrBadRequest
	}

	rcReq.Header.Add("Cookie", cookie)

	rcResp, err := (&http.Client{Timeout: 30 * time.Second}).Do(rcReq)
	if err != nil {
		return "", err
	}

	defer rcResp.Body.Close()

	// Roundcube spits out "not json" (usually a redirect) if the user isn't logged in
	if !strings.Contains(rcResp.Header.Get("Content-Type"), "json") {
		return "", echo.ErrForbidden
	}

	var rcJSON rcResponse
	if err := json.NewDecoder(rcResp.Body).Decode(&rcJSON); err != nil {
		return "", err
	}

	return rcJSON.Username, nil
}

func (r *rcox) handleGetEvents(c echo.Context) error {
	username, err := r.verifySessionWithRoundcube(c.Request())
	if err != nil {
		if errors.Is(err, &echo.HTTPError{}) {
			return err
		}
		r.echo.Logger.Error(err)

		return echo.ErrInternalServerError
	}

	any, _ := r.sseMap.LoadOrStore(username, newLazySSE())

	s := any.(*lazySSE).es()

	// Send this request to the SSE library for the remainder of it's life
	s.ServeHTTP(c.Response(), c.Request())

	// When the client disconnects the SSE library returns, if there are no
	// more clients we can shut it down
	if s.Len() == 0 {
		r.sseMap.Delete(username)
		s.Shutdown()
	}

	return nil
}

func (r *rcox) handleNotification(c echo.Context) error {
	var body oxNotification
	if err := c.Bind(&body); err != nil {
		return err
	}

	// Discard the request
	any, found := r.sseMap.Load(body.User)
	if !found {
		return nil
	}

	any.(*lazySSE).es().Emit(body)

	return nil
}
