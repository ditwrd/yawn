# Technology Stack

## Project Type
Fully microservice architecture with both backend and frontend components, designed for scalability and independent deployment of each service.

## Core Technologies

### Primary Language(s)
- **TypeScript**: Strict type safety for frontend micro-apps and dashboard components
- **Go**: High-performance backend services with concurrency support for workflow execution
- **Python**: Script execution runtime for custom workflow nodes with package management

### Key Dependencies/Libraries

#### Frontend Stack
- **React**: Component library for building interactive workflow builder and dashboard editor
- **Vite**: Fast build tool enabling hot reload for rapid micro-frontend development
- **Bun**: JavaScript runtime and package manager providing improved performance for local development
- **TanStack Query**: Server state management for real-time workflow execution status and dashboard data
- **shadcn/ui**: Customizable component library adapted for Yawn's minimalist interface
- **Tailwind CSS**: Utility-first CSS framework enabling rapid UI iteration and consistent styling
- **@xyflow/react**: Visual workflow canvas for creating node-based automation flows
- **@measured/puck**: Low-code dashboard editor for creating custom UIs from workflow data
- **Microsoft Monaco Editor**: Web-based IDE with syntax highlighting and Git integration for script editing
- **@micro-zoe/micro-app**: Micro-frontend framework enabling independent development of Scripts, Workflows, and Dashboards

#### Backend Stack
- **Echo**: High-performance web framework for building RESTful APIs and WebSocket connections
- **Cobra**: CLI library for creating command-line tools for deployment and management
- **Viper**: Configuration management supporting environment variables and config files
- **go.uber.org/fx**: Dependency injection framework for modular service architecture
- **GORM**: ORM with automatic migrations for workflow, script, and user data persistence
- **goqite**: Queue service for asynchronous workflow execution supporting both SQLite and PostgreSQL
- **nsjail**: Security sandbox for isolating user script execution in containers
- **tygo**: Type generation tool converting Go structs to TypeScript for end-to-end type safety

#### Development Tools
- **moonrepo.dev**: Monorepo management with templating for rapid microservice creation
- **mise**: Development environment management for consistent tooling across languages
- **uv**: Python package manager for handling script dependencies in isolated environments
- **Air**: Live reload tool for hot compilation during Go microservice development

### Application Architecture
- **Simplified Monorepo Structure**: 
  - **api**: DDD-based microservices for authentication, workflows, scripts, and dashboards
  - **web**: Core frontend and micro-frontends (Scripts, Workflow, Dashboard) with independent deployment
  - **templates**: Reusable API microservice, micro-frontend, and component templates
- **Microservices**: Independent backend and frontend services enabling scalable deployment
- **Domain-Driven Design (DDD)**: Clean separation between business logic (workflows, scripts) and infrastructure
- **Type-First Development**: Go types as source of truth with auto-generated TypeScript types
- **Event-Driven Architecture**: Queue-based communication for asynchronous workflow execution
- **Security-First Design**: Isolated execution environments using nsjail for user scripts

### Data Storage
- **Primary Storage**: PostgreSQL for production, SQLite for local development
- **Data Formats**: JSON for API communication, Protocol Buffers for internal service communication
- **File Storage**: Embedded static assets with support for cloud storage integration
- **Puck Editor Data**: Safe storage in designated database with proper validation and serialization

### External Integrations
- **APIs**: GitHub for script integration, Google Workspace for SSO, AWS S3 for file storage, Google Sheets for data
- **Protocols**: HTTP/REST for external APIs, WebSocket for real-time workflow updates, gRPC for internal services
- **Authentication**: OAuth 2.0 and Google SSO with JWT tokens and RBAC supporting Google Workspace groups
- **Group Management**: Native group system with Google Workspace synchronization for automated user management

### Monitoring & Dashboard Technologies
- **Dashboard Framework**: React primarily with @micro-zoe/micro-app for micro-frontend management
- **Real-time Communication**: WebSocket for live updates and collaboration
- **Visualization Libraries**: shadcn/ui charts for data visualization, custom components for workflow visualization
- **State Management**: TanStack Query for server state with progressive enhancement approach ensuring functionality without JavaScript

## Development Environment

### Build & Development Tools
- **Build System**: moonrepo.dev with project-specific configurations
- **Package Management**: Bun for frontend, go mod for backend, uv for Python scripts
- **Development Workflow**: Hot reload for frontend, Air for live Go compilation, watch mode for development
- **Tooling Standardization**: mise for managing multiple language versions and development tools

### Code Quality Tools
- **Static Analysis**: Biome for JavaScript/TypeScript, golangci-lint for Go, Python type checkers
- **Formatting**: Biome for consistent code formatting across all languages
- **Testing Framework**: Playwright for end-to-end testing, Jest for frontend, Go testing package, pytest for Python
- **Documentation**: Type-driven documentation from Go structs

### Version Control & Collaboration
- **VCS**: Git with GitHub integration
- **Branching Strategy**: Trunk-based development for continuous integration and delivery
- **Code Review Process**: Pull requests with automated checks and manual review

### Dashboard Development
- **Live Reload**: Hot module replacement with Vite for rapid dashboard development
- **Port Management**: Configurable ports with development defaults for micro-frontend testing
- **Multi-Instance Support**: Independent development of Scripts, Workflows, and Dashboards using @micro-zoe/micro-app

## Deployment & Distribution
- **Target Platforms**: Linux (AMD64 and ARM64), macOS and Windows for local development
- **Distribution Method**: Single binary deployments, Docker containers
- **Installation Requirements**: Docker or local Go/Node.js development environment
- **Update Mechanism**: Rolling updates for containerized deployments, binary replacement for standalone

## Technical Requirements & Constraints

### Performance Requirements
- **Response Time**: <100ms for API responses, <1s for workflow execution start
- **Throughput**: Support for 1000+ concurrent workflow executions
- **Memory Usage**: Optimized for container deployment with resource limits
- **Startup Time**: <5s cold start, <500ms warm start for services

### Compatibility Requirements
- **Platform Support**: Linux for production, macOS and Windows for local development
- **Architecture Support**: Full compatibility with both AMD64 and ARM64 for flexible deployment
- **Dependency Versions**: Stable versions with security patches and regular updates
- **Standards Compliance**: OpenAPI 3.0 for APIs, OAuth 2.0 for authentication

### Security & Compliance
- **Security Requirements**: Sandboxed script execution, encrypted data at rest and in transit
- **Compliance Standards**: GDPR-ready design, ISO standards, SOC2 compliance, audit logging for all actions
- **Threat Model**: Isolated execution environments, input validation, rate limiting

### Scalability & Reliability
- **Expected Load**: 100+ concurrent users, 1000+ daily workflow executions
- **Availability Requirements**: 99.9% uptime, graceful degradation
- **Growth Projections**: Horizontal scaling with container orchestration

## Technical Decisions & Rationale

### Decision Log
1. **Go for Backend Services**: Chosen for performance, concurrency, and single-binary deployment. The language's strong typing, garbage collection, and built-in concurrency primitives make it ideal for microservices. Alternatives considered: Node.js (lower performance), Python (slower execution).
2. **TypeScript Frontend**: Type safety and modern React ecosystem with latest features. The strict type system helps catch errors early and improves developer productivity. Alternatives considered: vanilla JavaScript (no type safety), PureScript (too niche).
3. **Monorepo with moonrepo.dev**: Simplified dependency management and code sharing with excellent templating capabilities. The tool's project-based approach fits well with microservice architecture. Alternatives considered: separate repos (complex dependency management), Nx (overly complex).
4. **GORM with PostgreSQL**: Battle-tested ORM with good Go integration and automatic migrations. Provides excellent developer experience while maintaining performance. Alternatives considered: sqlx (more manual), raw SQL (error-prone).
5. **Micro-frontend Architecture**: Independent development cycles for different features using @micro-zoe/micro-app. This approach allows teams to work independently while maintaining a cohesive user experience. Alternatives considered: monolithic frontend (coupled development), multiple SPAs (complex routing).
6. **nsjail for Security**: Proven sandbox technology for untrusted code execution. Provides strong isolation while maintaining good performance. Alternatives considered: Docker containers (heavier), gVisor (more complex).
7. **goqite for Queue**: Persistent queue solution supporting both SQLite and PostgreSQL. Eliminates data loss concerns during restarts and provides flexibility for different deployment scenarios. Alternatives considered: in-memory queues (data loss risk), Redis (additional infrastructure).
8. **@micro-zoe/micro-app**: Simplifies micro-frontend management with webcomponent-like rendering. Reduces complexity compared to other micro-frontend solutions while maintaining flexibility. Alternatives considered: single-spa (more complex), Module Federation (build complexity).
9. **Air for Go Development**: Provides excellent live reload experience for Go development. Streamlines the development workflow with automatic compilation and restart. Alternatives considered: manual builds (slower), other file watchers (less feature-rich).
10. **mise for Tooling Management**: Unified approach to managing multiple language versions and development tools. Replaces multiple version managers with a single, consistent interface. Alternatives considered: asdf (slower), individual version managers (fragmented).
11. **Biome for Code Quality**: Modern, fast toolchain for JavaScript/TypeScript providing both linting and formatting. Offers better performance and developer experience compared to traditional tools. Alternatives considered: ESLint + Prettier (slower, more configuration).
12. **Trunk-based Development**: Enables continuous integration and delivery with reduced merge conflicts. Simplifies the development workflow while maintaining code quality through automated checks. Alternatives considered: Git Flow (more complex, slower integration).

## Known Limitations
- **Python-only Script Execution**: Initial limitation to simplify security model. Impact: restricts language choice for custom nodes. Future solution: expand to other runtimes with appropriate sandboxing.
- **Complex Group Synchronization**: Native group system requires periodic synchronization with external providers. Impact: potential synchronization delays. Future solution: real-time synchronization with conflict resolution.
- **Single Database Instance**: Current design assumes single database per deployment. Impact: limits horizontal scaling for write operations. Future solution: implement read replicas and sharding.
- **Limited Third-party Integrations**: Initial focus on core integrations. Impact: may require custom development for niche services. Future solution: community-driven integration ecosystem.
- **Micro-frontend Complexity**: Managing multiple independent frontends adds operational complexity. Impact: increased deployment and monitoring overhead. Future solution: improved tooling and automation for micro-frontend management.