//go:build integration

package db_test

import (
	"context"
	"os"
	"testing"
	"time"

	"expensify/internal/db"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// testDB returns a connected *mongo.Database for integration tests.
// Set TEST_MONGO_URI and TEST_DB_NAME (or rely on defaults) before running.
func testDB(t *testing.T) *mongo.Database {
	t.Helper()

	uri := os.Getenv("TEST_MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	dbName := os.Getenv("TEST_DB_NAME")
	if dbName == "" {
		dbName = "expensify_test"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := db.Connect(ctx, uri, dbName)
	if err != nil {
		t.Fatalf("connecting to test mongo: %v", err)
	}

	t.Cleanup(func() {
		dropCtx, dropCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer dropCancel()
		// Drop the test database after every test run to keep tests hermetic.
		_ = client.DB.Drop(dropCtx)
		_ = client.Disconnect(context.Background())
	})

	return client.DB
}

// mustConnect is the same as testDB but returns the full Client (for index tests).
func mustConnect(t *testing.T) *mongo.Client {
	t.Helper()
	uri := os.Getenv("TEST_MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("raw mongo connect: %v", err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	return client
}
