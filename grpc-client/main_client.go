package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"

	pb "client/pb"
)

type PatientID struct {
	ID string `json:"patientid"`
}

type ResponseData struct {
	Data        []*pb.Patient `json:"data"`
	AverageTime time.Duration `json:"average_time"`
}

var startTime time.Time
var average time.Duration

func main() {
	http.HandleFunc("/", sendRequest)
	log.Print("Listening in port 5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}

func sendRequest(rw http.ResponseWriter, r *http.Request) {
	serviceURL := os.Getenv("SERVICE_URL")

	if serviceURL == "" {
		fmt.Println("SERVICE_URL environment variable is not set.")
		return
	}
	conn, err := grpc.Dial(serviceURL, grpc.WithInsecure())
	if err != nil {
		http.Error(rw, "Failed to connect to gRPC server", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	client := pb.NewMongoDBServiceClient(conn)

	jsonData, err := ioutil.ReadFile("output.json")
	if err != nil {
		http.Error(rw, "Failed to read JSON file", http.StatusInternalServerError)
		return
	}

	var patientIDs []PatientID
	err = json.Unmarshal(jsonData, &patientIDs)
	if err != nil {
		http.Error(rw, "Failed to unmarshal JSON data", http.StatusInternalServerError)
		return
	}

	var fetchedPatients []*pb.Patient
	batchsize := 100
	for i := 0; i < len(patientIDs); i += batchsize {
		end := i + batchsize
		if end > len(patientIDs) {
			end = len(patientIDs)
		}
		if i > 10000 {
			break
		}
		batchRequest := &pb.BatchFetchRequest{}
		for _, patientID := range patientIDs[i:end] {
			batchRequest.PatientIds = append(batchRequest.PatientIds, patientID.ID)
		}

		startTime = time.Now()
		resp, err := client.FetchDataBatchFromMongoDB(context.Background(), batchRequest)
		if err != nil {
			http.Error(rw, "Failed to fetch data from MongoDB", http.StatusInternalServerError)
			return
		}
		elapsedTime := time.Since(startTime)
		average += elapsedTime
		fetchedPatients = append(fetchedPatients, resp.FetchedData...)
	}
	average /= time.Duration(len(fetchedPatients))
	log.Println("Average time taken:", average)

	responseData := ResponseData{
		Data:        fetchedPatients,
		AverageTime: average,
	}

	responsePayload, err := json.Marshal(responseData)
	if err != nil {
		http.Error(rw, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")

	_, err = rw.Write(responsePayload)
	if err != nil {
		http.Error(rw, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
