package models

import "time"

// CSVFile represents an uploaded CSV file
type CSVFile struct {
	ID               int        `json:"id"`
	Filename         string     `json:"filename"`
	FileSize         int64      `json:"fileSize"`
	Status           string     `json:"status"` // processing, completed, failed
	RecordCount      int        `json:"recordCount"`
	ProcessingTimeMs int64      `json:"processingTimeMs"`
	ErrorMessage     string     `json:"errorMessage,omitempty"`
	UploadedAt       time.Time  `json:"uploadedAt"`
	CompletedAt      *time.Time `json:"completedAt,omitempty"`
}

// Record represents a single row from the CSV file after processing
type Record struct {
	ID              int               `json:"id"`
	CSVFileID       int               `json:"csvFileId"`
	OriginalData    map[string]string `json:"originalData"`
	CleanedData     map[string]string `json:"cleanedData"`
	GroupedCategory string            `json:"groupedCategory,omitempty"`
	CreatedAt       time.Time         `json:"createdAt"`
}

// UploadResponse represents the response after CSV upload
type UploadResponse struct {
	Message string   `json:"message"`
	FileID  int      `json:"fileId"`
	File    *CSVFile `json:"file"`
}

// DataResponse represents the response for getting all data
type DataResponse struct {
	Records    []*Record        `json:"records"`
	Groups     map[string][]int `json:"groups"` // category -> record IDs
	Count      int              `json:"count"`
	TotalCount int              `json:"totalCount"`
	Page       int              `json:"page"`
	PerPage    int              `json:"perPage"`
	HasMore    bool             `json:"hasMore"`
}

// FilesListResponse represents the list of all CSV files
type FilesListResponse struct {
	Files []*CSVFile `json:"files"`
	Count int        `json:"count"`
}
