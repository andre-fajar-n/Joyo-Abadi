package controllers

import (
	"github.com/gofiber/fiber/v2"
)

type IAuth interface {
	RenderLoginPage(c *fiber.Ctx) error
	RenderRegisterPage(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	Register(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
}
