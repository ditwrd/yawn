# Yawn - Product Mission & Vision

## Product Overview

**Yawn (Yet Another Workflow eNgine)** is a unified platform that combines the strengths of Airflow, Dagster, Windmill, AppSmith, and n8n into one cohesive environment. It enables both engineers and non-engineers to collaboratively build DAG pipelines, automations, and dashboards/apps in a shared, intuitive workspace.

## Core Vision

Yawn serves as a **central hub** where users can build, automate, and visualize workflows — all within one interface. The platform introduces an **asset-centric** rather than **task-centric** design philosophy, reducing cognitive load by focusing users on _what_ needs to be accomplished rather than _how_ it should be done.

## Asset-Centric Philosophy

Unlike traditional task-oriented models where dependencies are tightly coupled, Yawn's asset-oriented model treats each asset as a concrete anchor that defines what needs to exist before subsequent steps can occur.

**Example Workflow:**

1. Ingest data →
2. Generate a report →
3. Send an email with that report

In this model, each asset (ingested data, generated report, email sent) represents a tangible output that downstream assets depend on, making complex workflows more manageable and changes less disruptive.

## Target User Segments

### Power Users (Engineers / Data Engineers)

- **Primary users** who define complex Pipelines using **Python** and the **Yawn SDK**
- Push code to Git, which Yawn parses and syncs automatically
- Build sophisticated data workflows and integrations
- Focus on system architecture and optimization

### Non-Power Users (Analysts / Business Users / Executives)

- **Secondary users** who build **dashboards**, **data visualizations**, and **simple automations**
- Use the low-code interface without writing code
- Create business-focused dashboards and reports
- Leverage pre-built automation components

## Key Value Propositions

1. **Unified Platform**: One tool for workflows, automations, and dashboards
2. **Asset-Centric Design**: Focus on outputs rather than implementation details
3. **Collaborative Environment**: Bridge between technical and non-technical users
4. **GitOps Integration**: Code-based workflow definitions with visual editing
5. **Extensible Architecture**: Python SDK for custom logic and integrations

## Product Goals

### Short-term

- Establish core asset-pipeline-board functionality
- Enable basic GitOps workflow with the gitops service in backend
- Provide intuitive visual editor for non-technical users
- Support local development and testing

### Long-term

- Become the de-facto platform for workflow automation
- Enable enterprise-scale collaboration across teams
- Provide extensive library of pre-built assets and integrations
- Support multi-language execution environments

## Success Metrics

- **Developer Adoption**: Number of active Git repositories using Yawn SDK
- **User Engagement**: Daily active users across both technical and non-technical segments
- **Workflow Complexity**: Average number of assets per pipeline
- **Collaboration**: Number of shared projects and cross-team workflows
- **Platform Reliability**: Uptime, execution success rates, and system performance

## Competitive Differentiation

1. **Asset-Centric vs Task-Centric**: Focus on outputs rather than processes
2. **Dual Interface**: Both code-based and visual editing capabilities
3. **Unified Ecosystem**: Single platform for pipelines, automations, and dashboards
4. **Modern Tech Stack**: Go backend, React frontend, Python SDK
5. **Git-First Approach**: Native GitOps integration from day one
