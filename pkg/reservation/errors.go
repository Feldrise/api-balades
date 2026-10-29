package reservation

import (
	"feldrise.com/balade/pkg/errors"
	"github.com/go-chi/render"
)

func renderError(err error) render.Renderer {
	return errors.ErrServerError(err)
}
