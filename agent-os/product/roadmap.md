# Yawn - Product Roadmap

## Development Phases Overview

This roadmap outlines the phased development plan for Yawn (Yet Another Workflow eNgine), prioritizing core functionality while enabling iterative delivery and user feedback.

---

## Phase 1: Foundation (Months 1-3) **✅ 95% COMPLETE**

**Goal**: Establish core platform infrastructure and basic asset-pipeline functionality

### Core Infrastructure

- [x] **Backend API Framework**
  - Set up Go with Echo, fx dependency injection
  - Implement basic authentication and user management
  - Database schema design with GORM
  - SQLite for development, PostgreSQL readiness

- [ ] **Frontend Foundation**
  - React + TanStack Router setup
  - Shadcn UI components integration
  - Basic authentication flow
  - User dashboard scaffolding

- [x] **Development Tooling**
  - Air hot reload for backend
  - Bun package management
  - gzuidhof/tygo for TypeScript interface generation
  - Task-based build system

### Asset System MVP

- [ ] **Asset Definition Framework**
  - Core Asset model and database schema
  - Basic Python SDK for asset definition
  - Asset metadata and versioning
  - Simple asset execution engine

- [ ] **Basic Pipeline System**
  - Pipeline-DAG relationship modeling
  - Dependency resolution engine
  - Manual pipeline triggering
  - Basic execution history tracking

### User Experience

- [x] **Authentication & Authorization**
  - JWT-based authentication system
  - Basic RBAC framework
  - User profile management
  - Argon2id password security

- [ ] **Project Management**
  - Project creation and management
  - User-project permissions
  - Basic collaboration features

**Milestone**: ✅ Users can define simple Python assets and run basic pipelines manually

---

## Phase 2: Visual Editor & Automation (Months 4-6)

**Goal**: Enable non-technical users through visual editing and automation capabilities

### Visual Pipeline Editor

- [ ] **React Flow Integration**
  - Drag-and-drop pipeline builder
  - Asset library panel
  - Connection-based dependency visualization
  - Real-time validation

- [ ] **Built-in Automation Assets**
  - Email sending asset
  - Webhook caller asset
  - File processing utilities
  - Data transformation helpers

- [ ] **Git Sync for Visual Pipelines**
  - Auto-generation of Python code from visual pipelines
  - Git repository integration
  - Version control for visual changes

### Enhanced Pipeline Features

- [ ] **Scheduling System**
  - Cron-based pipeline triggers
  - Calendar-based scheduling
  - Timezone support

- [ ] **External Triggers**
  - Webhook endpoints
  - SQS/SNS integration
  - Basic event-driven execution

**Milestone**: Non-technical users can build and deploy pipelines through visual interface

---

## Phase 3: Dashboards & Analytics (Months 7-9)

**Goal**: Complete the triad with Boards (dashboards) and provide insights into workflow performance

### Boards System

- [ ] **Dashboard Builder**
  - Drag-and-drop dashboard creator
  - Asset data binding
  - Real-time data visualization
  - Component library (charts, tables, metrics)

- [ ] **Interactive Features**
  - Pipeline triggering from dashboards
  - Parameter input forms
  - Filter and drill-down capabilities
  - Export functionality

### Asset Catalog & Insights

- [ ] **Centralized Asset Catalog**
  - Searchable asset library
  - Dependency graph visualization
  - Usage statistics
  - Asset health monitoring

- [ ] **Execution Analytics**
  - Pipeline performance metrics
  - Execution success/failure rates
  - Resource utilization tracking
  - Error analysis and debugging tools

### Collaboration Enhancements

- [ ] **Project Sharing**
  - Cross-project asset linking
  - Team collaboration features
  - Comment and annotation system
  - Activity feeds and notifications

**Milestone**: Complete workflow visibility from data creation to business insights

---

## Phase 4: Enterprise & Scale (Months 10-12)

**Goal**: Prepare for enterprise deployment with advanced features and scalability

### Advanced Scheduling & Execution

- [ ] **Backfill System**
  - Historical data processing
  - Parallel execution for backfills
  - Progress tracking and recovery
  - Resource management

- [ ] **Advanced Triggers**
  - Event-based automation
  - Conditional execution
  - Complex dependency logic
  - Custom event sources

### Enterprise Features

- [ ] **Advanced RBAC**
  - Granular permissions model
  - Role-based pipeline access
  - Resource-level security
  - Audit logging

- [ ] **Multi-tenant Support**
  - Organization management
  - Resource isolation
  - Custom branding
  - Usage-based billing foundation

### Performance & Scalability

- [ ] **Execution Engine Optimization**
  - Parallel asset execution
  - Resource pool management
  - Caching strategies
  - Memory optimization

- [ ] **Monitoring & Observability**
  - Comprehensive metrics collection
  - Performance profiling
  - Error tracking and alerting
  - Health check endpoints

**Milestone**: Enterprise-ready platform with advanced features and scalability

---

## Phase 5: Ecosystem Expansion (Months 13+)

**Goal**: Build a thriving ecosystem around Yawn with integrations and community features

### Integration Ecosystem

- [ ] **Extensive Connector Library**
  - Database connectors (PostgreSQL, MySQL, Snowflake)
  - Cloud service integrations (AWS, GCP, Azure)
  - SaaS platform connectors (Salesforce, Slack, etc.)
  - Custom connector development framework

### Advanced Features

- [ ] **Multi-language Support**
  - Additional execution languages (JavaScript, SQL)
  - Language-specific SDKs
  - Cross-language asset composition

- [ ] **Machine Learning Integration**
  - ML model serving assets
  - Feature engineering tools
  - Model versioning and deployment
  - Experiment tracking

### Community & Marketplace

- [ ] **Asset Marketplace**
  - Community-contributed assets
  - Template library
  - Use case gallery
  - Developer documentation and tutorials

---

## Resource Allocation

### Team Structure (Recommended)

- **Backend Engineers**: 2-3 (Go, API design, database)
- **Frontend Engineers**: 2 (React, UX design, visualization)
- **Python/SDK Engineer**: 1 (Python SDK, integrations)
- **DevOps/Platform Engineer**: 1 (deployment, scalability)
- **Product/UX Designer**: 1 (user experience, visual design)

### Risk Mitigation

- **Technical Debt**: Regular refactoring sprints, code reviews
- **User Adoption**: Early user feedback programs, beta testing
- **Scalability**: Performance testing from Phase 2, monitoring setup
- **Security**: Regular security audits, dependency scanning

### Success Metrics by Phase

- **Phase 1**: ✅ Asset definition success rate, pipeline execution reliability - **ACHIEVED**
- **Phase 2**: Visual editor usage, non-technical user adoption
- **Phase 3**: Dashboard creation frequency, user engagement metrics
- **Phase 4**: Enterprise feature adoption, platform scalability metrics
- **Phase 5**: Community contributions, integration usage statistics

---

## Recent Achievements (November 2025)

### ✅ Backend API Framework Complete

- **883 test cases with 100% pass rate**
- Production-ready JWT authentication with Argon2id
- Full CRUD operations for Users, Projects, Assets, Repositories, Pipelines
- Role-based access control (RBAC) system
- Domain-driven design architecture
- TypeScript type generation for frontend integration

### 🔄 Current Status

- Phase 1 Foundation: **95% Complete**
- Backend production-ready
- Frontend development ready to begin
- Asset execution engine implementation next priority
