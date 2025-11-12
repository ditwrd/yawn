# Yawn Technology Stack

## Backend (Go)
- **Language**: Go 1.25.4
- **Framework**: Echo with uber-go/fx dependency injection
- **Database**: GORM ORM with SQLite (dev) / PostgreSQL (prod)
- **Authentication**: JWT with Argon2id password hashing
- **Architecture**: Domain-Driven Design with clear layer separation
- **UUID**: gofrs/uuid v4.4.0
- **Logging**: zerolog structured logging
- **CLI**: Cobra CLI framework
- **Configuration**: Viper for configuration management

## Frontend (React)
- **Language**: TypeScript 5.7.2
- **Framework**: React 19.2.0
- **Build Tool**: Vite 7.1.7
- **Routing**: TanStack Router with file-based routing
- **State Management**: Zustand (planned) + TanStack Query
- **UI Components**: Shadcn components with Tailwind CSS 4.0.6
- **Package Manager**: Bun 1.3.1
- **Linting/Formatting**: Biome 2.2.4
- **Testing**: Vitest + Testing Library

## Python SDK
- **Language**: Python 3.12+
- **Package Manager**: uv 0.9.7
- **Purpose**: Asset Definition and Pipeline Composition with decorator-based asset creation

## Development Tools
- **Task Runner**: Task (Taskfile.yml)
- **Hot Reload**: air for Go development
- **Code Quality**: golangci-lint 2.6.1
- **Type Generation**: tygo for Go-to-TypeScript conversion
- **Environment Management**: mise for tool versioning

## Key Dependencies
### Go Backend
- github.com/labstack/echo/v4 (web framework)
- go.uber.org/fx (dependency injection)
- gorm.io/gorm (ORM)
- github.com/golang-jwt/jwt/v5 (JWT tokens)
- github.com/rs/zerolog (logging)
- github.com/spf13/cobra (CLI)
- github.com/spf13/viper (config)

### Frontend  
- @tanstack/react-router (routing)
- @tanstack/react-query (server state)
- class-variance-authority (UI variants)
- clsx + tailwind-merge (styling)
- lucide-react (icons)

## Architecture Patterns
- **Dependency Injection**: All dependencies wired through fx providers
- **Domain-Driven Design**: Clear layer separation (Domain, Infrastructure, Interface, App)
- **Asset-Centric Model**: Focus on what needs to exist rather than task sequences
- **Type Safety**: Generated TypeScript interfaces from Go structs
- **Single Binary**: Frontend embedded in Go binary for production deployment