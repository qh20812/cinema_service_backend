package ticket

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Ticket struct {
	ID         bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID     bson.ObjectID `json:"user_id" bson:"user_id"`
	MovieID    bson.ObjectID `json:"movie_id" bson:"movie_id"`
	SeatNumber int           `json:"seat_number" bson:"seat_number"`
	GuardianConfirmed bool   `json:"guardian_confirmed" bson:"guardian_confirmed"`
	Quantity   int           `json:"quantity" bson:"quantity"`
	AgeRating  string        `json:"age_rating" bson:"age_rating"` // snapshot độ tuổi phim khi mua
	CreatedAt  time.Time     `json:"created_at,omitempty" bson:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at,omitempty" bson:"updated_at"`
}

func (t *Ticket) PrepareForCreate() {
	if t.ID.IsZero() {
		t.ID = bson.NewObjectID()
	}
	if t.Quantity < 1 {
		t.Quantity = 1
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = t.CreatedAt
	}
}

func (t *Ticket) Validate() error {
	if t.UserID.IsZero() {
		return errors.New("user_id is required")
	}
	if t.MovieID.IsZero() {
		return errors.New("movie_id is required")
	}
	if t.SeatNumber <= 0 {
		return errors.New("seat_number must be greater than zero")
	}
	// guardian_confirmed is validated by controller for K movies; here we only keep the field.
	if t.Quantity <= 0 {
		return errors.New("quantity must be greater than zero")
	}
	return nil
}

func (t *Ticket) SetAgeRating(r string) {
	t.AgeRating = r
}

func (t *Ticket) UpdateQuantity(delta int) error {
	if t.Quantity+delta <= 0 {
		return errors.New("quantity must be at least 1")
	}
	t.Quantity += delta
	t.UpdatedAt = time.Now().UTC()
	return nil
}
