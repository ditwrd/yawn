# Yawn Project Overview

## Project Purpose
Yawn is a GitOps CI/CD platform that combines the strengths of Airflow, Dagster, Windmill, AppSmith, and n8n into one cohesive environment. It enables both engineers and non-engineers to collaboratively build DAG pipelines, automations, and dashboards in a shared, intuitive workspace.

The platform introduces an **asset-centric** rather than **task-centric** design philosophy, focusing users on what needs to be accomplished rather than how it should be done.

## Architecture Overview
Yawn uses a modern, polyglot technology stack with clear separation of concerns:

- **Backend**: Go with Domain-Driven Design (DDD) using Echo framework
- **Frontend**: React + TypeScript with Vite 
- **Python SDK**: Asset and pipeline definitions
- **Shared Types**: TypeScript generated from Go structs

### Directory Structure
```
yawn/
├── ui/           # Frontend (React + TypeScript)
├── api/          # Backend (Go with DDD)
├── shared/       # Shared models and build artifacts
└── library/      # Yawn Python SDK
```

## Current Development Status
**Phase 1: Foundation** - Core infrastructure completed:
- ✅ Go API framework with Echo and fx dependency injection
- ✅ Authentication system with JWT tokens and Argon2id hashing  
- ✅ User management with CRUD operations
- ✅ Project management with role-based access control
- ✅ Database layer with GORM and proper relationships
- ✅ Comprehensive test suite with 75%+ coverage
- ✅ Frontend foundation with React + TanStack Router
- ✅ Shadcn UI components integration
- ✅ TypeScript interface generation from Go structs

**In Progress**:
- Asset definition framework
- Python SDK for asset creation  
- Pipeline execution engine
- Visual pipeline editor
- Dashboard system

## Target Users
1. **Power Users (Engineers/Data Engineers)**: Define complex pipelines using Python and Yawn SDK
2. **Non-Power Users (Analysts/Business Users)**: Build dashboards and simple automations using low-code interface