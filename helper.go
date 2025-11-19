package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/21strive/commonuser/account"
	"github.com/21strive/commonuser/jwt_impl"
	"github.com/21strive/item"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"
)

var Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

type ServiceError struct {
	Code string `json:"code"`
	ID   string `json:"id"`
}

func CreatePostgresConnection(host string, port string, user string, password string, dbname string, sslmode string) *sql.DB {
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to open database connection: %w", err))
	}

	if err := db.Ping(); err != nil {
		db.Close()
		log.Fatal(fmt.Errorf("failed to ping database: %w", err))
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Successfully connected to PostgreSQL database")
	return db
}

func ConnectRedis(host string, username string, password string, isClustered bool) redis.UniversalClient {
	if host == "" {
		log.Fatal("REDIS_HOST environment variable not set")
	}

	if isClustered {
		clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    []string{host},
			Username: username,
			Password: password,
		})

		_, err := clusterClient.Ping(context.Background()).Result()
		if err != nil {
			log.Fatal(err)
		}

		return clusterClient
	}

	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Username: username,
		Password: password,
		DB:       0,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func ErrorResponse(c *fiber.Ctx, status int, error error, appCode string, source ...string) error {
	errorId := item.RandId()

	response := ServiceError{
		Code: appCode,
		ID:   errorId,
	}

	type LogEntry struct {
		json.RawMessage
	}

	var inputBody json.RawMessage
	if c.Request().Body() != nil && len(c.Request().Body()) > 0 {
		var compactJSON bytes.Buffer
		json.Compact(&compactJSON, c.Request().Body())
		inputBody = compactJSON.Bytes()
	}

	logEntry := LogEntry{inputBody}

	var sourceStr string
	if len(source) > 1 {
		sourceStr = strings.Join(source, ".")
	} else if len(source) == 1 {
		sourceStr = source[0]
	}

	var returnedError string
	if error != nil {
		returnedError = error.Error()
	} else {
		returnedError = errors.New("error").Error()
	}
	Logger.Error("endpoint-error",
		"component", "paystore", "source", sourceStr, "appCode", appCode,
		"error", returnedError, "ID", errorId, "input", logEntry)

	c.Set("Content-Type", "application/json")
	return c.Status(status).JSON(response)
}

func MiddlewareTokenAuth(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return ErrorResponse(c, fiber.StatusUnauthorized, errors.New("authorization header required"), "E1006", "MiddlewareTokenAuth.MissingHeader")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return ErrorResponse(c, fiber.StatusUnauthorized, errors.New("bearer token required"), "E1006", "MiddlewareTokenAuth.InvalidFormat")
	}

	jwtHandler := jwt_impl.NewJWTHandler(SystemJWTSecret, SystemJWTIssuer, int(SystemJWTLifespan))
	userClaims, err := jwtHandler.ParseAccessToken(tokenString)
	if err != nil {
		return ErrorResponse(c, fiber.StatusUnauthorized, err, "E1006", "MiddlewareTokenAuth.ParseToken")
	}

	account := account.New()
	account.UUID = userClaims.UUID
	account.RandId = userClaims.RandId
	account.Base.Name = userClaims.Name
	account.Base.Username = userClaims.Username
	account.Base.Email = userClaims.Email
	account.Base.Avatar = userClaims.Avatar

	c.Locals("account", account)
	c.Locals("sessionid", userClaims.SessionID)
	return c.Next()
}
