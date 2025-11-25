package services

import (
	"io"
	"log"
	"time"
)

type AsyncProcessor struct {
	csvProcessor *CSVProcessor
	dbService    *DBService
}

func NewAsyncProcessor(dbService *DBService) *AsyncProcessor {
	return &AsyncProcessor{
		csvProcessor: NewCSVProcessor(),
		dbService:    dbService,
	}
}

// ProcessCSVAsync processes CSV file in the background
func (p *AsyncProcessor) ProcessCSVAsync(fileID int, file io.Reader) {
	go func() {
		startTime := time.Now()

		// Process CSV
		records, processingTime, err := p.csvProcessor.ProcessCSV(file)
		if err != nil {
			log.Printf("Error processing CSV file %d: %v", fileID, err)
			p.dbService.UpdateCSVFileStatus(fileID, "failed", 0, 0, err.Error())
			return
		}

		// Add file ID to all records
		for _, record := range records {
			record.CSVFileID = fileID
		}

		// Insert records into database
		err = p.dbService.InsertRecords(records)
		if err != nil {
			log.Printf("Error inserting records for file %d: %v", fileID, err)
			p.dbService.UpdateCSVFileStatus(fileID, "failed", 0, 0, err.Error())
			return
		}

		// Update file status
		totalTime := time.Since(startTime).Milliseconds()
		err = p.dbService.UpdateCSVFileStatus(fileID, "completed", len(records), totalTime, "")
		if err != nil {
			log.Printf("Error updating file status for %d: %v", fileID, err)
		}

		log.Printf("Successfully processed file %d: %d records in %dms", fileID, len(records), processingTime)
	}()
}
