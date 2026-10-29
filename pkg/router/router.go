package router

import (
	"net/http"

	"feldrise.com/balade/config"
	"feldrise.com/balade/pkg/authentication"
	"feldrise.com/balade/pkg/guide"
	apimiddleware "feldrise.com/balade/pkg/middleware"
	"feldrise.com/balade/pkg/payment"
	"feldrise.com/balade/pkg/ramble"
	"feldrise.com/balade/pkg/reservation"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/rs/cors"
)

// New builds the complete HTTP router of the API.
func New(configuration *config.Config) *chi.Mux {
	router := chi.NewRouter()

	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.Recoverer,
		apimiddleware.MaxBodySize(apimiddleware.DefaultMaxBodySize),
		authentication.Middelware(configuration),
	)

	// We initialize CORS
	router.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
	}).Handler)

	router.Route("/api/v1", func(r chi.Router) {
		r.Mount("/authentication", authentication.New(configuration).Routes())
		r.Mount("/guides", guide.New(configuration).Routes())
		r.Mount("/rambles", ramble.New(configuration).Routes())
		r.Mount("/reservations", reservation.New(configuration).Routes())
		r.Mount("/payments", payment.Routes(configuration))
	})

	fs := http.FileServer(http.Dir(configuration.Constants.DataPath + "/uploads"))
	router.Handle("/uploads/*", http.StripPrefix("/uploads/", fs))

	return router
}
