package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client wraps the mongo client and exposes the named database.
type Client struct {
	client *mongo.Client
	DB     *mongo.Database
}

// Connect establishes a MongoDB connection and verifies it with a ping.
func Connect(ctx context.Context, uri, dbName string) (*Client, error) {
	opts := options.Client().ApplyURI(uri).SetConnectTimeout(10 * time.Second)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("connecting to mongo: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		return nil, fmt.Errorf("pinging mongo: %w", err)
	}

	return &Client{
		client: client,
		DB:     client.Database(dbName),
	}, nil
}

// Disconnect gracefully closes the MongoDB connection.
func (c *Client) Disconnect(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}
