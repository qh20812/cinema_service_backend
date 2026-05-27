package ticket

import (
	"net/http"
	"time"

	"main/src/auth"
	"main/src/movie"
	"main/src/user"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Controller struct {
	Repo      *Repository
	MovieRepo *movie.Repository
	UserRepo  *user.Repository
}

func NewController(repo *Repository, movieRepo *movie.Repository, userRepo *user.Repository) *Controller {
	return &Controller{Repo: repo, MovieRepo: movieRepo, UserRepo: userRepo}
}

type CreateTicketInput struct {
	MovieID           string `json:"movie_id" binding:"required"`
	SeatNumber        int    `json:"seat_number" binding:"required,gt=0"`
	GuardianConfirmed bool   `json:"guardian_confirmed"`
}

func (ctrl *Controller) Create(c *gin.Context) {
	var input CreateTicketInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	movieID, err := bson.ObjectIDFromHex(input.MovieID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie_id"})
		return
	}

	userIDValue, exists := c.Get(auth.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDstr, ok := userIDValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID người dùng không hợp lệ"})
		return
	}

	userID, err := bson.ObjectIDFromHex(userIDstr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID người dùng không hợp lệ"})
		return
	}

	usr := ctrl.UserRepo.FindByID(userID)
	if usr == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy người dùng"})
		return
	}

	age, err := usr.Age()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tuổi không hợp lệ"})
		return
	}

	mv, err := ctrl.MovieRepo.FindByID(movieID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy phim"})
		return
	}

	if !mv.CanBook(1) {
		c.JSON(http.StatusConflict, gin.H{"error": "Rạp đã đầy"})
		return
	}

	if !mv.Showtime.After(time.Now().UTC()) {
		c.JSON(http.StatusConflict, gin.H{"error": "Suất chiếu đã qua"})
		return
	}

	if existing, err := ctrl.Repo.FindByMovieAndSeat(movieID, input.SeatNumber); err == nil && existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Ghế đã được đặt"})
		return
	}

	switch mv.AgeRating {
	case movie.AgeRatingP:
		// no limit
	case movie.AgeRatingK:
		if !input.GuardianConfirmed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Phim loại K cần xác nhận người giám hộ"})
			return
		}
	case movie.AgeRatingT13:
		if age < 13 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cấm khán giả dưới 13 tuổi"})
			return
		}
	case movie.AgeRatingT16:
		if age < 16 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cấm khán giả dưới 16 tuổi"})
			return
		}
	case movie.AgeRatingT18:
		if age < 18 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cấm khán giả dưới 18 tuổi"})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Xếp hạng tuổi không hợp lệ"})
		return
	}

	if err := ctrl.MovieRepo.ReserveSeats(movieID, 1); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	ticket := Ticket{
		UserID:            userID,
		MovieID:           movieID,
		SeatNumber:        input.SeatNumber,
		GuardianConfirmed: input.GuardianConfirmed,
		Quantity:          1,
		AgeRating:         string(mv.AgeRating),
	}

	if err := ctrl.Repo.Create(&ticket); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo vé"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"ticket": ticket})
}

func (ctrl *Controller) GetByUser(c *gin.Context) {
	userIDValue, exists := c.Get(auth.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDstr, ok := userIDValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID người dùng không hợp lệ"})
		return
	}
	userID, err := bson.ObjectIDFromHex(userIDstr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ID người dùng không hợp lệ"})
		return
	}

	tickets, err := ctrl.Repo.FindByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy vé"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tickets": tickets})
}
