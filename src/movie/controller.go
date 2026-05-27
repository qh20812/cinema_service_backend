package movie

import (
	"main/src/user"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Controller struct {
	Repo          *Repository
	UserRepo      *user.Repository
	TicketCounter  interface{ CountByMovieID(movieID bson.ObjectID) (int64, error) }
}

func NewController(repo *Repository, userRepo *user.Repository, ticketCounter interface{ CountByMovieID(movieID bson.ObjectID) (int64, error) }) *Controller {
	return &Controller{Repo: repo, UserRepo: userRepo, TicketCounter: ticketCounter}
}

type CreateMovieInput struct {
	Title       string `json:"title" binding:"required"`
	AgeRating   string `json:"age_rating" binding:"required,oneof=P K T13 T16 T18"`
	DurationMin int    `json:"duration_min" binding:"required,gt=0"`
	Showtime    string `json:"showtime" binding:"required"`
	TotalSeats  int    `json:"total_seats" binding:"required,gt=0"`
	ImageUrl    string `json:"image_url" binding:"required"`
}

const ShowtimeLayout = "15:04 02-01-2006"

func parseShowtime(input string) (time.Time, error) {
	if t, err := time.Parse(ShowtimeLayout, input); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, input)
}

func (ctrl *Controller) GetAll(c *gin.Context) {
	movies, err := ctrl.Repo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch movies"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"movies": movies})
}

func (ctrl *Controller) GetCustomerMovies(c *gin.Context) {
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

	showtime, err := parseShowtime(input.Showtime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "showtime must be format 'hh:mm dd-mm-yyyy' or RFC3339"})
		return
	}

	movie := Movie{
		Title:          input.Title,
		AgeRating:      AgeRating(input.AgeRating),
		DurationMin:    input.DurationMin,
		Showtime:       showtime,
		TotalSeats:     input.TotalSeats,
		ImageUrl:       input.ImageUrl,
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
		parsed, err := parseShowtime(showtimeRaw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "showtime must be format 'hh:mm dd-mm-yyyy' or RFC3339"})
			return
		}
		input["showtime"] = parsed
	}

	currentMovie, err := ctrl.Repo.FindByID(id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find movie"})
		return
	}

	if totalSeatsRaw, ok := input["total_seats"]; ok {
		var newTotal int
		switch value := totalSeatsRaw.(type) {
		case float64:
			newTotal = int(value)
		case int:
			newTotal = value
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "total_seats must be a number"})
			return
		}

		bookedSeats := currentMovie.TotalSeats - currentMovie.AvailableSeats
		if newTotal < bookedSeats {
			c.JSON(http.StatusBadRequest, gin.H{"error": "total_seats cannot be lower than booked seats"})
			return
		}
		input["total_seats"] = newTotal
		input["available_seats"] = newTotal - bookedSeats
	}

	if duration, ok := input["duration_min"].(float64); ok {
		input["duration_min"] = int(duration)
	}
	if imageURLRaw, exists := input["image_url"]; exists {
		imageURL, ok := imageURLRaw.(string)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "image_url must be a string"})
			return
		}
		input["image_url"] = strings.TrimSpace(imageURL)
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

	if ctrl.TicketCounter != nil {
		count, countErr := ctrl.TicketCounter.CountByMovieID(id)
		if countErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify tickets"})
			return
		}
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete a movie that already has tickets"})
			return
		}
	}

	if err := ctrl.Repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete movie"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Movie deleted successfully"})
}
