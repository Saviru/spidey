package main

import (
	"log"

	"testapp/api/controllers"
	"testapp/lib/pages"
	"testapp/lib/router"
)

func main() {
	app := router.New()

	// Expose public folder for CSS, JS, etc.
	app.Static("/public/", "./public")

	// Setup API group and register controllers
	api := app.Group("/api")
	controllers.RegisterAPI(api)

	// Auto-register any file-based routes
	pages.RegisterRoutes(app)

	log.Println("🚀 Spidey Tracker running on http://localhost:3000")
	app.Listen("3000")
}
