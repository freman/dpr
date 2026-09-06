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
		r.echo.Logger.Warn("rejecting /events request: no Cookie header present")
		return "", echo.ErrBadRequest
	}

	rcReq.Header.Add("Cookie", cookie)

	rcResp, err := (&http.Client{Timeout: 30 * time.Second}).Do(rcReq)
	if err != nil {
		return "", err
	}

	defer rcResp.Body.Close()

	// Roundcube spits out "not json" (usually a redirect to login, meaning the
	// session tied to this cookie is no longer valid) if the user isn't logged in
	contentType := rcResp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "json") {
		r.echo.Logger.Warnf(
			"roundcube session check failed: status=%d content-type=%q (session likely expired or invalid)",
			rcResp.StatusCode, contentType,
		)
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
		// echo.ErrForbidden/echo.ErrBadRequest are *echo.HTTPError values returned
		// by verifySessionWithRoundcube above; pass those straight through so the
		// client sees the real 403/400 instead of a masking 500.
		var httpErr *echo.HTTPError
		if errors.As(err, &httpErr) {
			return httpErr
		}

		r.echo.Logger.Error(err)

		return echo.ErrInternalServerError
	}

	r.echo.Logger.Infof("subscribing %q to push events", username)

	any, _ := r.sseMap.LoadOrStore(username, newLazySSE())

	s := any.(*lazySSE).es()

	// Send this request to the SSE library for the remainder of it's life
	s.ServeHTTP(c.Response(), c.Request())

	// When the client disconnects the SSE library returns, if there are no
	// more clients we can shut it down
	if s.Len() == 0 {
		r.echo.Logger.Infof("unsubscribing %q: no more listeners, shutting down stream", username)
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
		r.echo.Logger.Infof("discarding %q notification for %q: no active /events subscriber", body.Event, body.User)
		return nil
	}

	any.(*lazySSE).es().Emit(body)

	return nil
}
