# Task Completion Checklist

## Code Quality Checks (Required for All Tasks)

### Backend Tasks
- [ ] **Tests Pass**: `go test -v ./...` (75% coverage minimum)
- [ ] **Code Formatted**: `gofmt -s -w . && goimports -w .`
- [ ] **Linting Clean**: `golangci-lint run`
- [ ] **Build Success**: `go build ./...`
- [ ] **Types Generated**: `tygo generate` (if models changed)

### Frontend Tasks
- [ ] **Tests Pass**: `bun test`
- [ ] **Type Check**: `tsc --noEmit`
- [ ] **Linting Clean**: `bun run lint`
- [ ] **Format Check**: `bun run format`
- [ ] **Build Success**: `bun run build`

### Integration Tasks
- [ ] **Full Application Builds**: `task build`
- [ ] **Type Generation**: `task gen-types` (if Go models changed)
- [ ] **Cross-component Integration**: Frontend can communicate with backend

## Architecture Compliance

### Domain-Driven Design
- [ ] **Layer Separation**: Business logic in services, data access in repositories, HTTP in handlers
- [ ] **Dependency Injection**: New dependencies properly registered in `app.go`
- [ ] **Interface Segregation**: Repository interfaces defined separately from implementations
- [ ] **Error Handling**: Proper error wrapping and context

### Database & Models
- [ ] **GORM Annotations**: Proper tags for relationships and constraints
- [ ] **UUID v7**: Primary keys use UUID v7
- [ ] **Soft Deletes**: `DeletedAt` field included where appropriate
- [ ] **Timestamps**: `CreatedAt`, `UpdatedAt` fields
- [ ] **Foreign Keys**: Proper relationships defined

### API Design
- [ ] **RESTful Routes**: Proper HTTP methods and resource naming
- [ ] **Authentication**: JWT middleware applied where needed
- [ ] **Authorization**: RBAC checks implemented
- [ ] **Validation**: DTO validation tags present
- [ ] **Error Responses**: Consistent error format

## Testing Requirements

### Coverage Standards
- [ ] **75% Minimum Coverage**: Focus on business logic
- [ ] **Unit Tests**: Service layer thoroughly tested
- [ ] **Integration Tests**: Repository and handler integration
- [ ] **Mock Usage**: Repository interfaces mocked in service tests
- [ ] **Deterministic Tests**: No flaky UUIDs or time dependencies

### Test Quality
- [ ] **Table-Driven Tests**: Multiple scenarios covered
- [ ] **Edge Cases**: Error conditions and boundary cases
- [ ] **Test Naming**: Clear, descriptive test names
- [ ] **Test Data**: Proper setup and teardown

## Security Checklist

### Authentication & Authorization
- [ ] **Password Security**: Argon2id hashing implemented
- [ ] **JWT Security**: Proper token validation and refresh
- [ ] **RBAC**: System-level permissions enforced
- [ ] **Project Permissions**: Project-level roles respected
- [ ] **Input Validation**: All user inputs validated

### Data Security
- [ ] **SQL Injection**: GORM parameterized queries
- [ ] **XSS Prevention**: Proper input sanitization
- [ ] **CSRF Protection**: CSRF tokens where appropriate
- [ ] **Sensitive Data**: No secrets in logs or error messages

## Documentation Requirements

### Code Documentation
- [ ] **Package Docs**: Clear package documentation
- [ ] **Function Comments**: Public functions documented
- [ ] **API Documentation**: Endpoints documented (future: OpenAPI/Swagger)
- [ ] **README Updates**: Usage examples updated if needed

### Type Documentation
- [ ] **TypeScript Types**: Generated from Go structs
- [ ] **Interface Updates**: Frontend types synced with backend
- [ ] **Breaking Changes**: Documented and communicated

## Performance Considerations

### Database Performance
- [ ] **Query Optimization**: Efficient GORM queries
- [ ] **Indexing**: Proper database indexes
- [ ] **N+1 Prevention**: Eager loading where appropriate
- [ ] **Connection Pooling**: Database pool configured

### Application Performance
- [ ] **Memory Usage**: No memory leaks in tests
- [ ] **Response Times**: API endpoints responsive
- [ ] **Concurrent Safety**: Thread-safe code where needed
- [ ] **Resource Cleanup**: Proper cleanup in defer statements

## Pre-Commit Final Checks
1. **Full Test Suite**: `task test`
2. **Complete Build**: `task build`
3. **Type Generation**: `task gen-types`
4. **Linting**: `task lint`
5. **Manual Testing**: Critical paths tested manually
6. **Documentation**: Updated if behavior changed
7. **Git Status**: Only intended changes staged

## Environment-Specific Checks

### Development Environment
- [ ] **Hot Reload**: Working with air
- [ ] **Environment Variables**: Proper development config
- [ ] **Database**: SQLite for local development

### Production Readiness
- [ ] **Environment Config**: Production variables set
- [ ] **Database**: PostgreSQL configuration tested
- [ ] **Logging**: Structured logging configured
- [ ] **Monitoring**: Health checks implemented
- [ ] **Single Binary**: Embedded frontend working

Remember: Quality is everyone's responsibility. Don't skip checks!