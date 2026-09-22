package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/ares/dp-vc-webApp/configs/env"
)

var (
	client     *mongo.Client
	database   *mongo.Database
	timeoutCtx = 10 * time.Second
)

// Init initializes MongoDB connection
func Init(config *env.Config) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Set client options
	clientOptions := options.Client().
		ApplyURI(config.MongoURI).
		SetServerSelectionTimeout(10 * time.Second).
		SetConnectTimeout(10 * time.Second)

	// Connect to MongoDB
	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Get database
	database = client.Database(config.MongoDBName)

	return client, nil
}

// Close closes the MongoDB connection
func Close() {
	if client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = client.Disconnect(ctx)
	}
}

// Disconnect disconnects from MongoDB (alias for Close)
func Disconnect(ctx context.Context) {
	if client != nil {
		_ = client.Disconnect(ctx)
	}
}

// Ping verifies the connection is still alive
func Ping(ctx context.Context) error {
	if client == nil {
		return fmt.Errorf("MongoDB client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, timeoutCtx)
	defer cancel()

	return client.Ping(ctx, readpref.Primary())
}

// GetDatabase returns the database instance
func GetDatabase() *mongo.Database {
	return database
}

// GetClient returns the MongoDB client
func GetClient() *mongo.Client {
	return client
}

// GetCollection returns a collection from the database
func GetCollection(name string) *mongo.Collection {
	return database.Collection(name)
}

// FindOne finds a single document
func FindOne(ctx context.Context, collection string, filter bson.M, result interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, timeoutCtx)
	defer cancel()

	coll := GetCollection(collection)
	return coll.FindOne(ctx, filter).Decode(result)
}

// FindMany finds multiple documents
func FindMany(ctx context.Context, collection string, filter bson.M, results interface{}, opts ...*options.FindOptions) error {
	ctx, cancel := context.WithTimeout(ctx, timeoutCtx)
	defer cancel()

	coll := GetCollection(collection)
	cursor, err := coll.Find(ctx, filter, opts...)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	return cursor.All(ctx, results)
}

// InsertOne inserts a single document
func InsertOne(ctx context.Context, collection string, document interface{}) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, timeoutCtx)
	defer cancel()

	coll := GetCollection(collection)
	result, err := coll.InsertOne(ctx, document)
	if err != nil {
		return nil, err
	}
	return result.InsertedID, nil
}

// UpdateOne updates a single document
func UpdateOne(ctx context.Context, collection string, filter bson.M, update interface{}) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeoutCtx)
	defer cancel()

	coll := GetCollection(collection)
	result, err := coll.UpdateOne(ctx, filter, bson.M{"$set": update})
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, nil
}

// DeleteOne deletes a single document
func DeleteOne(ctx context.Context, collection string, filter bson.M) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeoutCtx)
	defer cancel()

	coll := GetCollection(collection)
	result, err := coll.DeleteOne(ctx, filter)
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

// CountDocuments counts documents matching the filter
func CountDocuments(ctx context.Context, collection string, filter bson.M) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, timeoutCtx)
	defer cancel()

	coll := GetCollection(collection)
	return coll.CountDocuments(ctx, filter)
}
