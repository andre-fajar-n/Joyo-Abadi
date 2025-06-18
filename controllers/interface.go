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

type IProduct interface {
	ListProducts(c *fiber.Ctx) error
	RenderCreateProduct(c *fiber.Ctx) error
	CreateProduct(c *fiber.Ctx) error
	RenderEditProduct(c *fiber.Ctx) error
	UpdateProduct(c *fiber.Ctx) error
	DeleteProduct(c *fiber.Ctx) error
	GetProductDetail(c *fiber.Ctx) error
}

type IHome interface {
	Home(c *fiber.Ctx) error
}
