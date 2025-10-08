package main

import (
	"dnsmesh/internal/database"
	"dnsmesh/internal/handlers"
	"dnsmesh/internal/middleware"
	"dnsmesh/pkg/crypto"
	"log"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize crypto
	if err := crypto.Initialize(); err != nil {
		log.Fatalf("Failed to initialize crypto: %v", err)
	}

	// Initialize database
	if err := database.Initialize(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Setup Gin
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Setup CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Setup sessions
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		sessionSecret = "default-secret-change-in-production"
	}
	store := cookie.NewStore([]byte(sessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: 1,
	})
	r.Use(sessions.Sessions("dnsmesh_session", store))

	// Public routes
	r.POST("/api/auth/login", handlers.Login)

	// Protected routes
	protected := r.Group("/api")
	protected.Use(middleware.AuthRequired())
	{
		protected.POST("/auth/logout", handlers.Logout)
		protected.GET("/auth/user", handlers.GetCurrentUser)
		protected.POST("/auth/change-password", handlers.ChangePassword)
		protected.POST("/auth/change-username", handlers.ChangeUsername)

		// Provider routes
		protected.GET("/providers", handlers.GetProviders)
		protected.POST("/providers", handlers.CreateProvider)
		protected.PUT("/providers/:id", handlers.UpdateProvider)
		protected.DELETE("/providers/:id", handlers.DeleteProvider)
		protected.POST("/providers/:id/sync", handlers.SyncProvider)

		// DNS Record routes
		protected.GET("/records", handlers.GetRecords)
		protected.POST("/records", handlers.CreateRecord)
		protected.PUT("/records/:id", handlers.UpdateRecord)
		protected.POST("/records/:id/hide", handlers.HideRecord)
		protected.DELETE("/records/:id", handlers.DeleteRecord)
		protected.POST("/records/import", handlers.ImportRecords)
		protected.POST("/records/reanalyze", handlers.ReanalyzeRecords)

		// Audit log routes
		protected.GET("/audit-logs", handlers.GetAuditLogs)
	}

	// Serve static files (frontend)
	r.Static("/assets", "./public/assets")
	r.StaticFile("/", "./public/index.html")
	r.NoRoute(func(c *gin.Context) {
		c.File("./public/index.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
