package model

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type Patientid struct {
	ID string `json:"patientid" bson:"patientid"`
}

func PrintJson() {

	ctx, cancel := context.WithTimeout(context.Background(), 10000*time.Second)
	defer cancel()
	collection := Client.Database("crud").Collection("patients")
	cursor, _ := collection.Find(ctx, bson.M{})
	var patients []Patientid
	fmt.Print("Fetched data")
	i := 0
	for cursor.Next(ctx) {
		var patient Patientid

		if err := cursor.Decode(&patient); err != nil {
			log.Fatal(err)
		}

		patients = append(patients, patient)
		i++
		if i == 100000 {
			break
		}
	}
	fmt.Print("Hii")

	jsonData, err := json.Marshal(patients)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("ids.json", jsonData, 0644)
	if err != nil {
		log.Fatal(err)
	}

}
