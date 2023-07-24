package model

import (
	"context"
	"log"

	"fmt"

	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Patient struct {
	PatientID       string `json:"patientID" bson:"patientID"`
	FirstName       string `json:"firstName" bson:"firstName"`
	LastName        string `json:"lastName" bson:"lastName"`
	DateofBirth     string `json:"dateOfBirth" bson:"dateOfBirth"`
	Gender          string `json:"gender" bson:"gender"`
	ContactNumber   string `json:"contactNumber" bson:"contactNumber"`
	MedicalHistory  string `json:"medicalHistory" bson:"medicalHistory"`
	DateOfDischarge string `json:"dateOfDischarge" bson:"dateOfDischarge"`
}

var Client *mongo.Client

func Connection() (*mongo.Client, error) {
	var err error
	const uri = "mongodb://localhost:27017"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	Client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println("Error in connecting to mongodb", err)
		return nil, err
	}

	err = Client.Ping(ctx, nil)
	if err != nil {
		Client.Disconnect(ctx)
		return nil, err
	}

	fmt.Println("Connected to MongoDB successfully!")
	return Client, nil
}

func DisconnectDB(client *mongo.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client.Disconnect(ctx)
}

func FetchData(id string) (Patient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	collection := Client.Database("crud").Collection("patients")
	var patient Patient
	err := collection.FindOne(ctx, bson.M{"patientID": id}).Decode(&patient)
	return patient, err
}

// grpc methods

type MongoDBModel struct {
	Client *mongo.Client
}

func (m *MongoDBModel) FetchData(ctx context.Context, patientID string) Patient {
	collection := m.Client.Database("crud").Collection("patients")
	var patient Patient

	err := collection.FindOne(ctx, bson.M{"patientID": patientID}).Decode(&patient)
	if err != nil {
		log.Printf("Failed to fetch data for patient ID %s: %v", patientID, err)
		return patient
	}

	return patient
}
