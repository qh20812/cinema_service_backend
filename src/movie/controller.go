package movie

import (
	"main/src/user"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Controller struct {
	Repo *Repository
	UserRepo *user.Repository
}

func NewController(repo *Repository, userRepo *user.Repository) *Controller {
	return &Controller{Repo: repo, UserRepo: userRepo}
}

type CreateMovieInput struct {
	Title       string `json:"title" binding:"required"`
	AgeRating   string `json:"age_rating" binding:"required,oneof=P K T13 T16 T18"`
	DurationMin int    `json:"duration_min" binding:"required,gt=0"`
	Showtime    string `json:"showtime" binding:"required"`
	TotalSeats  int    `json:"total_seats" binding:"required,gt=0"`
}

func (ctrl *Controller) GetAll(c *gin.Context) {
	movies, err := ctrl.Repo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch movies"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"movies": movies})
}

func (ctrl *Controller) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie id"})
		return
	}

	movie, err := ctrl.Repo.FindByID(id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find movie"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"movie": movie})
}

func (ctrl *Controller) Create(c *gin.Context) {
	var input CreateMovieInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	showtime, err := time.Parse(time.RFC3339, input.Showtime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "showtime must be RFC3339 format"})
		return
	}

	movie := Movie{
		Title:          input.Title,
		AgeRating:      AgeRating(input.AgeRating),
		DurationMin:    input.DurationMin,
		Showtime:       showtime,
		TotalSeats:     input.TotalSeats,
		AvailableSeats: input.TotalSeats,
	}

	movie.PrepareForCreate()
	if err := movie.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.Repo.Create(&movie); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create movie"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"movie": movie})
}

func (ctrl *Controller) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie id"})
		return
	}

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if showtimeRaw, ok := input["showtime"].(string); ok && showtimeRaw != "" {
		parsed, err := time.Parse(time.RFC3339, showtimeRaw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "showtime must be RFC3339 format"})
			return
		}
		input["showtime"] = parsed
	}

	if totalSeats, ok := input["total_seats"].(float64); ok {
		input["total_seats"] = int(totalSeats)
	}
	if duration, ok := input["duration_min"].(float64); ok {
		input["duration_min"] = int(duration)
	}

	if err := ctrl.Repo.Update(id, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update movie"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Movie updated successfully"})
}

func (ctrl *Controller) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie id"})
		return
	}

	if err := ctrl.Repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete movie"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Movie deleted successfully"})
}
