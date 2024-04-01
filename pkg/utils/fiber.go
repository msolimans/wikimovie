package utils

import "github.com/gofiber/fiber/v2"

type IHandler interface {
	Routes(app *fiber.App, group fiber.Router)
}

func NewRouter(prefix string, app *fiber.App, handler IHandler) {
	group := app.Group(prefix)
	handler.Routes(app, group)
}
