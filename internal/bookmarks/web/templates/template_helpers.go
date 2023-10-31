package templates

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"golang.binggl.net/monorepo/pkg/develop"
)

func PageReloadClientJS() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, develop.PageReloadClientJS)
		return err
	})
}
