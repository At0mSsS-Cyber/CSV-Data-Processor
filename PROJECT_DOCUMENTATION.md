# CSV Data Processing System - Project Documentation

## Overview

This document explains the technical choices and development process for building a CSV data processing system. The goal was to create something that can handle large CSV files, clean messy data, group similar categories, and make everything searchable quickly.

**Development Time:** Around 18-20 hours
- Learning Go basics: 5 hours
- Building everything: 13-15 hours

I'm pretty new to Go, so I used AI tools (ChatGPT, Copilot) to speed things up.

---

## What I Built

The system needs to:
- Accept CSV uploads with thousands of rows
- Clean up messy text (extra spaces, weird characters, inconsistent formatting)
- Group similar things together (like "cardiologist" and "neurologist" both become "doctor")
- Make everything searchable fast
- Show the results in a simple UI


---


## Why I Chose Each Technology

### Backend: Go vs Python vs Node.js vs Rust

I compared four options for the backend:

**Go (What I Chose)**
- Really fast at parsing CSV files - way faster than Python or Node
- Built-in concurrency with goroutines made it easy to process rows in parallel
- Compiles to a single binary, so deployment is simple
- The learning curve wasn't too bad, especially with AI help
- Memory efficient, which matters when dealing with large files

**Python**
- Would've been easier since I know it better
- pandas is great for CSV work, but it's slow compared to Go
- The GIL (Global Interpreter Lock) makes true parallelism tricky
- Uses more memory

**Node.js**
- I'm comfortable with JavaScript
- But it's single-threaded by default - worker threads exist but they're clunky
- CSV parsing is slower than Go
- Good for web APIs but not the best for heavy data processing
- Still would've been a solid choice honestly

**Rust**
- Fastest of them all, probably
- But the learning curve is steep - lifetimes, borrowing, ownership
- Would've taken me way longer than 5 hours to learn enough
- Overkill for "thousands of records"
- Maybe if this needed to handle millions, I'd consider it

**Why Go Won:** It hit the sweet spot between performance and ease of learning. Fast enough to handle large files, simple enough that I could learn it in a few hours. The concurrency model is really nice once you get it.

---

### Database: PostgreSQL

I went with PostgreSQL because:
- Has built-in full-text search (no need for Elasticsearch)
- JSONB columns let me store flexible CSV data
- GIN indexes make search really fast
- It's reliable and I've used it before

I thought about just keeping everything in memory, but that doesn't survive restarts. Elasticsearch seemed like overkill for this size of data.

---

### Data Cleaning Approach

Just went with simple rules:
- Trim whitespace
- Remove weird characters (arrows, bullets, etc.)
- Collapse multiple spaces
- Standardize to title case

I considered using machine learning for text cleaning but that felt like over-engineering. Simple regex patterns work fine and they're predictable.

---

### Grouping Strategy

This was the interesting part. I used a three-step matching approach:

1. **Exact match first** - "doctor" matches "doctor"
2. **Partial word match** - "heart doctor" contains "doctor"
3. **Fuzzy match for typos** - "docter" is close enough to "doctor"

I keep all the category mappings in a simple map structure. Like:
```
"doctor" → ["cardiologist", "neurologist", "surgeon", ...]
"software engineer" → ["developer", "programmer", "coder", ...]
```

Initially I built a fancy Trie-based system with similarity scoring, but I realized it was overkill and removed it all. The simple approach works better and it's way easier to maintain.

I thought about using AI embeddings or semantic similarity, but that would be slow and expensive. This isn't a text classification problem - the categories are well-defined and structured.

---

### Frontend: React

Just used React because:
- I know it already
- Component model makes the UI easy to organize
- Has a good ecosystem
- Quick to prototype

Nothing fancy here - functional components, some basic state management, and Tailwind for styling.

---

### Concurrency Pattern

Go's goroutines made this easy. I process CSV rows in batches of 1000, and use 10 worker goroutines to handle them in parallel. 

The trick is using a "semaphore" pattern (a buffered channel) to limit how many goroutines run at once. Without it, spawning thousands of goroutines would overwhelm the CPU.

---

### Deployment: Docker Compose

Everything runs in Docker containers:
- PostgreSQL container
- Go backend container
- React frontend (served by nginx)

One command starts everything: `docker-compose up --build`

Health checks make sure services start in the right order. The backend waits for Postgres to be ready, and the frontend waits for the backend.

This makes it really easy to run the same setup on my machine and in production.

---

## Development Timeline

### Phase 1: Learning Go (5 hours)

I spent a day going through Go tutorials:
- Basic syntax (types, structs, interfaces)
- Goroutines and channels
- How to build a web server
- Database connections

Resources I used:
- Official Go tour
- "Effective Go" docs
- Some YouTube tutorials
- Asked ChatGPT when I got stuck

### Phase 2: Building the Backend (8-10 hours)

**First few hours:** Set up the project structure, got CSV parsing working, built the basic HTTP server.

**Middle hours:** This is where I spent the most time - building the data cleaning logic, the category grouping system, and getting concurrency right.

**Last hours:** Database integration, building the search functionality, and hooking up all the API endpoints. Also spent time removing code that I realized I didn't need.

### Phase 3: Building the Frontend (3-4 hours)

Created the React app, built the components (file upload, search bar, data table, groups view). Most of this was straightforward since I know React.

The trickiest part was getting the pagination and infinite scroll working smoothly.

### Phase 4: Docker and Deployment (2 hours)

Wrote the Dockerfiles, set up docker-compose, fixed some networking issues, added health checks. Tested the whole thing end-to-end.

---

## Using AI Tools

I used ChatGPT and GitHub Copilot throughout the project. They helped a lot:

**Where AI helped:**
- Generated boilerplate code (saved me from typing lots of repetitive stuff)
- Explained Go concepts I didn't understand
- Helped debug issues (especially concurrency bugs)
- Wrote SQL queries
- Suggested better patterns

**Where I didn't use AI:**
- The core logic for category grouping (I wanted to think through this myself)
- Architecture decisions (AI can suggest but you need to decide)
- Understanding the requirements (obviously)

**Time savings:** Probably saved 40-50% of the time. Without AI, this would've taken 30+ hours instead of 18-20.

The biggest benefit wasn't just speed - it was learning faster. Instead of reading docs for hours, I could ask questions and get instant explanations.

---