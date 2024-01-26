package templates

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/a-h/templ"
)

func PageReloadClientJS(jsBlock string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, jsBlock)
		return err
	})
}

func EnsureTrailingSlash(entry string) string {
	if strings.HasSuffix(entry, "/") {
		return entry
	}
	return entry + "/"
}

func Ellipsis(entry string, length int, indicator string) string {
	if entry == "" {
		return ""
	}
	if len(entry) < length {
		return entry
	}
	return entry[:length] + indicator
}

const (
	MsgSuccess = "success"
	MsgError   = "error"
)

type ToastMessage struct {
	Event ToastMessageContent `json:"toastMessage,omitempty"`
}

type ToastMessageContent struct {
	Type  string `json:"type,omitempty"`
	Title string `json:"title,omitempty"`
	Text  string `json:"text,omitempty"`
}

func SuccessToast(title, message string) string {
	return createToastMessage(title, message, MsgSuccess)
}

func ErrorToast(title, message string) string {
	return createToastMessage(title, message, MsgError)
}

func createToastMessage(title, message, msgType string) string {
	toast := ToastMessage{
		Event: ToastMessageContent{
			Type:  msgType,
			Title: title,
			Text:  message,
		},
	}
	return Json(toast)
}

func Json[T any](data T) string {
	payload, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Sprintf("could not marshall data; %v", err))
	}
	return string(payload)
}
