package web

import "strings"

// functions commonly used in templates

func trailingSlash(entry string) string {
	if strings.HasSuffix(entry, "/") {
		return entry
	}
	return entry + "/"
}

func cssShowWhenTrue(show bool) string {
	if show {
		return "show"
	}
	return ""
}

func showErrorToast(title, message string) MessageModel {
	return showToastMessage(title, message, MessageTypeError)
}

func showInfoToast(title, message string) MessageModel {
	return showToastMessage(title, message, MessageTypeInfo)
}

func showSuccessToast(title, message string) MessageModel {
	return showToastMessage(title, message, MessageTypeSuccess)
}

func showToastMessage(title, message, msgType string) MessageModel {
	return MessageModel{
		Title: title,
		Text:  message,
		Type:  msgType,
		Show:  true,
	}
}
