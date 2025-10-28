package main

import (
	"database/sql"
	"github.com/21strive/commonuser"
	"github.com/21strive/commonuser/account"
	"github.com/gofiber/fiber/v2"
)

type HTTPHandler struct {
	writeDB    *sql.DB
	commonuser *commonuser.Service
}

func (h *HTTPHandler) Registration(c *fiber.Ctx) error {
	var requestBody NativeRegister
	if err := c.BodyParser(&requestBody); err != nil {
		return ReturnErrorResponse(c, fiber.StatusBadRequest, err, "invalid-request-body")
	}

	tx, errInitTx := h.writeDB.Begin()
	if errInitTx != nil {
		return errInitTx
	}
	defer tx.Rollback()

	newAccount := account.New()
	newAccount.Name = requestBody.Name
	newAccount.Username = requestBody.Username
	newAccount.Email = requestBody.Email

	verification, regError := h.commonuser.Register(tx, newAccount, true)
	if regError != nil {
		return regError
	}

	errCommit := tx.Commit()
	if errCommit != nil {
		return errCommit
	}

	return c.JSON(map[string]string{"verificationCode": verification.Code})
}

func (h *HTTPHandler) VerifyRegistration(c *fiber.Ctx) error {
	var requestBody VerifyRegistration
	if err := c.BodyParser(&requestBody); err != nil {
		return ReturnErrorResponse(c, fiber.StatusBadRequest, err, "invalid-request-body")
	}

	tx, errInitTx := h.writeDB.Begin()
	if errInitTx != nil {
		return errInitTx
	}
	defer tx.Rollback()

	isValid, errVerify := h.commonuser.Verification().Verify(tx, requestBody.AccountUUID, requestBody.VerificationCode)
	if errVerify != nil {
		return errVerify
	}
	if !isValid {
		return ReturnErrorResponse(c, fiber.StatusBadRequest, nil, "unauthorized")
	}

	return c.SendStatus(fiber.StatusOK)
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
