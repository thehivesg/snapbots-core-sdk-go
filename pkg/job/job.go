package job

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
)

type NewJobRequest struct {
	BotID      string `json:"bot_id"`
	ConsumerID string `json:"consumer_id"`
	RequestID  string `json:"request_id"`
}

type GetJobRequest struct {
	BotID      string `json:"bot_id"`
	ConsumerID string `json:"consumer_id"`
	JobID      string `json:"job_id"`
}

type GetJobResponse struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    JobResponse `json:"data"`
}

type JobResponse struct {
	ID            string        `json:"id"`
	BotID         string        `json:"bot_id"`
	ConsumerID    string        `json:"consumer_id"`
	CompletedAt   time.Time     `json:"completed_at"`
	Payload       interface{}   `json:"payload"`
	RequestID     string        `json:"request_id"`
	StartedAt     time.Time     `json:"started_at"`
	Status        string        `json:"status"`
	TotalDuration time.Duration `json:"total_duration"`
	TotalBytes    int64         `json:"total_bytes"`
}

type UpdateJobRequest struct {
	ConsumerID string      `json:"consumer_id"`
	RequestID  string      `json:"request_id"`
	Status     string      `json:"status"`
	Payload    interface{} `json:"payload"`
}

type Service struct {
	natsClient *nats.Conn
}

func NewService(natsClient *nats.Conn) *Service {
	return &Service{
		natsClient: natsClient,
	}
}

func (s *Service) NewJob(request NewJobRequest) (JobResponse, error) {
	jobReqBytes, err := json.Marshal(request)
	if err != nil {
		return JobResponse{}, err
	}

	respBytes, err := s.natsClient.Request("v1.job.request", jobReqBytes, 10*time.Second)
	if err != nil {
		return JobResponse{}, err
	}

	var resp JobResponse
	if err := json.Unmarshal(respBytes.Data, &resp); err != nil {
		return JobResponse{}, err
	}

	return resp, nil
}

func (s *Service) GetJob(request GetJobRequest) (GetJobResponse, error) {
	jobReqBytes, err := json.Marshal(request)
	if err != nil {
		return GetJobResponse{
			Success: false,
			Code:    http.StatusBadRequest,
			Message: "Failed to marshal job get request",
		}, err
	}

	respBytes, err := s.natsClient.Request("v1.job.get", jobReqBytes, 10*time.Second)
	if err != nil {
		return GetJobResponse{
			Success: false,
			Code:    http.StatusInternalServerError,
			Message: "Failed to get job",
		}, err
	}

	var resp GetJobResponse
	if err := json.Unmarshal(respBytes.Data, &resp); err != nil {
		return GetJobResponse{
			Success: false,
			Code:    http.StatusInternalServerError,
			Message: "Failed to unmarshal job get response",
		}, err
	}

	return resp, nil
}

func (s *Service) UpdateJob(request UpdateJobRequest) error {
	jobReqBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	err = s.natsClient.Publish("v1.job.update", jobReqBytes)
	if err != nil {
		return err
	}

	return nil
}
