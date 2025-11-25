-- Create csv_files table
CREATE TABLE IF NOT EXISTS csv_files (
    id SERIAL PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'processing',
    record_count INT DEFAULT 0,
    processing_time_ms BIGINT DEFAULT 0,
    error_message TEXT,
    uploaded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

-- Create records table
CREATE TABLE IF NOT EXISTS records (
    id SERIAL PRIMARY KEY,
    csv_file_id INT NOT NULL REFERENCES csv_files(id) ON DELETE CASCADE,
    original_data JSONB NOT NULL,
    cleaned_data JSONB NOT NULL,
    grouped_category VARCHAR(100),
    search_vector TSVECTOR,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for fast search
CREATE INDEX IF NOT EXISTS idx_records_csv_file_id ON records(csv_file_id);
CREATE INDEX IF NOT EXISTS idx_records_grouped_category ON records(grouped_category);
CREATE INDEX IF NOT EXISTS idx_records_search_vector ON records USING GIN(search_vector);
CREATE INDEX IF NOT EXISTS idx_records_cleaned_data ON records USING GIN(cleaned_data);
CREATE INDEX IF NOT EXISTS idx_csv_files_status ON csv_files(status);
CREATE INDEX IF NOT EXISTS idx_csv_files_uploaded_at ON csv_files(uploaded_at DESC);

-- Function to update search vector
CREATE OR REPLACE FUNCTION update_search_vector() RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector := to_tsvector('english', 
        COALESCE(NEW.cleaned_data::text, '') || ' ' || 
        COALESCE(NEW.grouped_category, '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update search vector
CREATE TRIGGER records_search_vector_update
    BEFORE INSERT OR UPDATE ON records
    FOR EACH ROW
    EXECUTE FUNCTION update_search_vector();
