package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Patient struct {
	ID string `json:"patientid"`
}

type ResponseData struct {
	Data        string        `json:"data"`
	AverageTime time.Duration `json:"average_time"`
}

func main() {
	fmt.Println("Starting server at port 8001")
	http.HandleFunc("/", sendRequest)
	log.Fatal(http.ListenAndServe(":8001", nil))
}

var startTime time.Time

func sendRequest(rw http.ResponseWriter, r *http.Request) {

	var average time.Duration
	for i := 0; i < 1; i++ {
		startTime = time.Now()
		elapsedTime := sendBulkRequest(rw)
		average += elapsedTime
		fmt.Println("Total time taken:", elapsedTime)
	}
	average /= 1
	fmt.Println("Average time taken : ", average)

	responseData := ResponseData{
		Data:        "",
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

func sendBulkRequest(rw http.ResponseWriter) time.Duration {

	jsonData, err := ioutil.ReadFile("output.json")
	if err != nil {
		log.Fatal(err)
	}

	var patients []Patient

	if err := json.Unmarshal(jsonData, &patients); err != nil {
		log.Fatal(err)
	}

	batchSize := 100
	totalPatients := len(patients)

	for i := 0; i < totalPatients; i += batchSize {
		end := i + batchSize
		if end > totalPatients {
			end = totalPatients
		}
		if i > 10000 {
			break
		}
		batch := patients[i:end]
		sendBatchRequest(batch, rw)
	}

	return time.Since(startTime)
}

func sendBatchRequest(batch []Patient, rw http.ResponseWriter) {

	var patientIDs []string
	for _, patient := range batch {
		patientIDs = append(patientIDs, patient.ID)
	}

	serviceURL := os.Getenv("SERVICE_URL")

	if serviceURL == "" {
		fmt.Println("SERVICE_URL environment variable is not set.")
		return
	}

	getURL := fmt.Sprintf("http://%s/details?ids=%s", serviceURL, strings.Join(patientIDs, ","))
	response, err := http.Get(getURL)
	if err != nil {
		log.Println("Error:", err)
		return
	}
	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	rw.Write(responseData)
}
