package lib

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConnect struct {
	client *mongo.Client
}

var Client *mongo.Client

func NewMongo() *MongoConnect {
	godotenv.Load()

	mongoDb := &MongoConnect{}
	mongoDb.Connect()
	return mongoDb
}
func (c *MongoConnect) Connect() {
	if Client != nil {
		c.client = Client
		return
	}
	uri := os.Getenv("DB_URL")
	con, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	c.client = con
	Client = con
}
func (c *MongoConnect) Close() {

	c.client.Disconnect(context.TODO())
	Client = nil

}

func (c *MongoConnect) Collection(name string) *mongo.Collection {
	var collection *mongo.Collection
	fmt.Println("c.client", c.client)
	fmt.Println("Client", Client)
	if c == nil {
		c = NewMongo()
	}
	if c.client == nil {
		c.Connect()
	}
	collection = c.client.Database(os.Getenv("DB_NAME")).Collection(name)
	return collection
}
