package utils

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/sirupsen/logrus"
)

// ArangoDB 环境变量键常量
const (
	ArangoURL      = "ARANGO_URL"
	ArangoDB       = "ARANGO_DB"
	ArangoUsername = "ARANGO_USERNAME"
	ArangoPassword = "ARANGO_PASSWORD"
)

// ArangoDBClient represents a connection to ArangoDB
type ArangoDBClient struct {
	Client        driver.Client
	DB            driver.Database
	OsintGraph    driver.Graph
	GraphSchemaMu sync.Mutex
}

// NewArangoDBClient creates a new ArangoDB client and connects to the database
func NewArangoDBClient() (*ArangoDBClient, error) {
	// Get ArangoDB connection details from environment variables
	arangoDBURL := os.Getenv(ArangoURL)
	if arangoDBURL == "" {
		logrus.Fatal("missing environment variable ARANGO_URL")
	}

	databaseName := os.Getenv(ArangoDB)
	if databaseName == "" {
		logrus.Fatal("missing environment variable ARANGO_DB")
	}

	username := os.Getenv(ArangoUsername)
	if username == "" {
		logrus.Fatal("missing environment variable ARANGO_USERNAME")
	}

	password := os.Getenv(ArangoPassword)
	if password == "" {
		logrus.Fatal("missing environment variable ARANGO_PASSWORD")
	}

	// Connect to ArangoDB
	logrus.Infof("Connecting to ArangoDB at %s using (%s, %s)", arangoDBURL, username, password)
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{arangoDBURL},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %v", err)
	}

	client, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(username, password),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	// Create or get database
	db, err := CreateOrGetDatabase(client, databaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %v", err)
	}

	ctx := context.Background()
	osintGraph, err := CreateOrGetGraph(db, ctx, "osint", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get osint graph: %v", err)
	}

	arangoClient := ArangoDBClient{
		Client:        client,
		DB:            db,
		OsintGraph:    osintGraph,
		GraphSchemaMu: sync.Mutex{},
	}

	return &arangoClient, nil
}

func (adb *ArangoDBClient) ParseDocID(id string) (collectionName string, key string, err error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid document ID format: %s", id)
	}
	return parts[0], parts[1], nil
}

func (c *ArangoDBClient) GetCreateCollection(ctx context.Context, name string, options driver.CreateVertexCollectionOptions) (driver.Collection, error) {
	exists, err := c.OsintGraph.VertexCollectionExists(ctx, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return c.OsintGraph.VertexCollection(ctx, name)
	}

	c.GraphSchemaMu.Lock()
	defer c.GraphSchemaMu.Unlock()

	exists, err = c.OsintGraph.VertexCollectionExists(ctx, name)
	if err != nil {
		return nil, err
	}

	if exists {
		return c.OsintGraph.VertexCollection(ctx, name)
	}

	return c.OsintGraph.CreateVertexCollectionWithOptions(ctx, name, options)
}

func (c *ArangoDBClient) GetCreateEdgeCollection(ctx context.Context, name string, constraints driver.VertexConstraints, options driver.CreateEdgeCollectionOptions) (driver.Collection, error) {
	exists, err := c.OsintGraph.EdgeCollectionExists(ctx, name)
	if err != nil {
		return nil, err
	}
	if exists {
		collection, _, err := c.OsintGraph.EdgeCollection(ctx, name)
		return collection, err
	}

	c.GraphSchemaMu.Lock()
	defer c.GraphSchemaMu.Unlock()

	exists, err = c.OsintGraph.EdgeCollectionExists(ctx, name)
	if err != nil {
		return nil, err
	}

	if exists {
		collection, _, err := c.OsintGraph.EdgeCollection(ctx, name)
		return collection, err
	}

	return c.OsintGraph.CreateEdgeCollectionWithOptions(ctx, name, constraints, options)
}

func CreateOrGetGraph(db driver.Database, ctx context.Context, name string, options *driver.CreateGraphOptions) (driver.Graph, error) {
	exists, err := db.GraphExists(ctx, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return db.Graph(ctx, name)
	}

	graph, err := db.CreateGraphV2(ctx, name, options)
	return graph, err
}

func CreateOrGetDatabase(client driver.Client, dbName string) (driver.Database, error) {
	ctx := context.Background()

	// Check if database exists
	exists, err := client.DatabaseExists(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to check database existence: %v", err)
	}

	if exists {
		fmt.Printf("📁 Using existing database: %s\n", dbName)
		return client.Database(ctx, dbName)
	}

	// Create new database
	fmt.Printf("🆕 Creating new database: %s\n", dbName)
	db, err := client.CreateDatabase(ctx, dbName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %v", err)
	}

	return db, nil
}
