package handlers

import (
	"bytes"
	"csv-processor/models"
	"csv-processor/services"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

type Handler struct {
	dbService      *services.DBService
	asyncProcessor *services.AsyncProcessor
}

func NewHandler(dbService *services.DBService, asyncProcessor *services.AsyncProcessor) *Handler {
	return &Handler{
		dbService:      dbService,
		asyncProcessor: asyncProcessor,
	}
}

// HandleUpload processes CSV file uploads
func (h *Handler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 100MB)
	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		http.Error(w, "File too large or invalid", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create CSV file record in database
	csvFile, err := h.dbService.CreateCSVFile(header.Filename, header.Size)
	if err != nil {
		http.Error(w, "Error creating file record: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Read file content into memory
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Process CSV asynchronously
	h.asyncProcessor.ProcessCSVAsync(csvFile.ID, bytes.NewReader(fileBytes))

	// Send immediate response
	response := models.UploadResponse{
		Message: "CSV file uploaded successfully. Processing in background.",
		FileID:  csvFile.ID,
		File:    csvFile,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetFiles returns all CSV files
func (h *Handler) HandleGetFiles(w http.ResponseWriter, r *http.Request) {
	files, err := h.dbService.GetAllCSVFiles()
	if err != nil {
		http.Error(w, "Error fetching files: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.FilesListResponse{
		Files: files,
		Count: len(files),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetFile returns a specific CSV file
func (h *Handler) HandleGetFile(w http.ResponseWriter, r *http.Request) {
	fileIDStr := r.URL.Query().Get("id")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	file, err := h.dbService.GetCSVFile(fileID)
	if err != nil {
		http.Error(w, "File not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(file)
}

// HandleGetRecords returns all records for a specific file with pagination and optional search
func (h *Handler) HandleGetRecords(w http.ResponseWriter, r *http.Request) {
	fileIDStr := r.URL.Query().Get("fileId")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	// Pagination parameters
	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("perPage")
	query := r.URL.Query().Get("q") // Optional search query
	
	page := 1
	perPage := 100 // Default page size
	
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 1000 {
			perPage = pp
		}
	}

	offset := (page - 1) * perPage

	// Choose between search and regular fetch based on query parameter
	var records []*models.Record
	var totalCount int
	
	if query != "" {
		// Perform optimized full-text search
		records, totalCount, err = h.dbService.SearchRecords(fileID, query, perPage, offset)
		if err != nil {
			http.Error(w, "Error searching records: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Regular fetch all records
		records, totalCount, err = h.dbService.GetRecordsByFileID(fileID, perPage, offset)
		if err != nil {
			http.Error(w, "Error fetching records: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Fetch groups only on first page request (without search)
	var groups map[string][]int
	if page == 1 && query == "" {
		groups, err = h.dbService.GetGroupsByFileID(fileID)
		if err != nil {
			http.Error(w, "Error fetching groups: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	response := models.DataResponse{
		Records:    records,
		Groups:     groups,
		Count:      len(records),
		TotalCount: totalCount,
		Page:       page,
		PerPage:    perPage,
		HasMore:    offset+len(records) < totalCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}



// HandleGetGroupRecords returns records for a specific group with pagination
func (h *Handler) HandleGetGroupRecords(w http.ResponseWriter, r *http.Request) {
	fileIDStr := r.URL.Query().Get("fileId")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	groupCategory := r.URL.Query().Get("group")
	if groupCategory == "" {
		http.Error(w, "Group parameter is required", http.StatusBadRequest)
		return
	}

	// Pagination parameters
	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("perPage")
	
	page := 1
	perPage := 20 // Default smaller page size for groups
	
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}

	offset := (page - 1) * perPage

	records, totalCount, err := h.dbService.GetRecordsByGroup(fileID, groupCategory, perPage, offset)
	if err != nil {
		http.Error(w, "Error fetching group records: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := models.DataResponse{
		Records:    records,
		Count:      len(records),
		TotalCount: totalCount,
		Page:       page,
		PerPage:    perPage,
		HasMore:    offset+len(records) < totalCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleHealth is a health check endpoint
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
