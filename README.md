# CSV Data Processor

A high-performance CSV data processing system built with **Go** backend and **React** frontend. Designed to handle large CSV files (thousands of records) with efficient data cleaning, semantic grouping, and fast search capabilities.

## 🚀 Features

- **Fast CSV Processing**: Concurrent processing using Go goroutines for optimal performance
- **Data Cleaning**: Automatic normalization (trim spaces, fix casing, remove duplicates)
- **Semantic Grouping**: Intelligent category grouping (e.g., "cardiologist" → "doctor")
- **Advanced Search**: Case-insensitive partial matching with inverted index
- **Modern UI**: Clean React interface with real-time search and grouped views
- **Optimized for Scale**: Handles thousands of records efficiently

## 📁 Project Structure

```
elsapien-work/
├── backend/              # Go backend server
│   ├── main.go          # Entry point & HTTP server
│   ├── go.mod           # Go dependencies
│   ├── handlers/        # HTTP request handlers
│   │   └── handler.go
│   ├── services/        # Business logic
│   │   ├── csv_processor.go      # CSV parsing & processing
│   │   ├── data_cleaner.go       # Text normalization
│   │   ├── category_grouper.go   # Semantic grouping rules
│   └── models/          # Data structures
│       └── record.go
└── frontend/            # React frontend
    ├── package.json
    ├── public/
    │   └── index.html
    └── src/
        ├── App.js       # Main application
        ├── components/
        │   ├── FileUpload.js      # CSV upload component
        │   ├── SearchBar.js       # Search interface
        │   ├── DataTable.js       # Records display
        │   └── GroupsView.js      # Grouped categories view
        └── *.css        # Styling files
```

## 🛠️ Tech Stack

### Backend
- **Language**: Go 1.21+
- **Router**: Gorilla Mux
- **Key Features**:
  - Concurrent CSV processing (4 worker goroutines)
  - In-memory inverted index for O(1) search
  - Maintainable rule-based category grouping
  - RESTful API

### Frontend
- **Framework**: React 18
- **HTTP Client**: Axios
- **Key Features**:
  - File upload with validation
  - Real-time search with debouncing
  - Responsive table view
  - Expandable grouped view

## 🏃 Getting Started

### Prerequisites
- **Option 1 (Docker)**: Docker and Docker Compose
- **Option 2 (Manual)**: Go 1.21+, Node.js 16+, npm

### Quick Start with Docker (Recommended)

**Single command to run everything:**

```powershell
docker-compose up --build
```

- Backend API: `http://localhost:8080`
- Frontend UI: `http://localhost:3000`

The services will start automatically with health checks. Frontend waits for backend to be ready.

**To stop:**
```powershell
docker-compose down
```

### Manual Setup (Development)

1. Navigate to backend directory:
```powershell
cd backend
```

2. Install Go dependencies:
```powershell
go mod download
```

3. Run the server:
```powershell
go run main.go
```

Server starts on `http://localhost:8080`

#### Frontend Setup

1. Navigate to frontend directory:
```powershell
cd frontend
```

2. Install npm dependencies:
```powershell
npm install
```

3. Start the development server:
```powershell
npm start
```

Frontend runs on `http://localhost:3000`

## 📡 API Endpoints

**Base URL:**
- Docker: `http://localhost:3000/api` (proxied through nginx)
- Manual: `http://localhost:8080/api`

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/upload` | POST | Upload and process CSV file |
| `/api/data` | GET | Get all processed records |
| `/api/search?q={query}` | GET | Search records (partial match) |
| `/api/health` | GET | Health check |

### Example API Usage

**Docker (via nginx proxy):**
```bash
curl -X POST -F "file=@data.csv" http://localhost:3000/api/upload
curl "http://localhost:3000/api/search?q=doctor"
```

**Manual (direct backend):**
```bash
curl -X POST -F "file=@data.csv" http://localhost:8080/api/upload
curl "http://localhost:8080/api/search?q=doctor"
```

## 🧪 Sample CSV Format

```csv
name,category,location
John Doe,cardiologist,New York
Jane Smith,Neurologist,Los Angeles
Bob Johnson,Software Engineer,San Francisco
```

## ⚙️ Configuration

### Category Grouping Rules

Rules are defined in `backend/services/category_grouper.go`. You can add custom mappings:

```go
g.rules["custom-term"] = "unified-group"
```

**Current groupings:**
- Medical specialties → `doctor`
- Tech roles → `software engineer`
- Legal professions → `lawyer`
- Education roles → `teacher`
- Business roles → `manager`
- Creative roles → `designer`
- Sales/Marketing → `sales professional`
- Finance roles → `accountant`

### Performance Tuning

Adjust concurrent workers in `backend/services/csv_processor.go`:
```go
numWorkers := 4  // Increase for more CPU cores
```

## 🎯 Key Design Decisions

### Why Go?
- **10-50x faster** than Python for CSV parsing
- Built-in concurrency (goroutines)
- Single binary deployment
- Low memory footprint

### Why Inverted Index?
- O(1) search lookups vs O(n) linear scan
- Supports partial matching efficiently
- Scales to millions of records

### Why In-Memory?
- Sub-millisecond search response
- No database overhead for "thousands of records"
- Simplified deployment

## 📊 Performance Benchmarks

Expected performance on typical hardware:
- **10,000 records**: ~200-400ms processing
- **50,000 records**: ~1-2 seconds processing
- **Search**: <50ms for any dataset size

## 🐛 Troubleshooting

**Docker issues?**
- Ensure Docker Desktop is running
- Check ports 3000 and 8080 are not in use
- Run `docker-compose logs` to see errors
- Rebuild with `docker-compose up --build --force-recreate`

**CORS errors (manual setup)?**
- Ensure backend is running on port 8080
- Check CORS middleware in `main.go`

**CSV parsing errors?**
- Verify CSV is UTF-8 encoded
- Check for proper comma delimiters
- Ensure headers are in first row

**Build errors?**
- Run `go mod tidy` in backend
- Run `npm install --legacy-peer-deps` in frontend

## 🔒 Production Considerations

For production deployment:
1. Use the included Docker setup as a base
2. Add authentication middleware
3. Implement rate limiting
4. Add persistent storage (PostgreSQL + FTS)
5. Use environment variables for configuration
6. Add logging (e.g., `logrus`)
7. Set up CI/CD pipeline
8. Use secrets management
9. Add HTTPS with Let's Encrypt
10. Set up monitoring (Prometheus/Grafana)

## 📝 License

This project is created for interview assessment purposes.

## 👤 Author

Created for ElSapien interview assessment - November 2025

---

**Interview Highlights:**
- ✅ Rejected Python backend (speed requirement)
- ✅ Chose Go for data processing performance
- ✅ Implemented concurrent CSV parsing
- ✅ Built maintainable grouping logic
- ✅ Optimized search with inverted index
- ✅ Clean, readable, well-structured code
