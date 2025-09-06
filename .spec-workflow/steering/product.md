# Product Overview

## Product Purpose
Yawn is an open-source platform as an alternative to various developer infrastructure tools, providing comprehensive solutions for internal tools, data pipelines, and low-code applications. The platform combines capabilities from Airflow/Dagster for data pipeline orchestration, n8n for low-code workflow automation, and Appsmith for low-code UI building, creating a unified solution for both technical and non-technical users.

## Target Users
- **Non-engineers**: Business users, C-level executives (marketing, sales), and analysts who need to create workflows and dashboards without coding
- **Engineers**: Developers who want to create custom workflow nodes, scripts, and integrate with external systems
- **Data Engineers**: Teams needing data pipeline orchestration and monitoring
- **DevOps/IT teams**: Responsible for deployment, security, and infrastructure management

## Key Features

1. **Scripts**: Code repository with Git-like branching, Web IDE (VSCode), GitHub integration, and RBAC for production deployment
2. **Workflows**: Visual workflow builder with comprehensive battery-included nodes similar to n8n's core functionality, including scheduled jobs, webhooks, logic nodes, and integration capabilities
3. **Dashboard**: Low-code UI builder following Appsmith standards for creating dashboards from script and workflow data with publishing controls

## Business Objectives
- Provide an open-source alternative to commercial workflow and data pipeline platforms
- Enable non-technical users to build automated workflows and dashboards without coding
- Give engineers full control over custom node development with scripts or Docker images
- Support both local development and cloud deployment
- Facilitate data-driven decision making through customizable dashboards
- Unify data pipeline, workflow automation, and UI building in a single platform

## Success Metrics
- **User Adoption**: Number of active non-engineer users creating workflows and dashboards
- **Developer Productivity**: Time saved in creating and deploying custom workflow nodes
- **Platform Stability**: Uptime and reliability of workflow and data pipeline execution
- **Self-hosting Adoption**: Number of successful self-hosted deployments
- **Integration Usage**: Frequency of third-party service integrations

## Product Principles
1. **Accessibility**: Non-engineers can use the platform without seeing code or setting up local environments
2. **Developer Control**: Engineers have full control over custom node development and deployment
3. **Minimalist Design**: Clean, frictionless interface avoiding unnecessary complexity
4. **Security**: Isolated execution environments for scripts and workflows
5. **Flexibility**: Support for both local development and cloud deployment
6. **Observability**: Comprehensive monitoring and logging for all workflows and data pipelines

## Monitoring & Visibility
- **Dashboard Type**: Web-based low-code UI builder following Appsmith standards
- **Real-time Updates**: WebSocket for live workflow execution status and data pipeline monitoring
- **Key Metrics Displayed**: Workflow execution logs, error rates, script performance, data pipeline status
- **Sharing Capabilities**: Public/private publishing with RBAC controls

## Future Vision

### Potential Enhancements
- **Remote Access**: Tunnel features for sharing dashboards with external stakeholders
- **Analytics**: Historical trends, performance metrics, and optimization insights
- **Collaboration**: Multi-user support, commenting, and real-time collaboration on workflows
- **Extended Runtime Support**: Additional programming languages beyond Python
- **Advanced Integrations**: More third-party service connectors
- **Enterprise Features**: Advanced RBAC, audit logs, and compliance reporting
- **Enhanced Data Pipeline**: More sophisticated data transformation and orchestration capabilities