package movie

import (
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AgeRating string

const (
	AgeRatingP   AgeRating = "P"
	AgeRatingK   AgeRating = "K"
	AgeRatingT13 AgeRating = "T13"
	AgeRatingT16 AgeRating = "T16"
	AgeRatingT18 AgeRating = "T18"
)

func ValidAgeRating(r AgeRating) bool {
	switch strings.ToUpper(string(r)) {
	case "P", "K", "T13", "T16", "T18":
		return true
	default:
		return false
	}
}

type Movie struct {
	ID             bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title          string        `json:"title" bson:"title"`
	AgeRating      AgeRating     `json:"age_rating" bson:"age_rating"`
	DurationMin    int           `json:"duration_min" bson:"duration_min"` // phút
	Showtime       time.Time     `json:"showtime" bson:"showtime"`
	TotalSeats     int           `json:"total_seats" bson:"total_seats"`
	AvailableSeats int           `json:"available_seats" bson:"available_seats"`
	CreatedAt      time.Time     `json:"created_at,omitempty" bson:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at,omitempty" bson:"updated_at"`
}

func (m *Movie) PrepareForCreate() {
	if m.ID.IsZero() {
		m.ID = bson.NewObjectID()
	}
	m.Title = strings.TrimSpace(m.Title)
	m.AgeRating = AgeRating(strings.TrimSpace(strings.ToUpper(string(m.AgeRating))))
	if !ValidAgeRating(m.AgeRating) {
		m.AgeRating = AgeRatingP
	}
	if m.TotalSeats < 0 {
		m.TotalSeats = 0
	}
	if m.AvailableSeats < 0 {
		m.AvailableSeats = 0
	}
	if m.AvailableSeats > m.TotalSeats {
		m.AvailableSeats = m.TotalSeats
	}
	m.CreatedAt = time.Now().UTC()
	m.UpdatedAt = m.CreatedAt
}

func (m *Movie) PrepareForUpdate() {
	m.UpdatedAt = time.Now().UTC()
	m.Title = strings.TrimSpace(m.Title)
	if m.AgeRating != "" {
		m.AgeRating = AgeRating(strings.TrimSpace(strings.ToUpper(string(m.AgeRating))))
		if !ValidAgeRating(m.AgeRating) {
			m.AgeRating = AgeRatingP
		}
	}
	if m.AvailableSeats > m.TotalSeats {
		m.AvailableSeats = m.TotalSeats
	}
}

func (m *Movie) CanBook(quantity int) bool {
	return quantity > 0 && m.AvailableSeats >= quantity
}

func (m *Movie) BookSeats(quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}
	if m.AvailableSeats <= 0 {
		return errors.New("rạp đã đầy")
	}
	if quantity > m.AvailableSeats {
		return errors.New("không đủ chỗ, rạp gần đầy")
	}
	m.AvailableSeats -= quantity
	m.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *Movie) ReleaseSeats(quantity int) {
	if quantity <= 0 {
		return
	}
	m.AvailableSeats += quantity
	if m.AvailableSeats > m.TotalSeats {
		m.AvailableSeats = m.TotalSeats
	}
	m.UpdatedAt = time.Now().UTC()
}

func (m *Movie) Validate() error {
	if strings.TrimSpace(m.Title) == "" {
		return errors.New("title is required")
	}
	if !ValidAgeRating(m.AgeRating) {
		return errors.New("invalid age_rating")
	}
	if m.DurationMin <= 0 {
		return errors.New("duration must be greater than zero")
	}
	if m.TotalSeats <= 0 {
		return errors.New("total_seats must be greater than zero")
	}
	if m.AvailableSeats < 0 || m.AvailableSeats > m.TotalSeats {
		return errors.New("available_seats must be between 0 and total_seats")
	}
	if m.Showtime.IsZero() {
		return errors.New("showtime is required")
	}
	return nil
}