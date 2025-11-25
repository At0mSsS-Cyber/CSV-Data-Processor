package services

import (
	"csv-processor/database"
	"csv-processor/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type DBService struct {
	db *sql.DB
}

func NewDBService() *DBService {
	return &DBService{
		db: database.DB,
	}
}

// CreateCSVFile creates a new CSV file record
func (s *DBService) CreateCSVFile(filename string, fileSize int64) (*models.CSVFile, error) {
	query := `
		INSERT INTO csv_files (filename, file_size, status, uploaded_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, filename, file_size, status, record_count, processing_time_ms, uploaded_at
	`

	file := &models.CSVFile{}
	err := s.db.QueryRow(query, filename, fileSize, "processing", time.Now()).Scan(
		&file.ID,
		&file.Filename,
		&file.FileSize,
		&file.Status,
		&file.RecordCount,
		&file.ProcessingTimeMs,
		&file.UploadedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create CSV file record: %w", err)
	}

	return file, nil
}

// UpdateCSVFileStatus updates the status of a CSV file
func (s *DBService) UpdateCSVFileStatus(fileID int, status string, recordCount int, processingTimeMs int64, errorMsg string) error {
	completedAt := time.Now()
	query := `
		UPDATE csv_files
		SET status = $1, record_count = $2, processing_time_ms = $3, error_message = $4, completed_at = $5
		WHERE id = $6
	`

	_, err := s.db.Exec(query, status, recordCount, processingTimeMs, errorMsg, completedAt, fileID)
	if err != nil {
		return fmt.Errorf("failed to update CSV file status: %w", err)
	}

	return nil
}

// InsertRecords inserts multiple records in batches for better performance
func (s *DBService) InsertRecords(records []*models.Record) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Process in batches of 2000 records
	batchSize := 2000
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[i:end]
		
		// Use COPY for PostgreSQL bulk insert (much faster)
		stmt, err := tx.Prepare(pq.CopyIn("records", "csv_file_id", "original_data", "cleaned_data", "grouped_category", "created_at"))
		if err != nil {
			return fmt.Errorf("failed to prepare copy statement: %w", err)
		}

		for _, record := range batch {
			originalJSON, err := json.Marshal(record.OriginalData)
			if err != nil {
				stmt.Close()
				return fmt.Errorf("failed to marshal original data: %w", err)
			}
			
			cleanedJSON, err := json.Marshal(record.CleanedData)
			if err != nil {
				stmt.Close()
				return fmt.Errorf("failed to marshal cleaned data: %w", err)
			}

			_, err = stmt.Exec(
				record.CSVFileID,
				string(originalJSON),
				string(cleanedJSON),
				record.GroupedCategory,
				time.Now(),
			)
			if err != nil {
				stmt.Close()
				return fmt.Errorf("failed to exec copy: %w", err)
			}
		}

		_, err = stmt.Exec()
		if err != nil {
			stmt.Close()
			return fmt.Errorf("failed to flush copy: %w", err)
		}
		
		stmt.Close()
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetAllCSVFiles retrieves all CSV files
func (s *DBService) GetAllCSVFiles() ([]*models.CSVFile, error) {
	query := `
		SELECT id, filename, file_size, status, record_count, processing_time_ms, 
		       COALESCE(error_message, ''), uploaded_at, completed_at
		FROM csv_files
		ORDER BY uploaded_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query CSV files: %w", err)
	}
	defer rows.Close()

	files := make([]*models.CSVFile, 0)
	for rows.Next() {
		file := &models.CSVFile{}
		var completedAt sql.NullTime

		err := rows.Scan(
			&file.ID,
			&file.Filename,
			&file.FileSize,
			&file.Status,
			&file.RecordCount,
			&file.ProcessingTimeMs,
			&file.ErrorMessage,
			&file.UploadedAt,
			&completedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan CSV file: %w", err)
		}

		if completedAt.Valid {
			file.CompletedAt = &completedAt.Time
		}

		files = append(files, file)
	}

	return files, nil
}

// GetCSVFile retrieves a single CSV file by ID
func (s *DBService) GetCSVFile(fileID int) (*models.CSVFile, error) {
	query := `
		SELECT id, filename, file_size, status, record_count, processing_time_ms,
		       COALESCE(error_message, ''), uploaded_at, completed_at
		FROM csv_files
		WHERE id = $1
	`

	file := &models.CSVFile{}
	var completedAt sql.NullTime

	err := s.db.QueryRow(query, fileID).Scan(
		&file.ID,
		&file.Filename,
		&file.FileSize,
		&file.Status,
		&file.RecordCount,
		&file.ProcessingTimeMs,
		&file.ErrorMessage,
		&file.UploadedAt,
		&completedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("CSV file not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get CSV file: %w", err)
	}

	if completedAt.Valid {
		file.CompletedAt = &completedAt.Time
	}

	return file, nil
}

// GetRecordsByFileID retrieves all records for a specific CSV file with pagination
func (s *DBService) GetRecordsByFileID(fileID int, limit, offset int) ([]*models.Record, int, error) {
	// Get total count
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM records WHERE csv_file_id = $1`
	err := s.db.QueryRow(countQuery, fileID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get record count: %w", err)
	}

	// Get paginated records
	query := `
		SELECT id, csv_file_id, original_data, cleaned_data, 
		       COALESCE(grouped_category, ''), created_at
		FROM records
		WHERE csv_file_id = $1
		ORDER BY id
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.Query(query, fileID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query records: %w", err)
	}
	defer rows.Close()

	records, err := s.scanRecords(rows)
	if err != nil {
		return nil, 0, err
	}

	return records, totalCount, nil
}

// SearchRecords performs full-text search on records for a specific file with pagination
func (s *DBService) SearchRecords(fileID int, query string, limit, offset int) ([]*models.Record, int, error) {
	likePattern := "%" + query + "%"

	// Get total count of matching records
	var totalCount int
	countQuery := `
		SELECT COUNT(*)
		FROM records
		WHERE csv_file_id = $1 
		  AND (
		    search_vector @@ plainto_tsquery('english', $2)
		    OR cleaned_data::text ILIKE $3
		    OR grouped_category ILIKE $3
		  )
	`
	err := s.db.QueryRow(countQuery, fileID, query, likePattern).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get search count: %w", err)
	}

	// Get paginated search results
	sqlQuery := `
		SELECT id, csv_file_id, original_data, cleaned_data, 
		       COALESCE(grouped_category, ''), created_at
		FROM records
		WHERE csv_file_id = $1 
		  AND (
		    search_vector @@ plainto_tsquery('english', $2)
		    OR cleaned_data::text ILIKE $3
		    OR grouped_category ILIKE $3
		  )
		ORDER BY id
		LIMIT $4 OFFSET $5
	`

	rows, err := s.db.Query(sqlQuery, fileID, query, likePattern, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search records: %w", err)
	}
	defer rows.Close()

	records, err := s.scanRecords(rows)
	if err != nil {
		return nil, 0, err
	}

	return records, totalCount, nil
}

// scanRecords is a helper function to scan rows into Record structs
func (s *DBService) scanRecords(rows *sql.Rows) ([]*models.Record, error) {
	records := make([]*models.Record, 0)

	for rows.Next() {
		record := &models.Record{}
		var originalJSON, cleanedJSON []byte

		err := rows.Scan(
			&record.ID,
			&record.CSVFileID,
			&originalJSON,
			&cleanedJSON,
			&record.GroupedCategory,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan record: %w", err)
		}

		// Parse JSON
		json.Unmarshal(originalJSON, &record.OriginalData)
		json.Unmarshal(cleanedJSON, &record.CleanedData)

		records = append(records, record)
	}

	return records, nil
}

// GetGroupsByFileID retrieves grouped categories for a specific file
func (s *DBService) GetGroupsByFileID(fileID int) (map[string][]int, error) {
	query := `
		SELECT grouped_category, array_agg(id ORDER BY id) as record_ids
		FROM records
		WHERE csv_file_id = $1 AND grouped_category IS NOT NULL AND grouped_category != ''
		GROUP BY grouped_category
	`

	rows, err := s.db.Query(query, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to query groups: %w", err)
	}
	defer rows.Close()

	groups := make(map[string][]int)
	for rows.Next() {
		var category string
		var recordIDs pq.Int64Array

		err := rows.Scan(&category, &recordIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}

		// Convert []int64 to []int
		intIDs := make([]int, len(recordIDs))
		for i, id := range recordIDs {
			intIDs[i] = int(id)
		}

		groups[category] = intIDs
	}

	return groups, nil
}

// GetRecordsByGroup retrieves records for a specific group category with pagination
func (s *DBService) GetRecordsByGroup(fileID int, groupCategory string, limit, offset int) ([]*models.Record, int, error) {
	// First get total count for this group
	countQuery := `
		SELECT COUNT(*)
		FROM records
		WHERE csv_file_id = $1 AND grouped_category = $2
	`
	var totalCount int
	err := s.db.QueryRow(countQuery, fileID, groupCategory).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count group records: %w", err)
	}

	// Then get paginated records
	query := `
		SELECT id, csv_file_id, original_data, cleaned_data, grouped_category, created_at
		FROM records
		WHERE csv_file_id = $1 AND grouped_category = $2
		ORDER BY id
		LIMIT $3 OFFSET $4
	`

	rows, err := s.db.Query(query, fileID, groupCategory, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query group records: %w", err)
	}
	defer rows.Close()

	records := make([]*models.Record, 0)
	for rows.Next() {
		record := &models.Record{}
		var originalDataJSON, cleanedDataJSON []byte
		var groupedCategory sql.NullString

		err := rows.Scan(
			&record.ID,
			&record.CSVFileID,
			&originalDataJSON,
			&cleanedDataJSON,
			&groupedCategory,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan record: %w", err)
		}

		// Parse JSON data
		if err := json.Unmarshal(originalDataJSON, &record.OriginalData); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal original data: %w", err)
		}
		if err := json.Unmarshal(cleanedDataJSON, &record.CleanedData); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal cleaned data: %w", err)
		}

		if groupedCategory.Valid {
			record.GroupedCategory = groupedCategory.String
		}

		records = append(records, record)
	}

	return records, totalCount, nil
}
