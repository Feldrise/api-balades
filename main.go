package main

import (
	"log"
	"net/http"

	"feldrise.com/balade/config"
	_ "feldrise.com/balade/docs"
	"feldrise.com/balade/pkg/router"
	"feldrise.com/balade/pkg/scheduler"
	"github.com/go-chi/chi"
	httpSwagger "github.com/swaggo/http-swagger"
)

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
		configuration.Clock,
		configuration.ApplicationURL,
	)
	registrationScheduler.StartScheduler()

	// We initialize the routes
	apiRouter := router.New(configuration)

	// Swagger configs
	apiRouter.Get("/swagger/*", httpSwagger.WrapHandler)

	// We show all the routes in the logs
	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("%s %s\n", method, route)
		return nil
	}
	if err := chi.Walk(apiRouter, walkFunc); err != nil {
		log.Panicf("Logging error: %s\n", err.Error())
	}

	// We serve the api
	log.Printf("connect to http://localhost:%s/swagger/index.html for documentation", configuration.Constants.Port)
	log.Fatal(http.ListenAndServe(":"+configuration.Constants.Port, apiRouter))
}
