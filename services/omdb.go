package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"

	"movie-api-go/models"
)

type OMDbService struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

func NewOMDbService() *OMDbService {
	return &OMDbService{
		APIKey:  os.Getenv("OMDB_API_KEY"),
		BaseURL: os.Getenv("OMDB_BASE_URL"),
		Client:  &http.Client{},
	}
}

// GetMovieByTitle fetches movie details by title
func (s *OMDbService) GetMovieByTitle(title string) (*models.OMDbResponse, error) {
	params := url.Values{}
	params.Add("apikey", s.APIKey)
	params.Add("t", title)
	params.Add("type", "movie")

	return s.makeRequest(params)
}

// GetEpisodeDetails fetches TV episode details
func (s *OMDbService) GetEpisodeDetails(seriesTitle string, season, episode int) (*models.OMDbResponse, error) {
	params := url.Values{}
	params.Add("apikey", s.APIKey)
	params.Add("t", seriesTitle)
	params.Add("Season", strconv.Itoa(season))
	params.Add("Episode", strconv.Itoa(episode))

	return s.makeRequest(params)
}

// SearchMoviesByGenre searches for movies by genre and returns top 15 by IMDb rating
func (s *OMDbService) SearchMoviesByGenre(genre string) ([]models.MovieBrief, error) {
	var allMovies []models.MovieBrief
	
	// Search with different popular movie titles to find movies of the specified genre
	searchTerms := []string{
		genre,
		fmt.Sprintf("%s movie", genre),
		fmt.Sprintf("best %s", genre),
	}
	
	// Also search by year to get more diverse results
	currentYear := 2024
	for year := currentYear; year >= currentYear-10; year-- {
		searchTerms = append(searchTerms, fmt.Sprintf("%s %d", genre, year))
	}
	
	for _, term := range searchTerms {
		movies, err := s.searchMovies(term, genre)
		if err != nil {
			continue
		}
		allMovies = append(allMovies, movies...)
		
		// Stop if we have enough movies
		if len(allMovies) >= 50 {
			break
		}
	}
	
	// Remove duplicates and filter by genre
	uniqueMovies := s.removeDuplicatesAndFilter(allMovies, genre)
	
	// Sort by IMDb rating
	sort.Slice(uniqueMovies, func(i, j int) bool {
		ratingI, _ := strconv.ParseFloat(uniqueMovies[i].ImdbRating, 64)
		ratingJ, _ := strconv.ParseFloat(uniqueMovies[j].ImdbRating, 64)
		return ratingI > ratingJ
	})
	
	// Return top 15
	if len(uniqueMovies) > 15 {
		uniqueMovies = uniqueMovies[:15]
	}
	
	return uniqueMovies, nil
}

// GetMovieRecommendations generates movie recommendations based on favorite movie
func (s *OMDbService) GetMovieRecommendations(favoriteTitle string) (*models.RecommendationResponse, error) {
	// Get favorite movie details
	favoriteMovie, err := s.GetMovieByTitle(favoriteTitle)
	if err != nil {
		return nil, err
	}
	
	if favoriteMovie.Response == "False" {
		return nil, fmt.Errorf("movie not found: %s", favoriteTitle)
	}
	
	response := &models.RecommendationResponse{
		FavoriteMovie: models.MovieBrief{
			Title:      favoriteMovie.Title,
			Year:       favoriteMovie.Year,
			ImdbRating: favoriteMovie.ImdbRating,
			Genre:      favoriteMovie.Genre,
			Director:   favoriteMovie.Director,
			Plot:       favoriteMovie.Plot,
		},
		Recommendations: []models.MovieLevel{},
	}
	
	// Level 1: Genre-based recommendations
	genres := strings.Split(favoriteMovie.Genre, ", ")
	var level1Movies []models.MovieBrief
	
	for _, genre := range genres {
		movies, err := s.searchMoviesForRecommendation(genre, favoriteTitle)
		if err != nil {
			continue
		}
		level1Movies = append(level1Movies, movies...)
	}
	
	level1Movies = s.removeDuplicatesAndLimit(level1Movies, 20)
	if len(level1Movies) > 0 {
		response.Recommendations = append(response.Recommendations, models.MovieLevel{
			Level:       1,
			Description: "Movies in the same genre",
			Movies:      level1Movies,
		})
	}
	
	// Level 2: Director-based recommendations
	directors := strings.Split(favoriteMovie.Director, ", ")
	var level2Movies []models.MovieBrief
	
	for _, director := range directors {
		if director != "N/A" && director != "" {
			movies, err := s.searchMoviesForRecommendation(director, favoriteTitle)
			if err != nil {
				continue
			}
			level2Movies = append(level2Movies, movies...)
		}
	}
	
	level2Movies = s.removeDuplicatesAndLimit(level2Movies, 20)
	if len(level2Movies) > 0 {
		response.Recommendations = append(response.Recommendations, models.MovieLevel{
			Level:       2,
			Description: "Movies by the same director",
			Movies:      level2Movies,
		})
	}
	
	// Level 3: Actor-based recommendations
	actors := strings.Split(favoriteMovie.Actors, ", ")
	var level3Movies []models.MovieBrief
	
	for i, actor := range actors {
		if i >= 2 { // Only use first 2 main actors
			break
		}
		if actor != "N/A" && actor != "" {
			movies, err := s.searchMoviesForRecommendation(actor, favoriteTitle)
			if err != nil {
				continue
			}
			level3Movies = append(level3Movies, movies...)
		}
	}
	
	level3Movies = s.removeDuplicatesAndLimit(level3Movies, 20)
	if len(level3Movies) > 0 {
		response.Recommendations = append(response.Recommendations, models.MovieLevel{
			Level:       3,
			Description: "Movies with the same main actors",
			Movies:      level3Movies,
		})
	}
	
	return response, nil
}

// Helper methods

func (s *OMDbService) makeRequest(params url.Values) (*models.OMDbResponse, error) {
	reqURL := fmt.Sprintf("%s?%s", s.BaseURL, params.Encode())
	
	resp, err := s.Client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	var omdbResp models.OMDbResponse
	if err := json.Unmarshal(body, &omdbResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	return &omdbResp, nil
}

func (s *OMDbService) searchMovies(searchTerm, targetGenre string) ([]models.MovieBrief, error) {
	params := url.Values{}
	params.Add("apikey", s.APIKey)
	params.Add("s", searchTerm)
	params.Add("type", "movie")
	
	reqURL := fmt.Sprintf("%s?%s", s.BaseURL, params.Encode())
	
	resp, err := s.Client.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var searchResp models.SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, err
	}
	
	if searchResp.Response == "False" {
		return []models.MovieBrief{}, nil
	}
	
	var movies []models.MovieBrief
	for _, result := range searchResp.Search {
		// Get detailed info for each movie
		movieDetails, err := s.GetMovieByTitle(result.Title)
		if err != nil {
			continue
		}
		
		if movieDetails.Response == "False" {
			continue
		}
		
		// Check if movie contains the target genre
		if strings.Contains(strings.ToLower(movieDetails.Genre), strings.ToLower(targetGenre)) {
			movies = append(movies, models.MovieBrief{
				Title:      movieDetails.Title,
				Year:       movieDetails.Year,
				ImdbRating: movieDetails.ImdbRating,
				Genre:      movieDetails.Genre,
				Director:   movieDetails.Director,
				Plot:       movieDetails.Plot,
			})
		}
	}
	
	return movies, nil
}

func (s *OMDbService) searchMoviesForRecommendation(searchTerm, excludeTitle string) ([]models.MovieBrief, error) {
	params := url.Values{}
	params.Add("apikey", s.APIKey)
	params.Add("s", searchTerm)
	params.Add("type", "movie")
	
	reqURL := fmt.Sprintf("%s?%s", s.BaseURL, params.Encode())
	
	resp, err := s.Client.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var searchResp models.SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, err
	}
	
	if searchResp.Response == "False" {
		return []models.MovieBrief{}, nil
	}
	
	var movies []models.MovieBrief
	for _, result := range searchResp.Search {
		// Skip the original movie
		if strings.EqualFold(result.Title, excludeTitle) {
			continue
		}
		
		// Get detailed info for each movie
		movieDetails, err := s.GetMovieByTitle(result.Title)
		if err != nil {
			continue
		}
		
		if movieDetails.Response == "False" {
			continue
		}
		
		movies = append(movies, models.MovieBrief{
			Title:      movieDetails.Title,
			Year:       movieDetails.Year,
			ImdbRating: movieDetails.ImdbRating,
			Genre:      movieDetails.Genre,
			Director:   movieDetails.Director,
			Plot:       movieDetails.Plot,
		})
	}
	
	return movies, nil
}

func (s *OMDbService) removeDuplicatesAndFilter(movies []models.MovieBrief, targetGenre string) []models.MovieBrief {
	seen := make(map[string]bool)
	var unique []models.MovieBrief
	
	for _, movie := range movies {
		key := strings.ToLower(movie.Title + movie.Year)
		if !seen[key] && strings.Contains(strings.ToLower(movie.Genre), strings.ToLower(targetGenre)) {
			// Only include movies with valid IMDb ratings
			if rating, err := strconv.ParseFloat(movie.ImdbRating, 64); err == nil && rating > 0 {
				seen[key] = true
				unique = append(unique, movie)
			}
		}
	}
	
	return unique
}

func (s *OMDbService) removeDuplicatesAndLimit(movies []models.MovieBrief, limit int) []models.MovieBrief {
	seen := make(map[string]bool)
	var unique []models.MovieBrief
	
	// Sort by IMDb rating first
	sort.Slice(movies, func(i, j int) bool {
		ratingI, _ := strconv.ParseFloat(movies[i].ImdbRating, 64)
		ratingJ, _ := strconv.ParseFloat(movies[j].ImdbRating, 64)
		return ratingI > ratingJ
	})
	
	for _, movie := range movies {
		key := strings.ToLower(movie.Title + movie.Year)
		if !seen[key] {
			// Only include movies with valid IMDb ratings
			if rating, err := strconv.ParseFloat(movie.ImdbRating, 64); err == nil && rating > 0 {
				seen[key] = true
				unique = append(unique, movie)
				
				if len(unique) >= limit {
					break
				}
			}
		}
	}
	
	return unique
}
