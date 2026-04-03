package user

import (
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID          bson.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Email       string        `json:"email" bson:"email"`
	Password    string        `json:"password,omitempty" bson:"password"`
	DateOfBirth string        `json:"date_of_birth,omitempty" bson:"date_of_birth"`
	Role        string        `json:"role" bson:"role"` // admin, customer
	CreatedAt   time.Time     `json:"created_at,omitempty" bson:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at,omitempty" bson:"updated_at"`
}

func (u *User) PrepareForCreate() {
	if u.ID.IsZero() {
		u.ID = bson.NewObjectID()
	}
	u.CreatedAt = time.Now().UTC()
	u.UpdatedAt = u.CreatedAt
	u.Email = strings.TrimSpace(strings.ToLower(u.Email))
	u.Role = strings.TrimSpace(strings.ToLower(u.Role))
	if u.Role == "" {
		u.Role = "customer"
	}
}

func (u *User) PrepareForUpdate() {
	u.UpdatedAt = time.Now().UTC()
	if u.Email != "" {
		u.Email = strings.TrimSpace(strings.ToLower(u.Email))
	}
	if u.Role != "" {
		u.Role = strings.TrimSpace(strings.ToLower(u.Role))
	}
}

func (u *User) Sanitize() {
	u.Password = ""
}

func (u *User) Age() (int, error) {
	if u.DateOfBirth == "" {
		return 0, errors.New("date_of_birth is empty")
	}
	//yyyy-mm-dd
	dob, err := time.Parse("2006-01-02", u.DateOfBirth)
	if err != nil {
		// try RFC3339 fallback
		dob, err = time.Parse(time.RFC3339, u.DateOfBirth)
		if err != nil {
			return 0, err
		}
	}
	now := time.Now().UTC()
	age := now.Year() - dob.Year()
	if now.Month() < dob.Month() || (now.Month() == dob.Month() && now.Day() < dob.Day()) {
		age--
	}
	return age, nil
}

func (u *User) Validate() error {
	if strings.TrimSpace(u.Email) == "" {
		return errors.New("email is required")
	}
	if strings.TrimSpace(u.Password) == "" {
		return errors.New("password is required")
	}
	if strings.TrimSpace(u.DateOfBirth) == "" {
		return errors.New("date_of_birth is required")
	}
	if u.Role != "" && u.Role != "admin" && u.Role != "customer" {
		return errors.New("role must be 'admin' or 'customer'")
	}
	return nil
}
