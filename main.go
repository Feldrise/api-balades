package main

import (
	"log"
	"net/http"

	"feldrise.com/balade/config"
	_ "feldrise.com/balade/docs"
	"feldrise.com/balade/pkg/authentication"
	"feldrise.com/balade/pkg/guide"
	"feldrise.com/balade/pkg/payment"
	"feldrise.com/balade/pkg/ramble"
	"feldrise.com/balade/pkg/registration"
	"feldrise.com/balade/pkg/scheduler"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

func Routes(configuration *config.Config) *chi.Mux {
	router := chi.NewRouter()

	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.Recoverer,
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
		r.Mount("/registrations", registration.New(configuration).Routes())
		r.Mount("/payments", payment.Routes(configuration.PaymentService))
	})

	fs := http.FileServer(http.Dir(configuration.Constants.DataPath + "/uploads"))
	router.Handle("/uploads/*", http.StripPrefix("/uploads/", fs))

	return router
}

// @title balade API
// @version 1.0
// @description This is the API for balade, a diagnostic for enterpreneurs.
// @contact contact@feldrise.com
// @host api.balade.feldrise.com
// @BasePath /api/v1
func main() {
	// We initialize the project
	configuration, err := config.New()
	if err != nil {
		log.Panicln("Error initializing configuration:", err)
	}

	// Start the registration scheduler
	registrationScheduler := scheduler.NewRegistrationScheduler(
		configuration.RambleRepository,
		configuration.RambleRegistrationRepository,
		configuration.RambleRegistrationGroupRepository,
		configuration.EmailService,
		configuration.ApplicationURL,
	)
	registrationScheduler.StartScheduler()

	// We initialize the routes
	router := Routes(configuration)

	// Swagger configs
	router.Get("/swagger/*", httpSwagger.WrapHandler)

	// We show all the routes in the logs
	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("%s %s\n", method, route)
		return nil
	}
	if err := chi.Walk(router, walkFunc); err != nil {
		log.Panicf("Logging error: %s\n", err.Error())
	}

	// We serve the api
	log.Printf("connect to http://localhost:%s/swagger/index.html for documentation", configuration.Constants.Port)
	log.Fatal(http.ListenAndServe(":"+configuration.Constants.Port, router))
}
