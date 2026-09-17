package main

import (
	"os"

	"backend/app"
)

func main() {

	server := app.CreateApp()

	port := os.Getenv("APP_PORT")

	if port == "" {
		port = "5000"
	}

	if err := server.Listen(":" + port); err != nil {
		panic(err)
	}
}