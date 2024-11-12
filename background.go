package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// deleteExpiredURLs removes URLs that have expired based on their expiration date
func deleteExpiredURLs() {
	filter := bson.M{"expiration_date": bson.M{"$lt": time.Now()}}
	result, err := urlCollection.DeleteMany(context.Background(), filter)
	if err != nil {
		log.Println("Error deleting expired URLs:", err)
		return
	}
	log.Printf("Deleted %d expired URLs\n", result.DeletedCount)
}

func startBackgroundTasks() {
	go func() {
		for {
			deleteExpiredURLs()
			time.Sleep(24 * time.Hour) // Run once every 24 hours
		}
	}()
}
