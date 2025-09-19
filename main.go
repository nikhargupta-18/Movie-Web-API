package main

import (
	"log"
	"os"

	"movie-api-go/handlers"
	"movie-api-go/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Validate required environment variables
	apiKey := os.Getenv("OMDB_API_KEY")
	if apiKey == "" || apiKey == "your_api_key_here" {
		log.Fatal("OMDB_API_KEY environment variable is required. Please set it in your .env file")
	}

	baseURL := os.Getenv("OMDB_BASE_URL")
	if baseURL == "" {
		baseURL = "http://www.omdbapi.com/"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize services
	omdbService := services.NewOMDbService()

	// Initialize handlers
	movieHandler := handlers.NewMovieHandler(omdbService)

	// Setup Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	router.GET("/health", movieHandler.HealthCheck)

	// API routes
	api := router.Group("/api")
	{
		// 1. Movie Details API
		api.GET("/movie", movieHandler.GetMovieDetails)

		// 2. TV Episode Details API
		api.GET("/episode", movieHandler.GetEpisodeDetails)

		// 3. Genre-Based Movie API
		api.GET("/movies/genre", movieHandler.GetMoviesByGenre)

		// 4. Movie Recommendation Engine
		api.GET("/recommendations", movieHandler.GetMovieRecommendations)
	}

	// Start server
	log.Printf("Starting server on port %s", port)
	log.Printf("API endpoints available:")
	log.Printf("  GET /health - Health check")
	log.Printf("  GET /api/movie?title=<movie_title> - Get movie details")
	log.Printf("  GET /api/episode?series_title=<series>&season=<num>&episode_number=<num> - Get episode details")
	log.Printf("  GET /api/movies/genre?genre=<genre> - Get top 15 movies by genre")
	log.Printf("  GET /api/recommendations?favorite_movie=<movie_title> - Get movie recommendations")

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
