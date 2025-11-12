# Yawn - Technology Stack

## Overview

Yawn uses a modern, polyglot technology stack optimized for performance, developer experience, and scalability. The architecture emphasizes strong backend-frontend coupling through shared type definitions and a Git-first development workflow.

---

## Project Structure

```
yawn/
├── ui/           # Frontend (React + TypeScript)
├── api/          # Backend (Go)
├── shared/       # Shared models and build artifacts
└── library/      # Yawn Python SDK
```

### Shared Directory

The `shared/` directory contains:

- **TypeScript models** generated from Go structs via **gzuidhof/tygo**
- **Frontend build artifacts** (dist) for production embedding
- Cross-platform type definitions ensuring backend-frontend consistency

### Library Directory

The `library/` folder contains the **Yawn Python SDK** for defining assets and pipelines.

---

## Backend Stack

### Language & Runtime

- **Language**: Go 1.25+
- **Architecture**: Domain-Driven Design (DDD)
- **Deployment**: Single binary with microservice-style folder structure
- **Modes**: Standalone (dev) or distributed (production)

### Core Libraries

| Library        | Purpose              | Key Features                                                                |
| -------------- | -------------------- | --------------------------------------------------------------------------- |
| **Echo**       | Web Framework        | High-performance HTTP server, built-in middleware                           |
| **fx**         | Dependency Injection | Compile-time DI, modular architecture                                       |
| **GORM**       | ORM                  | Database abstraction, migrations, relationships                             |
| **Viper**      | Configuration        | Environment variables, config files, secret management                      |
| **Cobra**      | CLI Framework        | Command-line interface, subcommands                                         |
| **Zerolog**    | Logging              | Structured logging, performance optimized                                   |
| **go-git**     | Git Integration      | Allow to manipulate git repository, this will help with implementing GitOps |
| **gofrs/uuid** | UUID                 | Allow for generating UUIDv7 for id                                          |

### Database Strategy

- **Development**: SQLite (embedded, zero-config)
- **Production**: PostgreSQL (recommended) or MySQL
- **Migrations**: GORM auto-migration with version control

### Development Tools

- **Air**: Hot reload for rapid development
- **Task**: Makefile-like command definitions
- **gofmt/golint**: Code formatting and linting
- **testing**: Built-in Go testing framework

### Code Standards

- **Style Guide**: [Uber Go Style Guide](https://github.com/uber-go/guide)
- **Architecture**: DDD with clear domain boundaries. DTO for all of the handler requests and response needed to be implemented to ensure auto openapi spec generation is working
- **Error Handling**: Explicit error returns with structured logging
- **Testing**: Unit tests + integration tests. Use `gotests -all -use_go_cmp -w -parallel <file_name>.go` to generate test for each file. Use "github.com/stretchr/testify/mock" to mock stuff
- **MCP**: Use context7, browsermcp and especially mcp-gopls to help with maintaining and fixing code
- **Test Coverage**: Ensure to have at least test coverage 75% coverage, the main point is to have business tests since this is DDD/Business driven
- **Package Management**: Only use bun and bunx commands - no pnpm, pnpx, npm, npx or yarn
- **Development Servers**: Never run dev mode for API or UI by yourself - always ask user to turn on

---

## Frontend Stack

### Language & Runtime

- **Language**: TypeScript (strict mode)
- **Framework**: React 18+ with functional components
- **Runtime**: Bun (JavaScript runtime)
- **Build Tool**: Vite (fast development and builds)

### Core Libraries

| Library             | Purpose          | Key Features                             |
| ------------------- | ---------------- | ---------------------------------------- |
| **TanStack Router** | Routing          | File-based routing, type-safe navigation |
| **Zustand**         | State Management | Lightweight, simple state management     |
| **TanStack Query**  | Server State     | Caching, synchronization, error handling |
| **React Flow**      | Visual Editor    | DAG visualization, drag-and-drop editing |
| **Tailwind CSS**    | Styling          | Utility-first CSS, rapid prototyping     |
| **Shadcn**          | UI Components    | High-quality, accessible components      |
| **Vitest**          | Testing          | Fast unit testing framework              |

### Development Strategy

- **Development**: Standard frontend dev server with hot reload
- **Production**: Frontend embedded in Go binary via `embed` directive
- **Type Safety**: Generated TypeScript interfaces from Go structs
- **Testing**: Use mcp to aid with the work, use browsermcp to interact with a browser and use serena to make finding component easier, use context7 to find libraries docs
- **Package Management**: Only use bun and bunx commands - no pnpm, pnpx, npm, npx or yarn
- **Development Servers**: Never run dev mode for API or UI by yourself - always ask user to turn on

### UI/UX Approach

- **Design System**: Minimal, clean, professional
- **Components**: Shadcn + custom components
- **Responsive**: Mobile-first responsive design
- **Accessibility**: WCAG 2.1 AA compliance

---

## Python SDK Stack

### Language & Runtime

- **Language**: Python 3.11+
- **Package Management**: uv
- **Distribution**: PyPI package alongside releases

### SDK Architecture

- **Asset Definition**: Decorator-based asset creation
- **Pipeline Composition**: Pythonic DAG definition
- **Type Hints**: Full type annotation support

### Key Libraries

- **Pydantic**: Data validation and serialization
- **Click**: CLI tooling for local development
- **Requests**: HTTP client for API communication
- **Polars**: Blazingly fast dataframe library

---

## Development Toolchain

### Environment Management

- **mise**: Universal toolchain manager (Go, Bun, Air, Task)
- **Docker**: Containerization for consistent environments
- **Task**: Command runner for development workflows

### Code Generation

- **gzuidhof/tygo**: TypeScript interfaces from Go structs
- **OpenAPI**: Auto-generated API documentation
- **Swagger UI**: Interactive API documentation

### Quality Assurance

- **ESLint/Prettier**: JavaScript/TypeScript formatting
- **gofmt/goimports**: Go code formatting
- **Black/Flake8**: Python code formatting and linting
- **Pre-commit hooks**: Automated code quality checks

### CI/CD Pipeline

- **GitHub Actions**: Continuous integration and deployment
- **Automated Testing**: Multi-language test suite
- **Security Scanning**: Dependency vulnerability scanning
- **Release Automation**: Semantic versioning and changelog generation

---

## Production Deployment

### Build Process

1. **Type Generation**: Generate TypeScript from Go structs
2. **Frontend Build**: Build React application with Vite
3. **Asset Embedding**: Embed frontend assets in Go binary
4. **Binary Compilation**: Single self-contained executable
5. **Package Creation**: Docker image and system packages

### Deployment Options

- **Single Binary**: Ideal for edge deployment and simple setups
- **Docker Container**: Standard containerized deployment
- **Kubernetes**: Orchestration for high availability
- **Cloud Services**: AWS ECS, Google Cloud Run, Azure Container Instances

### Monitoring & Observability

- **Metrics**: Prometheus-compatible metrics
- **Logging**: Structured JSON logs
- **Tracing**: OpenTelemetry integration
- **Health Checks**: Built-in health endpoint

---

## Integration Architecture

### External Systems

- **Version Control**: Git repositories (GitHub, GitLab, Bitbucket)
- **Authentication**: Google SSO, SAML support
- **Databases**: PostgreSQL, MySQL, SQLite
- **Message Queues**: SQS, SNS, Kafka (future)
- **File Storage**: Local filesystem, S3-compatible storage

### API Design

- **REST**: Primary API pattern with OpenAPI specification
- **GraphQL**: Future consideration for complex queries
- **gRPC**: Internal service communication (microservice mode)

### Security Model

- **Authentication**: JWT tokens with Google SSO integration
- **Authorization**: RBAC with resource-level permissions
- **Data Encryption**: TLS 1.3, at-rest encryption
- **Audit Logging**: Comprehensive activity tracking

---

## Performance Considerations

### Backend Performance

- **Concurrency**: Go goroutines for parallel execution
- **Database**: Connection pooling, query optimization
- **Caching**: In-memory caching for frequently accessed data
- **Resource Management**: Efficient memory and CPU utilization

### Frontend Performance

- **Bundle Size**: Code splitting, lazy loading
- **Runtime**: Vite's optimized build pipeline
- **Rendering**: React memoization, virtualization
- **Network**: HTTP/2, compression, CDN distribution

### Scalability Strategy

- **Horizontal Scaling**: Stateless backend design
- **Database Sharding**: Partitioning for large datasets
- **Caching Layers**: Redis for distributed caching
- **Load Balancing**: Application-aware routing
