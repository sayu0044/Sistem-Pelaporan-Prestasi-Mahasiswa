package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client
var MongoDB *mongo.Database

// ConnectMongoDB menghubungkan ke database MongoDB
func ConnectMongoDB(mongoURI, dbName string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Gagal menghubungkan ke MongoDB:", err)
	}

	// Test connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Gagal ping MongoDB:", err)
	}

	MongoClient = client
	MongoDB = client.Database(dbName)

	log.Println("MongoDB berhasil terhubung")
}

// DisconnectMongoDB menutup koneksi MongoDB
func DisconnectMongoDB() {
	if MongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := MongoClient.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting MongoDB: %v", err)
		}
		log.Println("MongoDB disconnected")
	}
}

