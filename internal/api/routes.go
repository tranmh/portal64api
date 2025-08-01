package api

import (
	"net/http"
	
	"portal64api/internal/api/handlers"
	"portal64api/internal/api/middleware"
	"portal64api/internal/database"
	"portal64api/internal/repositories"
	"portal64api/internal/services"

	"github.com/gin-gonic/gin"
	
	docs "portal64api/docs/generated" // swagger docs
)

// SetupRoutes configures all API routes
func SetupRoutes(dbs *database.Databases) *gin.Engine {
	// Ensure swagger docs are loaded
	_ = docs.SwaggerInfo
	
	// Create repositories
	playerRepo := repositories.NewPlayerRepository(dbs)
	clubRepo := repositories.NewClubRepository(dbs)
	tournamentRepo := repositories.NewTournamentRepository(dbs)
	addressRepo := repositories.NewAddressRepository(dbs)

	// Create services
	playerService := services.NewPlayerService(playerRepo, clubRepo)
	clubService := services.NewClubService(clubRepo)
	clubService.SetPlayerRepository(playerRepo) // Set player repo for club profile functionality
	tournamentService := services.NewTournamentService(tournamentRepo)
	addressService := services.NewAddressService(addressRepo)

	// Create handlers
	playerHandler := handlers.NewPlayerHandler(playerService)
	clubHandler := handlers.NewClubHandler(clubService)
	tournamentHandler := handlers.NewTournamentHandler(tournamentService)
	addressHandler := handlers.NewAddressHandler(addressService)

	// Create router
	router := gin.New()

	// Apply middleware
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.ErrorHandlingMiddleware())

	// Swagger documentation - Manual implementation since gin-swagger has issues
	// Serve the swagger JSON docs
	router.GET("/swagger/doc.json", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.String(200, docs.SwaggerInfo.ReadDoc())
	})
	
	// Serve a simple Swagger UI HTML page
	router.GET("/swagger/", func(c *gin.Context) {
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Portal64 API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui.css" />
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui-bundle.js"></script>
    <script>
        SwaggerUIBundle({
            url: '/swagger/doc.json',
            dom_id: '#swagger-ui',
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIBundle.presets.standalone
            ]
        });
    </script>
</body>
</html>`
		c.Header("Content-Type", "text/html")
		c.String(200, html)
	})
	
	// Also serve at index.html
	router.GET("/swagger/index.html", func(c *gin.Context) {
		c.Redirect(301, "/swagger/")
	})
	
	// Add base swagger route redirect
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/")
	})

	// Static files for demo interface
	router.Static("/demo", "./demo")
	
	// Redirect root to demo
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/demo/")
	})

	// Health check endpoint
	router.GET("/health", handlers.HealthCheck)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Player routes
		players := v1.Group("/players")
		{
			players.GET("", playerHandler.SearchPlayers)
			players.GET("/:id", playerHandler.GetPlayer)
			players.GET("/:id/rating-history", playerHandler.GetPlayerRatingHistory)
		}

		// Club routes
		clubs := v1.Group("/clubs")
		{
			clubs.GET("", clubHandler.SearchClubs)
			clubs.GET("/all", clubHandler.GetAllClubs)
			clubs.GET("/:id", clubHandler.GetClub)
			clubs.GET("/:id/players", playerHandler.GetPlayersByClub)
			clubs.GET("/:id/profile", clubHandler.GetClubProfile)
		}

		// Tournament routes
		tournaments := v1.Group("/tournaments")
		{
			tournaments.GET("", tournamentHandler.SearchTournaments)
			tournaments.GET("/upcoming", tournamentHandler.GetUpcomingTournaments)
			tournaments.GET("/recent", tournamentHandler.GetRecentTournaments)
			tournaments.GET("/date-range", tournamentHandler.GetTournamentsByDateRange)
			tournaments.GET("/:id", tournamentHandler.GetTournament)
		}

		// Address routes
		addresses := v1.Group("/addresses")
		{
			addresses.GET("/regions", addressHandler.GetAvailableRegions)
			addresses.GET("/:region", addressHandler.GetRegionAddresses)
			addresses.GET("/:region/types", addressHandler.GetAddressTypes)
			addresses.GET("/:region/:type", addressHandler.GetRegionAddressesByType)
		}
	}

	return router
}
