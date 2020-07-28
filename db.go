package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
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
	ctx, _ := context.WithTimeout(context.Background(), 50*time.Second)
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

	client.Database("bot420").Collection(cityName).Drop(ctx)
	collectionBot420 := client.Database("bot420").Collection(cityName)
	for _, shop := range shops {
		collectionBot420.InsertOne(ctx, shop)
	}

}

func Exist(CollectionName string, shop HydraShop) bool {
	ctx, client := ConnectToDb()
	collectionBot420 := client.Database("bot420").Collection(CollectionName)

	cur, err := collectionBot420.Find(ctx, bson.M{
		"category": shop.Category,
		"title":    shop.Title,
		"market":   shop.Market,
		"price":    shop.Price,
	})

	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(context.Background())
	if cur.TryNext(ctx) {
		log.Print("faund in base")
		return true
	}

	return false
}
