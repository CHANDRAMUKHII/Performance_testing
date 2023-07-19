package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"google.golang.org/grpc"

	pb "client/pb"
)

type PatientID struct {
	ID string `json:"patientid"`
}

func main() {
	conn, err := grpc.Dial("patient-service:5002", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewMongoDBServiceClient(conn)

	jsonData, err := ioutil.ReadFile("output.json")
	if err != nil {
		log.Fatalf("failed to read JSON file: %v", err)
	}

	var patientIDs []PatientID
	err = json.Unmarshal(jsonData, &patientIDs)
	if err != nil {
		log.Fatalf("failed to unmarshal JSON data: %v", err)
	}

	var average time.Duration
	for i := 0; i < 1; i++ {
		startTime := time.Now()

		batchSize := 10
		for i := 0; i < len(patientIDs); i += batchSize {
			end := i + batchSize
			if end > len(patientIDs) {
				end = len(patientIDs)
			}
			if i == 10000 {
				break
			}
			batchRequest := &pb.BatchFetchRequest{}
			for _, patientID := range patientIDs[i:end] {
				batchRequest.PatientIds = append(batchRequest.PatientIds, patientID.ID)
			}

			_, err = client.FetchDataBatchFromMongoDB(context.Background(), batchRequest)
			if err != nil {
				log.Fatalf("request failed: %v", err)
			}

			// for _, fetchedData := range resp.FetchedData {
			// 	printPatientDetails(fetchedData)
			// }

		}

		elapsedTime := time.Since(startTime)
		average += elapsedTime

		fmt.Println("Total time taken:", elapsedTime)
	}

	average /= 1
	fmt.Print("Average time taken : ", average)
}

func printPatientDetails(patient *pb.Patient) {
	fmt.Println("Patient ID:", patient.PatientID)
	fmt.Println("First Name:", patient.FirstName)
	fmt.Println("Last Name:", patient.LastName)
	fmt.Println("Date of Birth:", patient.DateofBirth)
	fmt.Println("Gender:", patient.Gender)
	fmt.Println("Contact Number:", patient.ContactNumber)
	fmt.Println("Medical History:", patient.MedicalHistory)
	fmt.Println("Date of Discharge:", patient.DateOfDischarge)
	fmt.Println("------------------------------------")
}
