package handler

import (
    "net/http"

    "backend/app"
    "github.com/gofiber/fiber/v2/middleware/adaptor"
)

var server = app.CreateApp()

func Handler(w http.ResponseWriter, r *http.Request) {
    adaptor.FiberApp(server)(w, r)
}