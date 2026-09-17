package handler

import (
	"net/http"
	"sync"

	"backend/app"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

var (
	server     *fiber.App
	serverErr  error
	serverOnce sync.Once
)

func Handler(w http.ResponseWriter, r *http.Request) {
	serverOnce.Do(func() {
		server, serverErr = app.CreateApp()
	})
	if serverErr != nil {
		http.Error(w, "Application startup failed: "+serverErr.Error(), http.StatusServiceUnavailable)
		return
	}

	adaptor.FiberApp(server)(w, r)
}
