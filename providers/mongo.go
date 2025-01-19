package providers

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/knadh/koanf/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ koanf.Provider = (*MongoDB)(nil)

// MongoDB is an implementation of koanf.Provider that reads/loads configuration
// stored in a Document in MongoDB.
type MongoDB struct {
	client       *mongo.Client
	database     string
	collection   string
	documentID   string
	watched      atomic.Uint32
	changeStream *mongo.ChangeStream
}

// MongoDBProvider initializes and returns a new MongoDB instance.
func MongoDBProvider(client *mongo.Client, db, collection, docId string) *MongoDB {
	if client == nil {
		panic("client is nil")
	}
	return &MongoDB{
		client:       client,
		database:     db,
		collection:   collection,
		documentID:   docId,
		watched:      atomic.Uint32{},
		changeStream: nil,
	}
}

// ReadBytes reads a MongoDB document and returns the data as bytes. The bytes
// will always be encoded as BSON.
func (m *MongoDB) ReadBytes() ([]byte, error) {
	collection := m.client.Database(m.database).Collection(m.collection)
	filter := bson.D{{"_id", m.documentID}}

	var result bson.M
	err := collection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	data, err := bson.Marshal(result)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Read is not supported by MongoDB and will always return an error.
func (m *MongoDB) Read() (map[string]interface{}, error) {
	return nil, fmt.Errorf("%T does not support Read()", m)
}

// Watch sets up a change stream to monitor a specific MongoDB document for updates
// and invokes the callback on changes.
func (m *MongoDB) Watch(cb func(event interface{}, err error)) error {
	activated := m.watched.CompareAndSwap(0, 1)
	if !activated {
		return fmt.Errorf("%T.Watch may only be invoked once", m)
	}

	collection := m.client.Database(m.database).Collection(m.collection)

	// We only care about changes to the specific document that holds the configuration
	pipeline := mongo.Pipeline{
		{{"$match", bson.D{{"documentKey._id", m.documentID}}}},
	}

	changeStream, err := collection.Watch(context.Background(), pipeline)
	if err != nil {
		return fmt.Errorf("failed to initialize change stream: %w", err)
	}
	m.changeStream = changeStream

	// Listen to the change stream and react by calling the callback when a change
	// is detected
	go func() {
		for m.changeStream.Next(context.Background()) {
			var event bson.M
			if err := m.changeStream.Decode(&event); err != nil {
				cb(nil, err)
			}

			operation, ok := event["operationType"].(string)
			if !ok {
				cb(nil, fmt.Errorf("failed to parse operation type: %v", event))
				continue
			}

			// If the document holding configuration is deleted we likely don't want
			// the application to reload its configuration since it is doomed to fail
			if operation == "delete" {
				continue
			}

			cb(event, nil)
		}
	}()

	return nil
}

// Close terminates the MongoDB change stream if active and returns any encountered
// error during closure.
func (m *MongoDB) Close() error {
	if m.watched.Load() == 1 && m.changeStream != nil {
		return m.changeStream.Close(context.Background())
	}
	return nil
}
