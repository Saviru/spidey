package main

import (
	"fmt"
	"log"

	"testapp/lib/pages"
	"testapp/lib/router"
)

func main() {
	app := router.New()
	pages.RegisterRoutes(app)

	app.Static("/assets/", "public/assets")

	fmt.Println("Server running on http://localhost:3000")
	log.Fatal(app.Listen("3000"))
}
