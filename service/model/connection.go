package model

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func PrintJson(rw http.ResponseWriter, r *http.Request) {
	_, cancel := context.WithTimeout(context.Background(), 10000*time.Second)
	defer cancel()
	collection := Client.Database("crud").Collection("patients")

	// Read JSON file
	jsonData, err := ioutil.ReadFile("records.json")
	if err != nil {
		log.Fatal(err)
	}

	var patients []Patient
	err = json.Unmarshal(jsonData, &patients)
	if err != nil {
		log.Fatal(err)
	}

	var insertModels []mongo.WriteModel

	// Iterate through the patients and create an insertOne model for each patient
	for _, patient := range patients {
		insertModel := mongo.NewInsertOneModel().SetDocument(patient)
		insertModels = append(insertModels, insertModel)
	}

	// Bulk insert the data into the collection
	bulkOptions := options.BulkWrite().SetOrdered(false) // SetOrdered(false) allows unordered bulk insert
	bulkResult, err := collection.BulkWrite(context.Background(), insertModels, bulkOptions)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("Inserted  documents into MongoDB!\n", bulkResult.InsertedCount)

}
