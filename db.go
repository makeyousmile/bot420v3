package main

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"time"
)
import "go.mongodb.org/mongo-driver/mongo"

type HydraShop struct {
	Category   string
	Title      string
	Text       string
	Market     string
	Price      string
	UpdateTime time.Time
}

func ConnectToDb() (context.Context, *mongo.Client) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Print(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Print(err)
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Print(err)
	}

	return ctx, client
}
func WriteToDb(cityName string, shops []HydraShop) {
	ctx, client := ConnectToDb()
	collection := client.Database("bot420").Collection(cityName)

	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)

	for _, shop := range shops {
		res, err := collection.InsertOne(ctx, shop)
		if err != nil {
			log.Print(err)
		}
		id := res.InsertedID
		log.Print(id)
	}

}
