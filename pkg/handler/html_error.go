package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"
	"time"

	"golang.binggl.net/monorepo/pkg/cookies"
)

// FlashKeyError is a key to retrieve an error message
const FlashKeyError = "error.flash"

// FlashKeyInfo is a key to retrieve an info message
const FlashKeyInfo = "info.flash"

const errorTemplate = "error.tmpl"

// TemplateHandler is used to display errors using HTML templates
type TemplateHandler struct {
	Handler
	CookieSettings cookies.Settings
	Version        string
	Build          string
	AppName        string
	TemplateDir    string
}

// HandleError returns a HTML template showing errors
func (h *TemplateHandler) HandleError(w http.ResponseWriter, r *http.Request) error {
	cookie := cookies.NewAppCookie(h.CookieSettings)
	var (
		msg       string
		err       string
		isError   bool
		isMessage bool
		init      sync.Once
		tmpl      *template.Template
		e         error
	)

	init.Do(func() {
		path := filepath.Join(h.TemplateDir, errorTemplate)
		tmpl, e = template.ParseFiles(path)
	})
	if e != nil {
		return e
	}

	// read (flash)
	err = cookie.Get(FlashKeyError, r)
	if err != "" {
		isError = true
	}
	msg = cookie.Get(FlashKeyInfo, r)
	if msg != "" {
		isMessage = true
	}

	// clear (flash)
	cookie.Del(FlashKeyError, w)
	cookie.Del(FlashKeyInfo, w)

	appName := "binggl.net"
	if h.AppName != "" {
		appName = h.AppName
	}

	data := map[string]interface{}{
		"year":      time.Now().Year(),
		"appname":   appName,
		"version":   fmt.Sprintf("%s-%s", h.Version, h.Build),
		"isError":   isError,
		"error":     err,
		"isMessage": isMessage,
		"msg":       msg,
	}

	return tmpl.Execute(w, data)
}
