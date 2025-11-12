# Yawn Development Conventions

## Code Standards

### Go Backend
- **Style Guide**: Follow Uber Go Style Guide
- **Architecture**: Domain-Driven Design with clear layer separation
- **Error Handling**: Structured error returns with proper error wrapping
- **Logging**: Use structured logging with zerolog
- **Testing**: Table-driven tests with comprehensive scenarios, 75% coverage requirement
- **Naming**: Use Go conventions - PascalCase for exported, camelCase for unexported
- **Dependencies**: All dependencies wired through fx providers in app.go

### Frontend (React/TypeScript)
- **Components**: Functional components with hooks
- **Type Safety**: Use TypeScript interfaces generated from Go structs
- **Styling**: Tailwind CSS utility classes with Shadcn components
- **State**: TanStack Query for server state, minimal local state
- **Code Style**: Biome formatter and linter

## Architecture Patterns

### Domain-Driven Design Layers
1. **Domain Layer** (`internal/domain/`):
   - Models: Core entities with GORM annotations
   - Services: Business logic and validation
   - Repositories: Data access interfaces

2. **Infrastructure Layer** (`internal/infrastructure/`):
   - Database: GORM setup
   - Web: Echo framework and middleware
   - Logger: Zerolog setup

3. **Interface Layer** (`internal/interfaces/`):
   - Handlers: HTTP request handlers
   - DTOs: Request/response data transfer objects

4. **App Layer** (`internal/app/`):
   - Dependency injection with uber-go/fx
   - Application bootstrap and lifecycle

### Key Design Patterns
- **Repository Pattern**: Interface-based data access
- **Service Layer**: Business logic encapsulation
- **DTO Pattern**: Request/response data transfer
- **Dependency Injection**: Compile-time DI with fx
- **Asset-Centric Model**: Focus on outputs rather than tasks

## Database Conventions
- **Primary Keys**: UUID v7 for all entities
- **Soft Deletes**: GORM's DeletedAt field
- **Timestamps**: CreatedAt, UpdatedAt automatic
- **Naming**: Snake case for database columns
- **Relationships**: Proper foreign key constraints

## Testing Strategy
- **Backend**: Table-driven tests with testify/mock
- **Coverage**: 75% minimum, focus on business logic
- **Test Data**: Deterministic UUIDs to avoid flakiness
- **Mock Strategy**: Mock repository interfaces
- **Frontend**: Vitest + Testing Library

## API Design
- **REST**: RESTful endpoints with proper HTTP methods
- **Authentication**: JWT with access/refresh tokens
- **Authorization**: RBAC for system permissions, project-level roles
- **Validation**: Struct validation on DTOs
- **Error Responses**: Consistent error format with proper status codes

## Development Workflow
1. **Feature Development**: Add models → repositories → services → handlers → DTOs
2. **Testing**: Write comprehensive tests for business logic
3. **Type Generation**: Use tygo to generate TypeScript interfaces
4. **Integration**: Update frontend to use new types
5. **Quality**: Run linters and formatters before commits