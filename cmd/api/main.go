package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/application/commands"
	identityinfra "github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/infrastructure/postgres"
	redisstore "github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/infrastructure/redis"
	identityhttp "github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/http"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/cache"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/config"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/database"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/mailer"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	auth.Init(cfg.JWTSecret)

	pool, err := database.Connect(cfg.DSN())
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	redis, err := cache.Connect(cfg.RedisAddr, cfg.RedisPass)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer redis.Close()

	var mail mailer.Mailer
	if cfg.SMTPHost != "" {
		mail = mailer.NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)
	} else {
		mail = &mailer.MockMailer{}
	}

	userRepo := identityinfra.NewPostgresUserRepository(pool)
	tokenStore := redisstore.NewTokenStore(redis)
	identitySvc := commands.NewService(userRepo, tokenStore, mail)
	identityHandler := identityhttp.NewHandler(identitySvc)

	app := fiber.New(fiber.Config{
		BodyLimit:    1 * 1024 * 1024, // 1 MB
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "INTERNAL_ERROR",
					"message": err.Error(),
				},
			})
		},
	})

	app.Use(recover.New())
	app.Use(logger.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	identityhttp.RegisterRoutes(app, identityHandler)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.AppPort)
		log.Printf("listening on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Printf("server: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down...")
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
