package main

import (
	"github.com/21strive/commonuser"
	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
)

func main() {
	commonuser := commonuser.New()
	httpHandler := NewHTTPHandler(commonuser)

	app := fiber.New()

	app.Post("/register", httpHandler.Registration)
	app.Post("/register/verify", httpHandler.VerifyRegistration)
	app.Post("/auth/username", httpHandler.AuthWithUsername)
	app.Post("/auth/email", httpHandler.AuthWithEmail)
	app.Patch("/account", httpHandler.UpdateAccount)
	app.Patch("/refresh", httpHandler.Refresh)
	app.Post("/email/update", httpHandler.UpdateEmail)
	app.Post("/email/update/validate", httpHandler.ValidateEmailUpdate)
	app.Post("/email/update/resend", httpHandler.ResendEmailUpdate)
	app.Post("/email/update/revoke", httpHandler.RevokeEmailUpdate)
	app.Post("/password/update", httpHandler.UpdatePassword)
	app.Post("/password/forgot", httpHandler.ForgotPassword)
	app.Post("/password/reset", httpHandler.ResetPassword)

	// Getter
	app.Get("/user", httpHandler.GetUser)

	err := app.Listen(":3000")
	if err != nil {
		panic(err)
	}
}
