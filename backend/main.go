package main

import (
	"fmt"
	"log"
	"os"
	"split-it/backend/config"
	"split-it/backend/routes"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found")
	}

	// Initialize database
	config.ConnectDB()

	// Initialize Firebase
	config.InitializeFirebase()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(helmet.New())

	// CORS configuration
	allowedOrigin := os.Getenv("CLIENT_URL")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:3000"
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigin + ",http://localhost:3000,http://localhost:3001",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	// Rate limiting for API routes
	app.Use("/api", limiter.New(limiter.Config{
		Max:        100,
		Expiration: 15 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Too many requests from this IP, please try again later.",
			})
		},
	}))

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success":   true,
			"message":   "Server is running",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Setup routes
	routes.SetupUserRoutes(app)
	routes.SetupGroupRoutes(app)

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Route not found",
		})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	fmt.Printf("🚀 Server running on port %s\n", port)
	fmt.Printf("📊 Environment: %s\n", os.Getenv("NODE_ENV"))
	fmt.Printf("🌐 Client URL: %s\n", allowedOrigin)

	log.Fatal(app.Listen(":" + port))
}
