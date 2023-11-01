package templates

import (
	"context"
	"io"
	"strings"

	"github.com/a-h/templ"
	"golang.binggl.net/monorepo/pkg/develop"
)

func PageReloadClientJS() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, develop.PageReloadClientJS)
		return err
	})
}

func EnsureTrailingSlash(entry string) string {
	if strings.HasSuffix(entry, "/") {
		return entry
	}
	return entry + "/"
}
