# Movie API Go

A comprehensive REST API built with Go and Gin framework that provides movie information, TV episode details, genre-based movie searches, and intelligent movie recommendations using the OMDb API.

## Features

### 1. Movie Details API
- **Endpoint**: `GET /api/movie?title=<movie_title>`
- **Description**: Fetches detailed information about a movie
- **Response**: Title, Year, Plot, Country, Awards, Director, Ratings

### 2. TV Episode Details API
- **Endpoint**: `GET /api/episode?series_title=<series>&season=<num>&episode_number=<num>`
- **Description**: Retrieves specific details for a TV show episode
- **Response**: Episode title, series info, plot, director, actors, ratings

### 3. Genre-Based Movie API
- **Endpoint**: `GET /api/movies/genre?genre=<genre>`
- **Description**: Returns top 15 movies in a specified genre, sorted by IMDb rating
- **Response**: List of movies with ratings, sorted by popularity

### 4. Movie Recommendation Engine
- **Endpoint**: `GET /api/recommendations?favorite_movie=<movie_title>`
- **Description**: Provides intelligent movie recommendations based on a favorite movie
- **Algorithm**: 
  - Level 1: Genre-based recommendations (highest priority)
  - Level 2: Director-based recommendations
  - Level 3: Actor-based recommendations (lowest priority)
- **Response**: Hierarchical recommendations with up to 20 movies per level

## Prerequisites

### Install Go
1. **macOS**: 
   ```bash
   brew install go
   ```
   
2. **Alternative**: Download from [https://golang.org/dl/](https://golang.org/dl/)

3. **Verify installation**:
   ```bash
   go version
   ```

### Get OMDb API Key
1. Visit [http://www.omdbapi.com/apikey.aspx](http://www.omdbapi.com/apikey.aspx)
2. Sign up for a free API key
3. Save your API key for the next step

## Setup Instructions

### 1. Clone/Navigate to Project
```bash
cd movie-api-go
```

### 2. Configure Environment Variables
Edit the `.env` file and replace `your_api_key_here` with your actual OMDb API key:

```env
# OMDb API Configuration
OMDB_API_KEY=your_actual_api_key_here
OMDB_BASE_URL=http://www.omdbapi.com/

# Server Configuration
PORT=8080
```

### 3. Install Dependencies
```bash
go mod tidy
```

### 4. Run the Application
```bash
go run main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### Health Check
```bash
curl http://localhost:8080/health
```

### 1. Get Movie Details
```bash
curl "http://localhost:8080/api/movie?title=The Matrix"
```

**Example Response:**
```json
{
  "title": "The Matrix",
  "year": "1999",
  "plot": "A computer programmer is led to fight an underground war against powerful computers who have constructed his entire reality with a system called the Matrix.",
  "country": "United States",
  "awards": "Won 4 Oscars. 42 wins & 51 nominations total",
  "director": "Lana Wachowski, Lilly Wachowski",
  "ratings": [
    {
      "Source": "Internet Movie Database",
      "Value": "8.7/10"
    }
  ]
}
```

### 2. Get Episode Details
```bash
curl "http://localhost:8080/api/episode?series_title=Breaking Bad&season=1&episode_number=1"
```

### 3. Get Movies by Genre
```bash
curl "http://localhost:8080/api/movies/genre?genre=Action"
```

### 4. Get Movie Recommendations
```bash
curl "http://localhost:8080/api/recommendations?favorite_movie=The Dark Knight"
```

**Example Response:**
```json
{
  "favorite_movie": {
    "title": "The Dark Knight",
    "year": "2008",
    "imdb_rating": "9.0",
    "genre": "Action, Crime, Drama",
    "director": "Christopher Nolan",
    "plot": "When the menace known as the Joker wreaks havoc..."
  },
  "recommendations": [
    {
      "level": 1,
      "description": "Movies in the same genre",
      "movies": [...]
    },
    {
      "level": 2,
      "description": "Movies by the same director",
      "movies": [...]
    },
    {
      "level": 3,
      "description": "Movies with the same main actors",
      "movies": [...]
    }
  ]
}
```

## Error Handling

The API returns appropriate HTTP status codes and error messages:

- **400 Bad Request**: Missing or invalid parameters
- **404 Not Found**: Movie/episode not found
- **500 Internal Server Error**: API or server errors

**Example Error Response:**
```json
{
  "error": "Not Found",
  "message": "Movie not found!",
  "code": 404
}
```

## Project Structure

```
movie-api-go/
├── main.go              # Application entry point
├── models/
│   └── models.go        # Data structures and models
├── services/
│   └── omdb.go         # OMDb API service layer
├── handlers/
│   └── handlers.go     # HTTP request handlers
├── go.mod              # Go module file
├── .env                # Environment variables
├── .gitignore          # Git ignore file
└── README.md           # This file
```

## Development

### Adding New Features
1. Add new models in `models/models.go`
2. Implement business logic in `services/omdb.go`
3. Create HTTP handlers in `handlers/handlers.go`
4. Register routes in `main.go`

### Testing
```bash
# Run tests (when implemented)
go test ./...

# Build the application
go build -o movie-api main.go

# Run the built binary
./movie-api
```

## Security Features

- Environment variables for API key management
- CORS middleware for cross-origin requests
- Input validation and sanitization
- Proper error handling without exposing sensitive information

## Rate Limiting

Be aware of OMDb API rate limits:
- Free tier: 1,000 requests per day


## Troubleshooting

### Common Issues

1. **"OMDB_API_KEY environment variable is required"**
   - Make sure you've set your API key in the `.env` file

2. **"go: command not found"**
   - Install Go using the instructions above

3. **API returns empty results**
   - Check your API key is valid
   - Verify the movie/show title spelling
   - Check OMDb API status

4. **Port already in use**
   - Change the PORT in `.env` file
   - Or kill the process using the port: `lsof -ti:8080 | xargs kill`

