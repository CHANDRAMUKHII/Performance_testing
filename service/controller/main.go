package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"patient-service/model"
	"strings"
	"sync"

	"patient-service/pb"
)

type Patient struct {
	model.Patient
}

func HandleBulkRequest(rw http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {

		ids := r.URL.Query().Get("ids")
		if ids != "" {
			idList := strings.Split(ids, ",")
			var patients []Patient
			patientChan := make(chan model.Patient)
			done := make(chan struct{})

			for _, patientID := range idList {
				go func(pid string) {
					defer func() {
						done <- struct{}{}
					}()

					patient, err := model.FetchData(pid)
					if err != nil {
						fmt.Printf("Error retrieving details for patient ID %s: %v\n", pid, err)
						patient.MedicalHistory = "Patient not found"
					}
					patientChan <- patient
				}(patientID)
			}

			go func() {
				for range idList {
					<-done
				}
				close(patientChan)
			}()

			for patient := range patientChan {
				patients = append(patients, Patient{Patient: patient})
			}

			responseJSON, err := json.Marshal(patients)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

			rw.Header().Set("Content-Type", "application/json")
			rw.Write(responseJSON)
			return
		}
	}

}

// GRPC CODE

type Server struct {
	pb.UnimplementedMongoDBServiceServer
	Model *model.MongoDBModel
}

func (s *Server) FetchDataBatchFromMongoDB(ctx context.Context, req *pb.BatchFetchRequest) (*pb.BatchFetchResponse, error) {
	fetchedData := make([]*pb.Patient, 0, len(req.PatientIds))
	responseChan := make(chan *pb.Patient, len(req.PatientIds))
	done := make(chan bool)

	var wg sync.WaitGroup
	wg.Add(len(req.PatientIds))

	for _, patientID := range req.PatientIds {
		go func(id string) {
			defer wg.Done()
			modelPatient := s.Model.FetchData(ctx, id)
			pbPatient := convertModelToPB(modelPatient)

			responseChan <- pbPatient
		}(patientID)
	}

	go func() {
		wg.Wait()
		done <- true
	}()

	for i := 0; i < len(req.PatientIds); i++ {
		data := <-responseChan
		fetchedData = append(fetchedData, data)
	}

	close(responseChan)
	<-done

	return &pb.BatchFetchResponse{
		FetchedData: fetchedData,
	}, nil
}

func convertModelToPB(modelPatient model.Patient) *pb.Patient {
	pbPatient := &pb.Patient{
		PatientID:      modelPatient.PatientID,
		FirstName:      modelPatient.FirstName,
		LastName:       modelPatient.LastName,
		DateofBirth:    modelPatient.DateofBirth,
		Gender:         modelPatient.Gender,
		ContactNumber:  modelPatient.ContactNumber,
		MedicalHistory: modelPatient.MedicalHistory,
	}
	return pbPatient
}
