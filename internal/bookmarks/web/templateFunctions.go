package web

import (
	"strings"

	"golang.binggl.net/monorepo/internal/bookmarks/app/bookmarks"
)

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

func valWhenNodeEq(expec string, val bookmarks.NodeType, returnVal string) string {
	if string(val) == expec {
		return returnVal
	}
	return ""
}

func valWhenTrue(expec bool, returnVal string) string {
	if expec {
		return returnVal
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
