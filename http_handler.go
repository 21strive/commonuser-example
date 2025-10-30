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
		return ErrorResponse(c, fiber.StatusBadRequest, err, "invalid-request-body")
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
	account := c.Locals("account").(*account.Account)

	var requestBody VerifyRegistration
	if err := c.BodyParser(&requestBody); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, err, "invalid-request-body")
	}

	tx, errInitTx := h.writeDB.Begin()
	if errInitTx != nil {
		return errInitTx
	}
	defer tx.Rollback()

	isValid, errVerify := h.commonuser.Verification().Verify(tx, account.GetUUID(), requestBody.VerificationCode)
	if errVerify != nil {
		return errVerify
	}
	if !isValid {
		return ErrorResponse(c, fiber.StatusBadRequest, nil, "unauthorized")
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *HTTPHandler) AuthWithEmail(c *fiber.Ctx) error {
	var requestBody LoginWithEmail
	if err := c.BodyParser(&requestBody); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, err, "invalid-request-body")
	}

	tx, errInitTx := h.writeDB.Begin()
	if errInitTx != nil {
		return errInitTx
	}
	defer tx.Rollback()

	deviceInfo := commonuser.DeviceInfo{
		DeviceId:   requestBody.DeviceId,
		DeviceType: requestBody.DeviceType,
		UserAgent:  requestBody.UserAgent,
	}

	accessToken, refreshToken, errToken := h.commonuser.AuthenticateByEmail(
		tx,
		requestBody.Email,
		requestBody.Password,
		deviceInfo,
	)
	if errToken != nil {
		return errToken
	}

	errCommit := tx.Commit()
	if errCommit != nil {
		return errCommit
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})

	return c.JSON(map[string]string{"accessToken": accessToken})
}

func (h *HTTPHandler) AuthWithUsername(c *fiber.Ctx) error {
	var requestBody LoginWithUsername
	if err := c.BodyParser(&requestBody); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, err, "invalid-request-body")
	}

	tx, errInitTx := h.writeDB.Begin()
	if errInitTx != nil {
		return errInitTx
	}
	defer tx.Rollback()

	deviceInfo := commonuser.DeviceInfo{
		DeviceId:   requestBody.DeviceId,
		DeviceType: requestBody.DeviceType,
		UserAgent:  requestBody.UserAgent,
	}

	accessToken, refreshToken, errToken := h.commonuser.AuthenticateByUsername(
		tx,
		requestBody.Username,
		requestBody.Password,
		deviceInfo,
	)
	if errToken != nil {
		return errToken
	}

	errCommit := tx.Commit()
	if errCommit != nil {
		return errCommit
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})

	return c.JSON(map[string]string{"accessToken": accessToken})
}

func (h *HTTPHandler) UpdateAccount(c *fiber.Ctx) error {
	account := c.Locals("account").(*account.Account)

	var requestBody UpdateAccount
	if err := c.BodyParser(&requestBody); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, err, "invalid-request-body")
	}

	tx, errInitTx := h.writeDB.Begin()
	if errInitTx != nil {
		return errInitTx
	}
	defer tx.Rollback()

	updateOpt := commonuser.UpdateOpt{
		NewName:     requestBody.Name,
		NewUsername: requestBody.Username,
		NewAvatar:   requestBody.Avatar,
	}

	err := h.commonuser.Update(tx, account.GetUUID(), updateOpt)
	if err != nil {
		return err
	}

	errCommit := tx.Commit()
	if errCommit != nil {
		return errCommit
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *HTTPHandler) Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refreshToken")
	if refreshToken == "" {
		return ErrorResponse(c, fiber.StatusUnauthorized, nil, "missing-refresh-token")
	}

	tx, errInitTx := h.writeDB.Begin()
	if errInitTx != nil {
		return errInitTx
	}
	defer tx.Rollback()

	// For refresh, we need to find the account first, but the method signature suggests
	// we need the account object. This might require additional implementation
	// depending on how the refresh token is linked to the account in your system

	// This is a placeholder - you'll need to implement account lookup from refresh token
	return ErrorResponse(c, fiber.StatusNotImplemented, nil, "not-implemented")
}

func (h *HTTPHandler) UpdateEmail(c *fiber.Ctx) error {
	account := c.Locals("account").(*account.Account)

	var requestBody UpdateEmail
	if err := c.BodyParser(&requestBody); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, err, "invalid-request-body")
	}

	tx, errInitTx := h.writeDB.Begin()
	if errInitTx != nil {
		return errInitTx
	}
	defer tx.Rollback()

	userAccount, err := h.commonuser.Find().ByUUID(account.GetUUID())
	if err != nil {
		return err
	}

	updateEmail, err := h.commonuser.EmailUpdate().RequestUpdate(tx, userAccount, requestBody.NewEmail)
	if err != nil {
		return err
	}

	errCommit := tx.Commit()
	if errCommit != nil {
		return errCommit
	}

	// NOTE: The returned token should be sent to the user's NEW email address
	// for verification. You should implement email delivery (SMTP, service provider, etc.)
	// to send this token to requestBody.NewEmail. The user must use this token
	// in the ValidateEmailUpdate endpoint to confirm the email change.
	return c.JSON(map[string]string{"token": updateEmail.Token})
}

func (h *HTTPHandler) ValidateEmailUpdate(c *fiber.Ctx) error {
	var requestBody ValidateUpdateEmail
	if err := c.BodyParser(&requestBody); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, err, "invalid-request-body")
	}

	tx, errInitTx := h.writeDB.Begin()
	if errInitTx != nil {
		return errInitTx
	}
	defer tx.Rollback()

	userAccount, err := h.commonuser.Find().ByUUID(requestBody.AccountUUID)
	if err != nil {
		return err
	}

	err = h.commonuser.EmailUpdate().ValidateUpdate(tx, userAccount, requestBody.Token)
	if err != nil {
		return err
	}

	errCommit := tx.Commit()
	if errCommit != nil {
		return errCommit
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *HTTPHandler) ResendEmailUpdate(c *fiber.Ctx) error {
	account := c.Locals("account").(*account.Account)

	var requestBody UpdateEmail
	if err := c.BodyParser(&requestBody); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, err, "invalid-request-body")
	}

	tx, errInitTx := h.writeDB.Begin()
	if errInitTx != nil {
		return errInitTx
	}
	defer tx.Rollback()

	userAccount, err := h.commonuser.Find().ByUUID(account.GetUUID())
	if err != nil {
		return err
	}

	updateEmail, err := h.commonuser.EmailUpdate().RequestUpdate(tx, userAccount, requestBody.NewEmail)
	if err != nil {
		return err
	}

	errCommit := tx.Commit()
	if errCommit != nil {
		return errCommit
	}

	// NOTE: The returned token should be sent to the user's NEW email address
	// for verification. You should implement email delivery (SMTP, service provider, etc.)
	// to send this token to requestBody.NewEmail. The user must use this token
	// in the ValidateEmailUpdate endpoint to confirm the email change.
	return c.JSON(map[string]string{"token": updateEmail.Token})
}

func (h *HTTPHandler) RevokeEmailUpdate(c *fiber.Ctx) error {
	var requestBody RevokeUpdateEmail
	if err := c.BodyParser(&requestBody); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, err, "invalid-request-body")
	}

	tx, errInitTx := h.writeDB.Begin()
	if errInitTx != nil {
		return errInitTx
	}
	defer tx.Rollback()

	userAccount, err := h.commonuser.Find().ByUUID(requestBody.AccountUUID)
	if err != nil {
		return err
	}

	err = h.commonuser.EmailUpdate().Delete(tx, userAccount)
	if err != nil {
		return err
	}

	errCommit := tx.Commit()
	if errCommit != nil {
		return errCommit
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *HTTPHandler) UpdatePassword(c *fiber.Ctx) error {
	account := c.Locals("account").(*account.Account)

	var requestBody UpdatePassword
	if err := c.BodyParser(&requestBody); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, err, "invalid-request-body")
	}

	tx, errInitTx := h.writeDB.Begin()
	if errInitTx != nil {
		return errInitTx
	}
	defer tx.Rollback()

	userAccount, err := h.commonuser.Find().ByUUID(account.GetUUID())
	if err != nil {
		return err
	}

	err = h.commonuser.Password().Update(tx, userAccount, requestBody.OldPassword, requestBody.NewPassword)
	if err != nil {
		return err
	}

	errCommit := tx.Commit()
	if errCommit != nil {
		return errCommit
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *HTTPHandler) ForgotPassword(c *fiber.Ctx) error {
	var requestBody ForgotPassword
	if err := c.BodyParser(&requestBody); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, err, "invalid-request-body")
	}

	tx, errInitTx := h.writeDB.Begin()
	if errInitTx != nil {
		return errInitTx
	}
	defer tx.Rollback()

	account, err := h.commonuser.Find().ByEmail(requestBody.Email)
	if err != nil {
		return err
	}

	resetPassword, err := h.commonuser.Password().RequestReset(tx, account, nil)
	if err != nil {
		return err
	}

	errCommit := tx.Commit()
	if errCommit != nil {
		return errCommit
	}

	// NOTE: The returned token should be sent to the user's email address
	// for password reset verification. You should implement email delivery
	// (SMTP, service provider, etc.) to send this token to requestBody.Email.
	// The user must use this token in the ResetPassword endpoint to complete
	// the password reset process.
	return c.JSON(map[string]string{"token": resetPassword.Token})
}

func (h *HTTPHandler) ResetPassword(c *fiber.Ctx) error {
	var requestBody ResetPassword
	if err := c.BodyParser(&requestBody); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, err, "invalid-request-body")
	}

	tx, errInitTx := h.writeDB.Begin()
	if errInitTx != nil {
		return errInitTx
	}
	defer tx.Rollback()

	userAccount, err := h.commonuser.Find().ByUUID(requestBody.AccountUUID)
	if err != nil {
		return err
	}

	err = h.commonuser.Password().ValidateReset(tx, userAccount, requestBody.NewPassword, requestBody.Token)
	if err != nil {
		return err
	}

	errCommit := tx.Commit()
	if errCommit != nil {
		return errCommit
	}

	return c.SendStatus(fiber.StatusOK)
}

func NewHTTPHandler(commonuser *commonuser.Service) *HTTPHandler {
	return &HTTPHandler{
		commonuser: commonuser,
	}
}
