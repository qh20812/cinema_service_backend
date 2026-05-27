package ticket

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct {
	Collection *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	col := db.Collection("tickets")
	if err := CreateTicketIndexes(col); err != nil {
		panic(err)
	}
	return &Repository{Collection: col}
}

func CreateTicketIndexes(col *mongo.Collection) error {
	_, err := col.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "movie_id", Value: 1}}},
		{Keys: bson.D{{Key: "movie_id", Value: 1}, {Key: "seat_number", Value: 1}}, Options: options.Index().SetUnique(true)},
	})
	return err
}

func (r *Repository) Create(ticket *Ticket) error {
	ticket.PrepareForCreate()
	if err := ticket.Validate(); err != nil {
		return err
	}
	_, err := r.Collection.InsertOne(context.TODO(), ticket)
	return err
}

func (r *Repository) FindByID(id bson.ObjectID) (*Ticket, error) {
	var t Ticket
	err := r.Collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) FindByUserID(userID bson.ObjectID) ([]Ticket, error) {
	cursor, err := r.Collection.Find(context.TODO(), bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var tickets []Ticket
	if err := cursor.All(context.TODO(), &tickets); err != nil {
		return nil, err
	}
	return tickets, nil
}

func (r *Repository) FindByMovieID(movieID bson.ObjectID) ([]Ticket, error) {
	cursor, err := r.Collection.Find(context.TODO(), bson.M{"movie_id": movieID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var tickets []Ticket
	if err := cursor.All(context.TODO(), &tickets); err != nil {
		return nil, err
	}
	return tickets, nil
}

func (r *Repository) CountByMovieID(movieID bson.ObjectID) (int64, error) {
	return r.Collection.CountDocuments(context.TODO(), bson.M{"movie_id": movieID})
}

func (r *Repository) FindByMovieAndSeat(movieID bson.ObjectID, seatNumber int) (*Ticket, error) {
	var t Ticket
	err := r.Collection.FindOne(context.TODO(), bson.M{"movie_id": movieID, "seat_number": seatNumber}).Decode(&t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) FindAll() ([]Ticket, error) {
	cursor, err := r.Collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var tickets []Ticket
	if err := cursor.All(context.TODO(), &tickets); err != nil {
		return nil, err
	}
	return tickets, nil
}

func (r *Repository) Delete(id bson.ObjectID) error {
	_, err := r.Collection.DeleteOne(context.TODO(), bson.M{"_id": id})
	return err
}
