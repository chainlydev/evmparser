package lib

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConnect struct {
	client *mongo.Client
}

var Client *mongo.Client

func NewMongo() *MongoConnect {
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
	con, _ := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))

	c.client = con
	Client = con
}

func (c *MongoConnect) Collection(name string) *mongo.Collection {
	var collection *mongo.Collection
	defer func() {
		if r := recover(); r != nil {
			err := c.client.Connect(context.TODO())
			if err != nil {
				c.Connect()
			}
			collection = c.client.Database(os.Getenv("DB_NAME")).Collection(name)
		}
	}()
	collection = c.client.Database(os.Getenv("DB_NAME")).Collection(name)
	return collection
}
