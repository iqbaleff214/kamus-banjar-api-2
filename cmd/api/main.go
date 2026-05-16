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
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	aicommands "github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/application/commands"
	aihttp "github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/http"
	aiinfra "github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/infrastructure/postgres"
	openrouterclient "github.com/iqbaleff214/kamus-banjar-api-2/internal/ai/infrastructure/openrouter"
	communitycmd "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/application/commands"
	communityhttp "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/http"
	communityinfra "github.com/iqbaleff214/kamus-banjar-api-2/internal/community/infrastructure/postgres"
	dictcmd "github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/application/commands"
	dictqry "github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/application/queries"
	dicthttp "github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/http"
	dictinfra "github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/infrastructure/postgres"
	"github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/application/commands"
	identityhttp "github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/http"
	identityinfra "github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/infrastructure/postgres"
	redisstore "github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/infrastructure/redis"
	moderationcmd "github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/application/commands"
	moderationhttp "github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/http"
	moderationinfra "github.com/iqbaleff214/kamus-banjar-api-2/internal/moderation/infrastructure/postgres"
	"github.com/iqbaleff214/kamus-banjar-api-2/internal/docs"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/auth"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/cache"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/config"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/database"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/logging"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/mailer"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/middleware"
	"github.com/iqbaleff214/kamus-banjar-api-2/pkg/ratelimit"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	logging.Init(cfg.AppEnv)
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
	defer func() { _ = redis.Close() }()

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
	app.Use(middleware.RequestID())
	app.Use(logging.RequestLogger())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	rlCounter := ratelimit.NewRedisCounter(redis)

	identityhttp.RegisterRoutes(app, identityHandler, rlCounter)

	wordRepo := dictinfra.NewPostgresWordRepository(pool)
	wordQry := dictqry.NewWordQueryService(wordRepo)
	wordCmd := dictcmd.NewWordCommandService(wordRepo)
	dicthttp.RegisterRoutes(app, dicthttp.NewHandler(wordQry, wordCmd), rlCounter)

	contribRepo := communityinfra.NewPostgresContributionRepository(pool)
	voteRepo := communityinfra.NewPostgresVoteRepository(pool)
	bookmarkRepo := communityinfra.NewPostgresBookmarkRepository(pool)
	commentRepo := communityinfra.NewPostgresCommentRepository(pool)
	wordChecker := communityinfra.NewWordExistChecker(pool)
	defChecker := communityinfra.NewDefinitionExistChecker(pool)
	emailVerifier := communityinfra.NewUserEmailVerifier(pool)

	contribSvc := communitycmd.NewContributionService(contribRepo, wordChecker, emailVerifier)
	voteSvc := communitycmd.NewVoteService(voteRepo, wordChecker, defChecker)
	bookmarkSvc := communitycmd.NewBookmarkService(bookmarkRepo, wordChecker)
	commentSvc := communitycmd.NewCommentService(commentRepo, wordChecker)

	communityHandler := communityhttp.NewHandler(contribSvc, voteSvc, bookmarkSvc, commentSvc)
	communityhttp.RegisterRoutes(app, communityHandler, rlCounter)

	modRepo := moderationinfra.NewPostgresModerationRepository(pool)
	moderationSvc := moderationcmd.NewModerationService(modRepo, contribRepo, wordRepo, userRepo)
	moderationhttp.RegisterRoutes(app, moderationhttp.NewHandler(moderationSvc))

	aiClient, err := openrouterclient.New(cfg)
	if err != nil {
		log.Fatalf("ai: %v", err)
	}
	aiSvc := aicommands.NewTranslateService(aiClient, cfg.OpenRouterModel)

	aiRequestRepo := aiinfra.NewPostgresAIRequestRepository(pool)
	enrichmentSvc := aicommands.NewEnrichmentService(aiRequestRepo, wordRepo, contribRepo, modRepo, aiClient, cfg.OpenRouterModel)
	aihttp.RegisterRoutes(app, aihttp.NewHandler(aiSvc), aihttp.NewAdminHandler(enrichmentSvc), rlCounter)

	docs.RegisterRoutes(app)

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
	if err := app.ShutdownWithTimeout(30 * time.Second); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
