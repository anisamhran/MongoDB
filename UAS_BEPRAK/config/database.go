package config

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database
func ConnectDB(){
	log.Println("Connecting to MongoDB..")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://anisa88:An1s4mhraN159075@anisa-be.sjzgl.mongodb.net/?retryWrites=true&w=majority&appName=Anisa-BE"))
	if err != nil {
		log.Fatal("Error connecting MongoDB:", err)
	}

	DB = client.Database("unairsatu")
	log.Println("Connected to MongoDB")
}

func GetCollection(users string) *mongo.Collection {
	//connectDB
	if DB == nil {
		//log.Fatal("Database connection is not initialized")
		ConnectDB()
	}

	return DB.Collection(users)
}