# Yawn - Yet Another Workflow Engine

[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org)
[![React Version](https://img.shields.io/badge/React-18+-blue.svg)](https://reactjs.org)
[![License](https://img.shields.io/badge/License-GPLv3-green.svg)](LICENSE)

A unified platform that combines the strengths of Airflow, Dagster, Windmill, AppSmith, and n8n into one cohesive environment. Yawn enables both engineers and non-engineers to collaboratively build DAG pipelines, automations, and dashboards in a shared, intuitive workspace.

## 🎯 Core Vision

Yawn serves as a **central hub** where users can build, automate, and visualize workflows — all within one interface. The platform introduces an **asset-centric** rather than **task-centric** design philosophy, reducing cognitive load by focusing users on _what_ needs to be accomplished rather than _how_ it should be done.

### Asset-Centric Philosophy

Unlike traditional task-oriented models where dependencies are tightly coupled, Yawn's asset-oriented model treats each asset as a concrete anchor that defines what needs to exist before subsequent steps can occur.

**Example Workflow:**

1. Ingest data →
2. Generate a report →
3. Send an email with that report

Each asset represents a tangible output that downstream assets depend on, making complex workflows more manageable and changes less disruptive.

## 🏗️ Architecture Overview

Yawn uses a modern, polyglot technology stack optimized for performance, developer experience, and scalability:

```
yawn/
├── ui/           # Frontend (Tanstack)
├── api/          # Backend (Go with Domain-Driven Design)
├── shared/       # Shared models and build artifacts
└── library/      # Yawn Python SDK
```

### Technology Stack

**Backend (Go)**:

- **Framework**: Echo with uber-go/fx dependency injection
- **Database**: GORM ORM with SQLite (dev) / PostgreSQL (prod)
- **Authentication**: JWT with Argon2id password hashing
- **Architecture**: Domain-Driven Design with clear layer separation
- **Deployment**: Single binary with embedded frontend

**Frontend (React)**:

- **Framework**: React 18+ with TypeScript
- **Routing**: TanStack Router with file-based routing
- **State**: Zustand for lightweight state management
- **UI**: Shadcn components with Tailwind CSS
- **Build**: Vite for fast development and optimized builds

**Python SDK**:

- **Asset Definition**: Decorator-based asset creation
- **Pipeline Composition**: Pythonic DAG definition
- **Type Safety**: Full type annotation support

## 🚀 Quick Start

### Prerequisites

- Go 1.25+
- Node.js/Bun (for frontend development)
- Python 3.11+ (for SDK development)

### Development Setup

1. **Clone the repository**:

   ```bash
   git clone https://github.com/ditwrd/yawn.git
   cd yawn
   ```

2. **Install dependencies using mise**:

   ```bash
   mise install
   ```

3. **Start the backend**:

   ```bash
   cd api
   go mod tidy
   go run main.go serve
   # Or with hot reload:
   air
   ```

4. **Start the frontend** (in another terminal):
   ```bash
   cd ui
   bun install
   bun dev
   ```

The API server will be available at `http://localhost:8080` and the frontend at `http://localhost:5173`.

## 📋 Current Status

We're currently in **Phase 1: Foundation** of our development roadmap. The following core infrastructure has been implemented:

### ✅ Completed Features

**Backend Infrastructure**:

- [x] Go API framework with Echo and fx dependency injection
- [x] Authentication system with JWT tokens and Argon2id hashing
- [x] User management with CRUD operations
- [x] Project management with role-based access control
- [x] Database layer with GORM and proper relationships
- [x] Comprehensive test suite with 75%+ coverage
- [x] Development tooling with hot reload and linting

**Frontend Foundation**:

- [x] React + TanStack Router setup
- [x] Shadcn UI components integration
- [x] TypeScript interface generation from Go structs
- [x] Basic project structure and build pipeline

### 🚧 In Progress

- [ ] Asset definition framework
- [ ] Python SDK for asset creation
- [ ] Pipeline execution engine
- [ ] Visual pipeline editor
- [ ] Dashboard system

### 📅 Next Milestones

**Phase 2 (Months 4-6)**: Visual Editor & Automation

- Drag-and-drop pipeline builder with React Flow
- Built-in automation assets (email, webhooks, file processing)
- Git sync for visual pipelines

**Phase 3 (Months 7-9)**: Dashboards & Analytics

- Dashboard builder with real-time data visualization
- Asset catalog and execution analytics
- Enhanced collaboration features

## 👥 Target Users

### Power Users (Engineers / Data Engineers)

- **Primary users** who define complex Pipelines using **Python** and the **Yawn SDK**
- Push code to Git, which Yawn parses and syncs automatically
- Build sophisticated data workflows and integrations

### Non-Power Users (Analysts / Business Users)

- **Secondary users** who build **dashboards**, **data visualizations**, and **simple automations**
- Use the low-code interface without writing code
- Create business-focused dashboards and reports

## 🛠️ Development

### Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test file
go test -v ./internal/domain/services

# Generate tests for a file
gotests -all -use_go_cmp -w -parallel <file_name>.go
```

### Code Quality

```bash
# Format code
gofmt -s -w .
goimports -w .

# Run linter
golangci-lint run

# Run linter on specific files
golangci-lint run ./internal/domain/services ./internal/interfaces/handlers
```

### Frontend Development

```bash
# Install UI components (Shadcn)
pnpx shadcn@latest add <component_name>

# Start development server
bun dev

# Build for production
bun build
```

## 📚 Documentation

- **[Architecture Guide](CLAUDE.md)** - Detailed technical documentation for developers
- **[Technology Stack](agent-os/product/tech-stack.md)** - Comprehensive tech stack overview
- **[Product Mission](agent-os/product/mission.md)** - Product vision and value proposition
- **[Development Roadmap](agent-os/product/roadmap.md)** - Phased development plan

## 🤝 Contributing

We welcome contributions! Please see our development guidelines:

- Follow the Uber Go Style Guide
- Maintain 75% test coverage with focus on business logic
- Use structured logging with zerolog
- Implement proper error handling and validation
- Respect domain boundaries in the DDD architecture

### Development Workflow

1. Fork the repository
2. Create a feature branch
3. Implement your changes with tests
4. Ensure all tests pass and code follows style guidelines
5. Submit a pull request with a clear description

## 📄 License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- **Documentation**: [Architecture Guide](CLAUDE.md)
- **Issues**: [GitHub Issues](https://github.com/ditwrd/yawn/issues)
- **Roadmap**: [Development Plan](agent-os/product/roadmap.md)
- **Tech Stack**: [Technology Overview](agent-os/product/tech-stack.md)

---

**Built with ❤️ for data engineers and business users alike**
