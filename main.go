package main

import (
	"github.com/21strive/commonuser"
	"github.com/21strive/commonuser/config"
	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
	"os"
	"time"
)

var SystemJWTSecret = "sgfhd0f2-ujfk2ws"
var SystemJWTIssuer = "commonuser-api.com"
var SystemJWTLifespan = 24 * 14 * time.Hour // 14 days

func main() {
	writeDB := CreatePostgresConnection(
		os.Getenv("DB_WRITE_HOST"), os.Getenv("DB_WRITE_PORT"), os.Getenv("DB_WRITE_USER"),
		os.Getenv("DB_WRITE_PASSWORD"), os.Getenv("DB_WRITE_NAME"), os.Getenv("DB_WRITE_SSLMODE"))
	defer writeDB.Close()
	readDB := CreatePostgresConnection(
		os.Getenv("DB_READ_HOST"), os.Getenv("DB_READ_PORT"), os.Getenv("DB_READ_USER"),
		os.Getenv("DB_READ_PASSWORD"), os.Getenv("DB_READ_NAME"), os.Getenv("DB_READ_SSLMODE"))
	defer readDB.Close()
	redis := ConnectRedis(os.Getenv("REDIS_HOST"), os.Getenv("REDIS_USER"),
		os.Getenv("REDIS_PASS"), false)

	config := config.DefaultConfig("account", SystemJWTSecret, SystemJWTIssuer, SystemJWTLifespan)

	commonuserService := commonuser.New(readDB, redis, config)
	commonuserFetchers := commonuser.NewFetchers(redis, config)
	httpHandler := NewHTTPHandler(commonuserService, commonuserFetchers, writeDB)

	app := fiber.New()

	app.Post("/register", httpHandler.Registration)
	app.Post("/register/verify", MiddlewareTokenAuth, httpHandler.VerifyRegistration)
	app.Post("/auth/username", httpHandler.AuthWithUsername)
	app.Post("/auth/google", httpHandler.AuthWithGoogle)
	app.Post("/auth/email", httpHandler.AuthWithEmail)
	app.Patch("/account", MiddlewareTokenAuth, httpHandler.UpdateAccount)
	app.Patch("/refresh", httpHandler.Refresh)
	app.Post("/email/update", MiddlewareTokenAuth, httpHandler.UpdateEmail)
	app.Post("/email/update/validate", httpHandler.ValidateEmailUpdate)
	app.Post("/email/update/revoke", httpHandler.RevokeEmailUpdate)
	app.Post("/password/update", MiddlewareTokenAuth, httpHandler.UpdatePassword)
	app.Post("/password/forgot", httpHandler.ForgotPassword)
	app.Post("/password/reset", httpHandler.ResetPassword)
	app.Get("/session", MiddlewareTokenAuth, httpHandler.FetchSession)
	app.Post("/session/revoke/:sessionUUID", MiddlewareTokenAuth, httpHandler.RevokeSession)
	app.Get("/content", MiddlewareTokenAuth, httpHandler.FetchContent)

	err := app.Listen(":3000")
	if err != nil {
		panic(err)
	}
}
