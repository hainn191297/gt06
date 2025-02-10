package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoDBModel defines an interface for MongoDB CRUD operations
type MongoDBModel interface {
	Insert(ctx context.Context, collectionName string, data bson.M) (*mongo.InsertOneResult, error)
	Update(ctx context.Context, collectionName string, filter bson.M, update bson.M) (*mongo.UpdateResult, error)
	Get(ctx context.Context, collectionName string, filter bson.M) (bson.M, error)
	Delete(ctx context.Context, collectionName string, filter bson.M) (*mongo.DeleteResult, error)
}

// mongoDBModel is the implementation of MongoDBModel
type mongoDBModel struct {
	db *mongo.Database
}

// TODO: Split into individual interface methods for each table to improve maintainability and clarity for business logic.
type CONCOXLoginInfoContent struct {
	TerminalID       string `bson:"terminal_id"`
	ModelCode        string `bson:"model_code"`
	TimeZoneLanguage string `bson:"time_zone_language"`
}

type CONCOXLocationInfoContent struct {
	DateTime            time.Time `bson:"date_time"`
	GPSSatellites       int       `bson:"gps_satellites"`
	Latitude            float64   `bson:"latitude"`
	Longitude           float64   `bson:"longitude"`
	Speed               int       `bson:"speed"`
	CourseStatus        int       `bson:"course_status"`
	MCC                 int       `bson:"mcc"`
	MNC                 int       `bson:"mnc"`
	LAC                 int       `bson:"lac"`
	CellID              int       `bson:"cell_id"`
	ACCStatus           int       `bson:"acc_status"`
	UploadMode          int       `bson:"upload_mode"`
	GPSRealTimeReupload int       `bson:"gps_real_time_reupload"`
	Mileage             int       `bson:"mileage"`
	SignalStrength      int       `bson:"signal_strength"`
	Altitude            int       `bson:"altitude"`
	BatteryLevel        int       `bson:"battery_level"`
	NetworkType         int       `bson:"network_type"`
	Temperature         int       `bson:"temperature"`
}

// NewMongoDBModel initializes a new MongoDBModel instance
func NewMongoDBModel(client *mongo.Client, dbName string) MongoDBModel {
	return &mongoDBModel{db: client.Database(dbName)}
}

// Insert inserts a document into the specified collection
func (m *mongoDBModel) Insert(ctx context.Context, collectionName string, data bson.M) (*mongo.InsertOneResult, error) {
	collection := m.db.Collection(collectionName)
	return collection.InsertOne(ctx, data)
}

// Update updates documents in the specified collection based on a filter
func (m *mongoDBModel) Update(ctx context.Context, collectionName string, filter bson.M, update bson.M) (*mongo.UpdateResult, error) {
	collection := m.db.Collection(collectionName)
	return collection.UpdateOne(ctx, filter, bson.M{"$set": update})
}

// Get retrieves a single document based on a filter
func (m *mongoDBModel) Get(ctx context.Context, collectionName string, filter bson.M) (bson.M, error) {
	collection := m.db.Collection(collectionName)
	var result bson.M
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return result, err
}

// Delete removes documents from the specified collection based on a filter
func (m *mongoDBModel) Delete(ctx context.Context, collectionName string, filter bson.M) (*mongo.DeleteResult, error) {
	collection := m.db.Collection(collectionName)
	return collection.DeleteOne(ctx, filter)
}
