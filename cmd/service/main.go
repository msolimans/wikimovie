package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/msolimans/wikimovie/pkg/appconf"
	"github.com/msolimans/wikimovie/pkg/es"
	"github.com/msolimans/wikimovie/pkg/routes/monitor"
	"github.com/msolimans/wikimovie/pkg/routes/query"
	"github.com/msolimans/wikimovie/pkg/utils"
	"github.com/sirupsen/logrus"
)

func main() {

	app := fiber.New()

	cfg := &appconf.Configuration{}
	if err := appconf.LoadConfig(".", cfg); err != nil {
		panic("Cannot load config files")
	}

	app.Use(limiter.New(limiter.Config{
		Max:               cfg.Service.RateLimit.Max,
		Expiration:        time.Duration(cfg.Service.RateLimit.ExpirationInSecs) * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
		LimitReached: func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusTooManyRequests)
		},

		//todo: this needs distributed cache like redis/memcached
		// Storage: customStorage{},
	}))

	esConf := &es.ESConfig{
		Urls: cfg.ElasticSearch.Urls,
		// Urls:                []string{"http://localhost.localstack.cloud:4571"},
		IdleConnTimeout:     cfg.ElasticSearch.IdleConnTimeout,
		MaxIdleConnsPerHost: cfg.ElasticSearch.MaxIdleConnsPerHost,
		MaxIdleConns:        cfg.ElasticSearch.MaxIdleConns,
	}

	esClient, err := es.NewESClient(esConf)
	if err != nil {
		panic("Can not connect to ES")
	}

	healthRouter, err := monitor.NewHealthHandler(esClient)
	if err != nil {
		panic("Can not instantiate health router")
	}
	utils.NewRouter("/health", app, healthRouter)

	//auth middleware (commented for testing purposes only)
	app.Use(authMiddleware)

	movieRouter, err := query.NewMovieHandler(esClient)
	if err != nil {
		panic("Can not instantiate movie router")
	}

	//movies routes
	utils.NewRouter("/movies", app, movieRouter)

	// Start the server in a separate goroutine
	go func() {
		if err := app.Listen(fmt.Sprintf(":%v", cfg.Service.Port)); err != nil {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Create a channel to listen for OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)

	// ctrl+c or any other term signals
	<-quit
	logrus.Info("Shutting down server...")

	// context with 60 timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logrus.Fatalf("Error shutting down server: %v", err)
	}

	logrus.Info("Server gracefully stopped")
}

func authMiddleware(c *fiber.Ctx) error {
	authenticated := checkAuthentication(c)

	if !authenticated {
		return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
	}

	return c.Next()
}

func checkAuthentication(c *fiber.Ctx) bool {

	token := c.Get("Authorization")
	// just giving an idea of how I am designing this (accepting any token for now)
	const prefix = "Bearer "
	if len(token) > len(prefix) && token[:len(prefix)] == prefix {
		return true
	}

	return false
}
