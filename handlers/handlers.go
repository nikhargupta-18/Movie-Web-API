package handlers

import (
	"net/http"
	"strconv"

	"movie-api-go/models"
	"movie-api-go/services"

	"github.com/gin-gonic/gin"
)

type MovieHandler struct {
	omdbService *services.OMDbService
}

func NewMovieHandler(omdbService *services.OMDbService) *MovieHandler {
	return &MovieHandler{
		omdbService: omdbService,
	}
}

// GetMovieDetails handles GET /api/movie?title=MovieTitle
func (h *MovieHandler) GetMovieDetails(c *gin.Context) {
	title := c.Query("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Title parameter is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	movie, err := h.omdbService.GetMovieByTitle(title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch movie details",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	if movie.Response == "False" {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Not Found",
			Message: movie.Error,
			Code:    http.StatusNotFound,
		})
		return
	}

	response := models.MovieDetailsResponse{
		Title:    movie.Title,
		Year:     movie.Year,
		Plot:     movie.Plot,
		Country:  movie.Country,
		Awards:   movie.Awards,
		Director: movie.Director,
		Ratings:  movie.Ratings,
	}

	c.JSON(http.StatusOK, response)
}

// GetEpisodeDetails handles GET /api/episode?series_title=SeriesTitle&season=1&episode_number=1
func (h *MovieHandler) GetEpisodeDetails(c *gin.Context) {
	seriesTitle := c.Query("series_title")
	seasonStr := c.Query("season")
	episodeStr := c.Query("episode_number")

	if seriesTitle == "" || seasonStr == "" || episodeStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "series_title, season, and episode_number parameters are required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	season, err := strconv.Atoi(seasonStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Season must be a valid number",
			Code:    http.StatusBadRequest,
		})
		return
	}

	episode, err := strconv.Atoi(episodeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Episode number must be a valid number",
			Code:    http.StatusBadRequest,
		})
		return
	}

	episodeDetails, err := h.omdbService.GetEpisodeDetails(seriesTitle, season, episode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch episode details",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	if episodeDetails.Response == "False" {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Not Found",
			Message: episodeDetails.Error,
			Code:    http.StatusNotFound,
		})
		return
	}

	response := models.EpisodeDetailsResponse{
		Title:       episodeDetails.Title,
		SeriesTitle: seriesTitle,
		Season:      episodeDetails.Season,
		Episode:     episodeDetails.Episode,
		Year:        episodeDetails.Year,
		Plot:        episodeDetails.Plot,
		Director:    episodeDetails.Director,
		Actors:      episodeDetails.Actors,
		ImdbRating:  episodeDetails.ImdbRating,
		Ratings:     episodeDetails.Ratings,
	}

	c.JSON(http.StatusOK, response)
}

// GetMoviesByGenre handles GET /api/movies/genre?genre=Action
func (h *MovieHandler) GetMoviesByGenre(c *gin.Context) {
	genre := c.Query("genre")
	if genre == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Genre parameter is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	movies, err := h.omdbService.SearchMoviesByGenre(genre)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to fetch movies by genre",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	if len(movies) == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Not Found",
			Message: "No movies found for the specified genre",
			Code:    http.StatusNotFound,
		})
		return
	}

	response := models.GenreMoviesResponse{
		Genre:  genre,
		Movies: movies,
		Total:  len(movies),
	}

	c.JSON(http.StatusOK, response)
}

// GetMovieRecommendations handles GET /api/recommendations?favorite_movie=MovieTitle
func (h *MovieHandler) GetMovieRecommendations(c *gin.Context) {
	favoriteMovie := c.Query("favorite_movie")
	if favoriteMovie == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "favorite_movie parameter is required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	recommendations, err := h.omdbService.GetMovieRecommendations(favoriteMovie)
	if err != nil {
		if err.Error() == "movie not found: "+favoriteMovie {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "Not Found",
				Message: "Favorite movie not found",
				Code:    http.StatusNotFound,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to generate recommendations",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, recommendations)
}

// HealthCheck handles GET /health
func (h *MovieHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "Movie API is running",
	})
}
