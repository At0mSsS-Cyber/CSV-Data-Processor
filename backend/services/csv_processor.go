package services

import (
	"csv-processor/models"
	"encoding/csv"
	"io"
	"strings"
	"sync"
	"time"
)

type CSVProcessor struct {
	records []*models.Record
	groups  map[string][]int // category -> record IDs
	mu      sync.RWMutex
	grouper *CategoryGrouper
	cleaner *DataCleaner
}

func NewCSVProcessor() *CSVProcessor {
	return &CSVProcessor{
		records: make([]*models.Record, 0),
		groups:  make(map[string][]int),
		grouper: NewCategoryGrouper(),
		cleaner: NewDataCleaner(),
	}
}

// ProcessCSV reads and processes a CSV file
func (p *CSVProcessor) ProcessCSV(file io.Reader) ([]*models.Record, int64, error) {
	startTime := time.Now()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, 0, err
	}

	// Clean headers
	for i, header := range headers {
		headers[i] = p.cleaner.CleanText(header)
	}

	// Auto-detect category column
	_ = p.detectCategoryColumn(headers)

	// Read all rows first
	allRows := make([][]string, 0, 1000) // Pre-allocate with reasonable capacity
	recordID := 1
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, 0, err
		}
		allRows = append(allRows, append([]string{string(rune(recordID))}, row...))
		recordID++
	}

	// Process rows in batches for better performance
	batchSize := 1000
	records := make([]*models.Record, 0, len(allRows))
	
	for i := 0; i < len(allRows); i += batchSize {
		end := i + batchSize
		if end > len(allRows) {
			end = len(allRows)
		}
		
		// Process batch concurrently
		batch := allRows[i:end]
		batchRecords := p.processBatch(headers, batch, i+1)
		records = append(records, batchRecords...)
	}

	// Store records and build groups
	p.mu.Lock()
	p.records = records
	p.buildGroups()
	p.mu.Unlock()

	processingTime := time.Since(startTime).Milliseconds()
	return records, processingTime, nil
}

// processBatch processes a batch of rows concurrently with thread-safe normalization
func (p *CSVProcessor) processBatch(headers []string, batch [][]string, startID int) []*models.Record {
	records := make([]*models.Record, len(batch))
	
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // Limit to 10 concurrent workers. Semaphore is a buffered channel
	
	for i, row := range batch {
		wg.Add(1)
		go func(idx int, rowData []string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release
			
			records[idx] = p.processRow(headers, rowData, startID+idx)
		}(i, row)
	}
	
	wg.Wait()
	return records
}

func (p *CSVProcessor) processRow(headers []string, row []string, id int) *models.Record {
	originalData := make(map[string]string)
	cleanedData := make(map[string]string)

	// Process each column
	for i, value := range row {
		if i == 0 {
			continue // Skip ID column
		}
		if i-1 < len(headers) {
			header := headers[i-1]
			originalData[header] = value
			
			// Clean the text
			cleaned := p.cleaner.CleanText(value)
			cleanedData[header] = cleaned
		}
	}

	// Detect category grouping from any available field
	groupedCategory := p.detectCategory(cleanedData)

	return &models.Record{
		ID:              id,
		OriginalData:    originalData,
		CleanedData:     cleanedData,
		GroupedCategory: groupedCategory,
	}
}

func (p *CSVProcessor) detectCategory(data map[string]string) string {
	// Priority-ordered list of category-like field names
	categoryFields := []string{
		"category", "type", "specialty", "profession", "occupation",
		"role", "title", "job", "position", "designation",
		"department", "field", "industry", "sector", "skill",
	}
	
	// First, try priority fields (case-insensitive lookup)
	for _, field := range categoryFields {
		// Try both lowercase and title case versions
		for key, value := range data {
			if strings.EqualFold(key, field) && value != "" {
				groupedCategory := p.grouper.GetGroup(value)
				if groupedCategory != "" {
					return groupedCategory
				}
				break
			}
		}
	}
	
	// For "name" field, only try grouping if it looks like a category
	// (avoid grouping random product names, company names, etc.)
	// Allow shorter names (>= 2 chars) to catch abbreviations like SEO, CRM, HR, IT
	for key, value := range data {
		if strings.EqualFold(key, "name") && value != "" && len(value) >= 2 {
			groupedCategory := p.grouper.GetGroup(value)
			// Only use if it actually mapped to a recognized group
			if groupedCategory != "" {
				return groupedCategory
			}
			break
		}
	}

	return ""
}

// detectCategoryColumn finds the most likely category column from headers
func (p *CSVProcessor) detectCategoryColumn(headers []string) string {
	// Keywords that indicate a category-like column (ordered by priority)
	categoryFields := []string{
		"category", "type", "specialty", "profession", "occupation",
		"role", "title", "job", "position", "designation",
		"department", "field", "industry", "sector", "work",
	}

	// First pass: exact match
	for _, header := range headers {
		headerLower := strings.ToLower(header)
		for _, keyword := range categoryFields {
			if headerLower == keyword {
				return header
			}
		}
	}

	// Second pass: contains match
	for _, header := range headers {
		headerLower := strings.ToLower(header)
		for _, keyword := range categoryFields {
			if strings.Contains(headerLower, keyword) {
				return header
			}
		}
	}

	return "" // No category column found
}

func (p *CSVProcessor) buildGroups() {
	p.groups = make(map[string][]int)
	
	for _, record := range p.records {
		if record.GroupedCategory != "" {
			p.groups[record.GroupedCategory] = append(p.groups[record.GroupedCategory], record.ID)
		}
	}
}

func (p *CSVProcessor) GetRecords() []*models.Record {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.records
}

func (p *CSVProcessor) GetGroups() map[string][]int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.groups
}
