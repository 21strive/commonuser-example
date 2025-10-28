package main

import (
	"github.com/21strive/commonuser"
	"github.com/gofiber/fiber/v2"
)

type HTTPHandler struct {
	commonuser *commonuser.Service
}

func (h *HTTPHandler) Registration(c *fiber.Ctx) error {
	return nil
}

func (h *HTTPHandler) VerifyRegistration(c *fiber.Ctx) error {
	return nil
}

func (h *HTTPHandler) AuthWithEmail(c *fiber.Ctx) error {
	return nil
}

func (h *HTTPHandler) AuthWithUsername(c *fiber.Ctx) error {
	return nil
}

func (h *HTTPHandler) UpdateAccount(c *fiber.Ctx) error {
	return nil
}

func (h *HTTPHandler) Refresh(c *fiber.Ctx) error {
	return nil
}

func (h *HTTPHandler) UpdateEmail(c *fiber.Ctx) error {
	return nil
}

func (h *HTTPHandler) ValidateEmailUpdate(c *fiber.Ctx) error {
	return nil
}

func (h *HTTPHandler) ResendEmailUpdate(c *fiber.Ctx) error {
	return nil
}

func (h *HTTPHandler) RevokeEmailUpdate(c *fiber.Ctx) error {
	return nil
}

func (h *HTTPHandler) UpdatePassword(c *fiber.Ctx) error {
	return nil
}

func (h *HTTPHandler) ForgotPassword(c *fiber.Ctx) error {
	return nil
}

func (h *HTTPHandler) ResetPassword(c *fiber.Ctx) error {
	return nil
}

func (h *HTTPHandler) GetUser(c *fiber.Ctx) error {
	return nil
}

func NewHTTPHandler(commonuser *commonuser.Service) *HTTPHandler {
	return &HTTPHandler{
		commonuser: commonuser,
	}
}
