package main

import (
	"database/sql"
	"github.com/21strive/commonuser"
	"github.com/21strive/commonuser/account"
	"github.com/21strive/commonuser/session"
	"github.com/gofiber/fiber/v2"
	"time"
)

type HTTPHandler struct {
	writeDB           *sql.DB
	commonuser        *commonuser.Service
	commonuserFetcher *commonuser.Fetchers
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
	newAccount.SetName(requestBody.Name)
	newAccount.SetUsername(requestBody.Username)
	newAccount.SetEmail(requestBody.Email)
	newAccount.SetPassword(requestBody.Password)
	newAccount.SetAvatar(requestBody.Avatar)

	newSession := session.NewSession()
	newSession.SetAccountUUID(newAccount.GetUUID())
	newSession.SetDeviceId(requestBody.DeviceId)
	newSession.SetDeviceType(requestBody.DeviceType)
	newSession.SetUserAgent(requestBody.UserAgent)
	newSession.SetLastActiveAt(time.Now())
	newSession.SetLifeSpan(h.commonuser.Config().TokenLifespan)
	newSession.GenerateRefreshToken()

	accessToken, errGen := newAccount.GenerateAccessToken(
		h.commonuser.Config().JWTSecret,
		h.commonuser.Config().JWTIssuer,
		h.commonuser.Config().JWTLifespan,
		newSession.GetRandId(),
	)
	if errGen != nil {
		return errGen
	}

	verification, regError := h.commonuser.Register(tx, newAccount, true)
	if regError != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, regError, "internal-server-error")
	}

	errorCreateSession := h.commonuser.Session().Create(tx, newSession)
	if errorCreateSession != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, errorCreateSession, "internal-server-error")
	}

	errCommit := tx.Commit()
	if errCommit != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, errCommit, "internal-server-error")
	}

	response := map[string]string{
		"accessToken": accessToken,
	}
	if verification != nil {
		response["verificationCode"] = *verification
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refreshToken",
		Value:    newSession.RefreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})
	return c.JSON(response)
}

func (h *HTTPHandler) VerifyRegistration(c *fiber.Ctx) error {
	account := c.Locals("account").(*account.Account)
	sessionId := c.Locals("sessionid").(string)

	var requestBody VerifyRegistration
	if err := c.BodyParser(&requestBody); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, err, "invalid-request-body")
	}

	tx, errInitTx := h.writeDB.Begin()
	if errInitTx != nil {
		return errInitTx
	}
	defer tx.Rollback()

	accountFromDB, isValid, errVerify := h.commonuser.Verification().Verify(tx, account.GetUUID(), requestBody.VerificationCode)
	if errVerify != nil {
		return errVerify
	}
	if !isValid {
		return ErrorResponse(c, fiber.StatusBadRequest, nil, "unauthorized")
	}

	newAccessToken, errGen := accountFromDB.GenerateAccessToken(
		h.commonuser.Config().JWTSecret,
		h.commonuser.Config().JWTIssuer,
		h.commonuser.Config().JWTLifespan,
		sessionId,
	)
	if errGen != nil {
		return errGen
	}

	errCommit := tx.Commit()
	if errCommit != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, errCommit, "internal-server-error")
	}

	return c.JSON(map[string]string{"accessToken": newAccessToken})
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

	accessToken, refreshToken, errToken := h.commonuser.NativeAuthenticate().ByEmail(
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

	accessToken, refreshToken, errToken := h.commonuser.NativeAuthenticate().ByUsername(
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

	updatedAccount, err := h.commonuser.Update(tx, account.GetUUID(), updateOpt)
	if err != nil {
		return err
	}

	errCommit := tx.Commit()
	if errCommit != nil {
		return errCommit
	}

	newAccessToken, errGenerate := updatedAccount.GenerateAccessToken(
		h.commonuser.Config().JWTSecret,
		h.commonuser.Config().JWTIssuer,
		h.commonuser.Config().JWTLifespan,
		account.GetUUID(),
	)
	if errGenerate != nil {
		return errGenerate
	}

	return c.JSON(map[string]string{"accessToken": newAccessToken})
}

func (h *HTTPHandler) Refresh(c *fiber.Ctx) error {
	account := c.Locals("account").(*account.Account)
	refreshToken := c.Cookies("refreshToken")
	if refreshToken == "" {
		return ErrorResponse(c, fiber.StatusUnauthorized, nil, "missing-refresh-token")
	}

	tx, errInitTx := h.writeDB.Begin()
	if errInitTx != nil {
		return errInitTx
	}

	defer tx.Rollback()

	newAccessToken, newRefreshToken, errRefresh := h.commonuser.Session().Refresh(tx, account, refreshToken)
	if errRefresh != nil {
		return errRefresh
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refreshToken",
		Value:    newRefreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	})
	return c.JSON(map[string]string{"accessToken": newAccessToken})
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
	return c.JSON(map[string]string{"token": updateEmail.Token, "revokeToken": updateEmail.RevokeToken})
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

	err := h.commonuser.EmailUpdate().ValidateUpdate(tx, requestBody.AccountUUID, requestBody.Token)
	if err != nil {
		return err
	}

	errCommit := tx.Commit()
	if errCommit != nil {
		return errCommit
	}

	errSeedSessions := h.commonuser.Session().SeedByAccount(requestBody.AccountUUID)
	if errSeedSessions != nil {
		return errSeedSessions
	}

	return c.SendStatus(fiber.StatusOK)
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

	err := h.commonuser.EmailUpdate().RevokeUpdate(tx, requestBody.AccountUUID, requestBody.RevokeToken)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, err, "failed-revoke")
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

	err := h.commonuser.Password().Update(tx, account.GetUUID(), requestBody.OldPassword, requestBody.NewPassword)
	if err != nil {
		return err
	}

	errCommit := tx.Commit()
	if errCommit != nil {
		return errCommit
	}

	errSeedSessions := h.commonuser.Session().SeedByAccount(account.GetUUID())
	if errSeedSessions != nil {
		return errSeedSessions
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

	errSeedSessions := h.commonuser.Session().SeedByAccount(userAccount.GetUUID())
	if errSeedSessions != nil {
		return errSeedSessions
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *HTTPHandler) FetchContent(c *fiber.Ctx) error {
	account := c.Locals("account").(*account.Account)
	sessionId := c.Locals("sessionid").(string)

	// check session validity from cache
	_, errCheckSession := h.commonuserFetcher.PingSession(sessionId)
	if errCheckSession != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	content := "Hi " + account.Name + ". If you can see this, you are authenticated."
	return c.SendString(content)
}

func NewHTTPHandler(commonuser *commonuser.Service, commonuserFetchers *commonuser.Fetchers, writeDB *sql.DB) *HTTPHandler {
	return &HTTPHandler{
		commonuser:        commonuser,
		commonuserFetcher: commonuserFetchers,
		writeDB:           writeDB,
	}
}
