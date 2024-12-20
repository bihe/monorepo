package common

import (
	"encoding/json"
	"fmt"
)

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
