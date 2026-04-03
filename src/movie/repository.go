package movie

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct {
	Collection *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	col := db.Collection("movies")
	if err := CreateMovieIndexes(col); err != nil {
		panic(err)
	}
	return &Repository{Collection: col}
}

func CreateMovieIndexes(col *mongo.Collection) error {
	_, err := col.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys:    bson.D{{Key: "title", Value: 1}},
		Options: options.Index().SetUnique(false),
	})
	if err != nil {
		return err
	}
	_, err = col.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys:    bson.D{{Key: "showtime", Value: 1}},
		Options: options.Index().SetUnique(false),
	})
	return err
}

func (r *Repository) Create(movie *Movie) error {
	movie.PrepareForCreate()
	if err := movie.Validate(); err != nil {
		return err
	}
	_, err := r.Collection.InsertOne(context.TODO(), movie)
	return err
}

func (r *Repository) FindByID(id bson.ObjectID) (*Movie, error) {
	var movie Movie
	err := r.Collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&movie)
	if err != nil {
		return nil, err
	}
	return &movie, nil
}

func (r *Repository) FindAll() ([]Movie, error) {
	cursor, err := r.Collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var movies []Movie
	if err := cursor.All(context.TODO(), &movies); err != nil {
		return nil, err
	}
	return movies, nil
}

func (r *Repository) Update(movieID bson.ObjectID, update bson.M) error {
	if len(update) == 0 {
		return nil
	}
	update["updated_at"] = time.Now().UTC()
	_, err := r.Collection.UpdateByID(context.TODO(), movieID, bson.M{"$set": update})
	return err
}

func (r *Repository) Delete(movieID bson.ObjectID) error {
	_, err := r.Collection.DeleteOne(context.TODO(), bson.M{"_id": movieID})
	return err
}

func (r *Repository) ReserveSeats(movieID bson.ObjectID, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("quantity must be greater than zero")
	}

	movie, err := r.FindByID(movieID)
	if err != nil {
		return err
	}
	if !movie.CanBook(quantity) {
		return fmt.Errorf("không đủ chỗ, rạp đã đầy hoặc số lượng yêu cầu vượt quá")
	}
	if err := movie.BookSeats(quantity); err != nil {
		return err
	}

	_, err = r.Collection.UpdateByID(context.TODO(), movieID, bson.M{"$set": bson.M{"available_seats": movie.AvailableSeats, "updated_at": movie.UpdatedAt}})
	return err
}
