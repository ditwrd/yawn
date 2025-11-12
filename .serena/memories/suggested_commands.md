# Yawn Development Commands

## Setup & Installation

### Environment Setup
```bash
# Install all tools using mise
mise install

# Install dependencies for each component
cd api && go mod tidy
cd ../ui && bun install  
cd ../library && uv sync
```

## Backend Development

### Running the API
```bash
# Start development server with hot reload
task dev                    # From project root
# OR
cd api && air              # From api directory

# Start server directly
cd api && go run main.go serve

# Build binary
cd api && go build -o ../dist/yawn .
```

### Testing (Backend)
```bash
# Run all tests
task test                   # From project root
# OR
cd api && go test -v ./...

# Run specific test file
cd api && go test -v ./internal/domain/services

# Run single test
cd api && go test -v ./internal/domain/services -run TestProjectService_Create

# Generate tests for a file
cd api && gotests -all -use_go_cmp -w -parallel <file_name>.go

# Run tests with coverage
task test-coverage
# OR
cd api && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
```

### Code Quality (Backend)
```bash
# Format code
gofmt -s -w .
goimports -w .

# Run linter
task lint                   # From project root
# OR
cd api && golangci-lint run

# Run linter on specific files
cd api && golangci-lint run ./internal/domain/services ./internal/interfaces/handlers

# Run all quality checks
gofmt -s -w . && goimports -w . && golangci-lint run
```

### Dependencies
```bash
cd api && go mod tidy
cd api && go mod download
```

## Frontend Development

### Running the Frontend
```bash
# Start development server
task dev-ui                 # From project root
# OR
cd ui && bun dev

# Build for production
cd ui && bun run build
```

### Frontend Testing & Quality
```bash
cd ui && bun test           # Run tests
cd ui && bun run lint       # Run linter
cd ui && bun run format     # Format code
cd ui && bun run check      # Run all checks
```

### UI Components
```bash
# Install Shadcn components
cd ui && pnpx shadcn@latest add <component_name>
```

## Type Generation

### Generate TypeScript from Go
```bash
task gen-types              # From project root
# OR
cd api && tygo generate
```

This generates TypeScript interfaces in shared/ from Go model structs and DTOs.

## Build & Deployment

### Full Build
```bash
task build                  # Build complete application
# This runs: gen-types → build-ui → build-api
```

### Individual Builds
```bash
task build-api              # Build Go binary
task build-ui               # Build React frontend
```

### Database
```bash
task db-migrate             # Run database migrations
# OR
cd api && go run main.go migrate
```

## Utilities

### Clean Build Artifacts
```bash
task clean                  # Clean all build artifacts
```

### Git
```bash
git status                  # Check status
git add .                   # Stage changes
git commit -m "message"     # Commit changes
git push                    # Push to remote
```

### File Operations
```bash
ls -la                      # List files
find . -name "*.go"         # Find Go files
grep -r "pattern" .         # Search in files
```

## Entry Points

### Main Applications
- **API Server**: `api/main.go serve` (port 8080)
- **Frontend Dev**: `ui/bun dev` (port 3000)
- **Production**: Single binary at `dist/yawn`

### Development URLs
- API: http://localhost:8080
- Frontend: http://localhost:3000
- API Docs: http://localhost:8080/docs (when implemented)

## MCP Integration
The project supports several MCP servers for enhanced development:
- **context7**: Documentation lookup
- **browsermcp**: Web automation 
- **mcp-gopls**: Go language server features
- **serena**: Semantic code navigation and editing

## Quality Checklist (Before Commit)
1. Run tests: `task test`
2. Check coverage: `task test-coverage` 
3. Run linter: `task lint`
4. Format code: `gofmt -s -w . && goimports -w .`
5. Generate types: `task gen-types`
6. Build application: `task build`