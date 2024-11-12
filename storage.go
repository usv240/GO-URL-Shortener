package main

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client
var urlCollection *mongo.Collection

type URLMapping struct {
	ShortCode      string    `bson:"short_code"`
	OriginalURL    string    `bson:"original_url"`
	CreatedAt      time.Time `bson:"created_at"`
	ExpirationDate time.Time `bson:"expiration_date,omitempty"`
}

// Initialize MongoDB
func initializeDatabase() error {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return err
	}

	mongoClient = client
	urlCollection = client.Database("url_shortener").Collection("urls")

	// Create unique indexes for short_code and original_url
	_, err = urlCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.M{"short_code": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	_, err = urlCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.M{"original_url": 1},
		Options: options.Index().SetUnique(true),
	})
	return err
}

// Retrieve a URL mapping by short code
func getURLMapping(shortCode string) (*URLMapping, error) {
	var mapping URLMapping
	err := urlCollection.FindOne(context.Background(), bson.M{"short_code": shortCode}).Decode(&mapping)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &mapping, err
}

// Retrieve a URL mapping by the original URL
func getURLMappingByOriginalURL(originalURL string) (*URLMapping, error) {
	var mapping URLMapping
	err := urlCollection.FindOne(context.Background(), bson.M{"original_url": originalURL}).Decode(&mapping)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &mapping, err
}

// Save a new URL mapping
func saveURLMapping(mapping URLMapping) error {
	_, err := urlCollection.InsertOne(context.Background(), mapping)
	if mongo.IsDuplicateKeyError(err) {
		return errors.New("duplicate key error")
	}
	return err
}
