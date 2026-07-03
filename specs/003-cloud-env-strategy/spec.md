# Feature Specification: Cloud & Environments Strategy

**Feature Branch**: `003-cloud-env-strategy`

**Created**: 2026-07-03

**Status**: Draft

**Input**: User description: "Define the cloud and environments strategy for the project. As an operations team, we want a foundation that specifies the target cloud provider, the environment topology (dev, staging, production), the CI/CD approach, secrets and configuration management, and a high-level cost strategy. Describe how infrastructure is provisioned and promoted across environments using IaC. Do not implement infrastructure yet — define the strategy and constraints only."

## Clarifications

### Session 2026-07-03

- Q: Environment isolation strategy — separate AWS accounts vs VPCs within single account? → A: VPCs within single account
- Q: Compute resource strategy — EC2, ECS Fargate, or Lambda? → A: ECS Fargate for backend (better for AI workloads: no time/payload limits, persistent connections), S3 + CloudFront for frontend
- Q: Terraform workspace organization — workspaces vs directory per environment vs hybrid modules? → A: Terraform workspaces (single codebase, workspace-specific state and tfvars)
- Q: GitHub Actions deployment credentials — OIDC federation, IAM user keys, or self-hosted runners? → A: OIDC federation (GitHub assumes AWS IAM role, no long-lived credentials)
- Q: RDS database configuration strategy across environments? → A: Production Multi-AZ (high availability), staging single-AZ minimal instance (cheapest config for MVP)
- Q: Environment topology revision — How many active environments for MVP? → A: Two environments only: Staging (active MVP with cheapest configuration) and Production (IaC-defined but not deployed until alpha release)
- Q: Is Lambda suitable for AI client workloads (bandwidth, latency, thresholds)? → A: No — Switch to ECS Fargate. Lambda has 15-min timeout and 10MB payload limits that risk blocking complex AI conversations; Fargate provides unlimited execution time, no payload constraints, and persistent HTTP connections to Anthropic API

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Infrastructure Provisioning & Environment Setup (Priority: P1)

As a **DevOps engineer**, I need to provision and configure cloud environments (staging and production) using Infrastructure as Code so that environments are reproducible, version-controlled, and auditable.

**Why this priority**: Without defined environments, no deployment can occur. This is the foundation for all subsequent operational work.

**Independent Test**: Can be fully tested by running IaC scripts to provision a new environment from scratch and verifying all required resources are created with correct configurations.

**Acceptance Scenarios**:

1. **Given** no existing infrastructure, **When** IaC scripts are executed for the staging environment, **Then** all required cloud resources (Application Load Balancer, ECS Fargate cluster, RDS database, S3 buckets, VPC, CloudFront) are provisioned and ready for deployment.
2. **Given** the staging environment exists, **When** IaC scripts are executed for production, **Then** production environment is created with higher reliability configuration (Multi-AZ RDS, increased redundancy) but remains unprovisioned/inactive until alpha release approval.
3. **Given** IaC scripts are modified, **When** changes are committed, **Then** CI validates the syntax and plans changes before applying them.

---

### User Story 2 - Secrets & Configuration Management (Priority: P1)

As a **developer or operator**, I need a secure and consistent way to manage secrets (API keys, database passwords, tokens) and environment-specific configuration so that credentials are never exposed in code and each environment uses appropriate values.

**Why this priority**: Security breach risk is unacceptable. Secrets management must be in place before any service deployment.

**Independent Test**: Can be tested by deploying a service that requires secrets and verifying it reads values from the secret manager without hardcoding, and values differ between environments.

**Acceptance Scenarios**:

1. **Given** a secret is stored in the secret manager, **When** the application starts in any environment, **Then** the secret is retrieved securely and never logged or exposed.
2. **Given** environment-specific config values (e.g., API URLs), **When** deploying to staging vs production, **Then** each environment uses its own config without code changes.
3. **Given** a secret needs rotation, **When** the secret is updated in the secret manager, **Then** running services can reload or reference the new value without redeployment.

---

### User Story 3 - CI/CD Pipeline & Deployment Promotion (Priority: P1)

As a **developer**, I need an automated CI/CD pipeline that builds, tests, and deploys code to the staging environment so that changes are deployed in a controlled, repeatable manner, with the ability to promote to production when ready for alpha release.

**Why this priority**: Manual deployments are error-prone and block development velocity. Automated pipelines are essential for rapid iteration and reliability.

**Independent Test**: Can be tested by pushing code changes and verifying the pipeline runs tests, builds artifacts, and deploys to staging automatically, with production deployment capability ready but not automatically triggered.

**Acceptance Scenarios**:

1. **Given** a feature branch is merged to main, **When** the CI pipeline runs, **Then** tests pass, artifacts are built (Docker images), and the code is automatically deployed to the staging environment.
2. **Given** staging deployment is successful and validated, **When** an alpha release is approved, **Then** the same artifacts can be deployed to production (not automatic for MVP; manual trigger only).
3. **Given** a deployment fails, **When** the pipeline detects errors, **Then** the deployment is rolled back automatically and alerts are sent to the team.

---

### User Story 4 - Cost Monitoring & Budget Controls (Priority: P2)

As a **project owner**, I need visibility into cloud spending and cost controls so that the project stays within budget and cost overruns are detected early.

**Why this priority**: Uncontrolled cloud costs can quickly become unsustainable. Cost awareness is necessary but can be implemented after basic infrastructure is operational.

**Independent Test**: Can be tested by reviewing monthly cost reports, verifying budget alerts trigger at thresholds, and confirming staging environment uses the cheapest possible configuration for each AWS service.

**Acceptance Scenarios**:

1. **Given** cloud resources are running in staging, **When** monthly costs are calculated, **Then** spending is broken down by service (ECS, RDS, S3, CloudFront) for transparency.
2. **Given** a budget threshold is defined, **When** actual spending approaches the threshold, **Then** alerts are sent to the project owner before exceeding the budget.
3. **Given** the staging environment is the active MVP environment, **When** resources are configured, **Then** every service uses the minimum viable configuration (smallest RDS instance, minimal ECS task size, S3 Intelligent-Tiering).

---

### User Story 5 - Infrastructure Observability & Health Monitoring (Priority: P2)

As an **operations team member**, I need observability into infrastructure health (resource utilization, uptime, error rates) so that issues are detected and resolved before impacting users.

**Why this priority**: Monitoring is critical for production reliability but can be implemented after basic deployment infrastructure is in place.

**Independent Test**: Can be tested by simulating infrastructure failures and verifying alerts are triggered, dashboards update, and metrics are retained for analysis.

**Acceptance Scenarios**:

1. **Given** infrastructure is running, **When** viewing the monitoring dashboard, **Then** key metrics (CPU, memory, request latency, error rate) are visible for all environments.
2. **Given** a resource threshold is breached (e.g., 80% CPU), **When** the condition persists, **Then** an alert is sent to the operations team.
3. **Given** historical metrics exist, **When** analyzing trends, **Then** data is retained for at least 30 days for troubleshooting and capacity planning.

---

### Edge Cases

- What happens when IaC execution fails mid-apply (partial state)?
- How does the system handle secret retrieval failures at application startup?
- What happens when a deployment to staging fails — does it block production promotion?
- How are infrastructure changes tested before applying to production?
- What happens when the cloud provider has an outage affecting a single region?
- How are costs managed if autoscaling triggers unexpectedly high resource usage?

## Requirements *(mandatory)*

### Functional Requirements

#### Cloud Provider & Region

- **FR-001**: The project MUST use AWS (Amazon Web Services) as the cloud provider for all environments.
- **FR-002**: All production infrastructure MUST reside in a single AWS region (us-east-1) to minimize latency for the primary user base (North America).
- **FR-003**: Non-production environments (dev, staging) MUST use the same region as production to ensure consistency.

#### Environment Topology

- **FR-004**: The project MUST maintain two distinct environments: **staging** and **production**.
- **FR-004a**: The staging environment MUST serve as the active MVP environment, running all user-facing functionality with the cheapest viable configuration for each AWS service.
- **FR-004b**: The production environment MUST be fully defined in IaC with production-grade configuration (Multi-AZ RDS, increased redundancy) but remain unprovisioned until alpha release approval.
- **FR-005**: Each environment MUST be logically isolated using separate VPCs within a single AWS account, with security groups and network ACLs preventing cross-environment traffic.
- **FR-006**: Staging environment MUST use minimum viable resource configurations: smallest RDS instance type (db.t4g.micro single-AZ), minimal ECS Fargate task sizes (0.25 vCPU, 0.5 GB RAM), S3 Intelligent-Tiering storage class.
- **FR-007**: Production environment (when provisioned) MUST use production-grade configuration: Multi-AZ RDS for high availability, larger ECS task sizes for performance headroom, and CloudFront edge caching optimizations.
- **FR-008**: All IaC scripts MUST support provisioning either environment independently; production infrastructure can be created at any time without affecting staging.

#### Infrastructure as Code (IaC)

- **FR-009**: All infrastructure MUST be defined using Terraform (version 1.5 or higher) as the IaC tool.
- **FR-009a**: Infrastructure definitions MUST include both staging (active) and production (dormant) environments, with production resources tagged as "unprovisioned" until alpha release.
- **FR-010**: Terraform state MUST be stored remotely in AWS S3 with DynamoDB state locking to enable team collaboration and prevent concurrent modification conflicts.
- **FR-011**: Infrastructure changes MUST be applied through a version-controlled workflow (Git) with code review before execution.
- **FR-012**: IaC scripts MUST be modular and reusable across environments using Terraform workspaces, with environment-specific configuration provided via workspace-specific .tfvars files.
- **FR-013**: Infrastructure provisioning MUST be idempotent — repeated execution of the same configuration produces the same result without errors.

#### CI/CD Pipeline

- **FR-014**: The project MUST use GitHub Actions as the CI/CD platform, triggered by Git events (push, pull request, tag).
- **FR-014a**: GitHub Actions MUST authenticate to AWS using OIDC federation (OpenID Connect) by assuming an IAM role; no long-lived access keys may be stored in GitHub Secrets.
- **FR-015**: The CI pipeline MUST run on every pull request and include: linting, unit tests, integration tests, security scans, and IaC validation (terraform plan).
- **FR-016**: Deployment to staging MUST be automatic on merge to the main branch after successful CI checks.
- **FR-017**: Deployment to production MUST be manual-only (no automatic triggers) and require explicit approval for alpha release; production deployment capability must be tested in staging first.
- **FR-018**: Production deployments (when activated) MUST use the same immutable artifacts (Docker images) deployed to staging, with version tagging (semantic versioning).
- **FR-019**: The CD pipeline MUST deploy Docker images to ECS Fargate for backend services and static build artifacts to S3+CloudFront for frontend, without rebuilding across environments.
- **FR-020**: Failed deployments MUST trigger automatic rollback to the last known-good version and send alerts to the team.

#### Secrets & Configuration Management

- **FR-021**: All secrets (database passwords, API keys, tokens, certificates) MUST be stored in AWS Secrets Manager, never in code or environment variables committed to Git.
- **FR-022**: Application configuration MUST distinguish between secrets (sensitive) and non-secret environment-specific values (e.g., API URLs, feature flags).
- **FR-023**: Non-secret configuration MUST be managed via environment variables sourced from a configuration management system (e.g., AWS Parameter Store or environment-specific config files in the deployment pipeline).
- **FR-024**: Services MUST retrieve secrets at runtime from AWS Secrets Manager using IAM role-based authentication (no hardcoded credentials).
- **FR-025**: Secret rotation MUST be supported — applications must handle updated secrets without requiring redeployment (e.g., periodic refresh or signal-based reload).

#### Access Control & Permissions

- **FR-026**: Access to production infrastructure MUST be restricted to authorized personnel using AWS IAM roles with least-privilege principles.
- **FR-027**: Developers MUST have read-only access to staging and production environments for observability and debugging; write access requires approval.
- **FR-028**: IaC execution (terraform apply) for staging and production MUST be restricted to the CI/CD pipeline using the OIDC-federated IAM role; manual executions require audit logging.
- **FR-029**: AWS resources MUST use IAM roles and policies for service-to-service authentication; long-lived access keys and passwords must be avoided.

#### Cost Strategy

- **FR-030**: The staging environment (active MVP) MUST use the cheapest viable configuration for every AWS service: db.t4g.micro RDS instance (single-AZ), 0.25 vCPU / 0.5 GB RAM ECS Fargate tasks (ARM Graviton2), S3 Intelligent-Tiering, CloudFront with minimal edge locations.
- **FR-031**: Production environment (when provisioned) MUST use production-grade resources: db.t4g.small Multi-AZ RDS, 0.5 vCPU / 1 GB RAM ECS tasks, reserved capacity or savings plans for predictable long-running resources.
- **FR-032**: Cloud spending MUST be tracked with AWS Cost Explorer and tagged by environment (staging, production) and service (ecs, rds, s3, cloudfront) for transparency.
- **FR-033**: Budget alerts MUST be configured to notify the project owner when spending exceeds 80% and 100% of the monthly budget.
- **FR-034**: ECS Fargate auto-scaling MUST be configured with upper limits (max 5 tasks for staging, max 20 tasks for production when active) to prevent runaway costs from unexpected traffic spikes or misconfigurations.

#### Observability & Monitoring

- **FR-035**: All environments MUST emit structured logs (JSON format) to AWS CloudWatch Logs with retention policies (7 days for staging, 30 days for production when active).
- **FR-036**: Infrastructure health metrics (CPU, memory, disk, network) MUST be collected via AWS CloudWatch and displayed on a centralized dashboard.
- **FR-037**: Application performance metrics (request latency, error rate, throughput) MUST be collected and visualized.
- **FR-038**: Critical alerts (service downtime, resource exhaustion, deployment failures) MUST be sent to the operations team via email or chat integration (e.g., Slack).
- **FR-039**: Monitoring dashboards MUST be accessible to the entire team for transparency and troubleshooting.

### Key Entities *(infrastructure resources)*

- **VPC (Virtual Private Cloud)**: Network isolation boundary for an environment (staging, production); each VPC has its own CIDR block, subnets, route tables, and security groups within the shared AWS account.
- **Terraform Workspace/Module**: Represents an environment (staging, production) as a Terraform workspace with isolated state; each workspace uses the same .tf module definitions but different variable values from workspace-specific .tfvars files.
- **Application Load Balancer (ALB)**: Entry point for HTTP/HTTPS traffic, routes requests to ECS Fargate tasks running the backend service; supports path-based routing and health checks.
- **ECS Fargate Cluster**: Managed container orchestration service running backend services without managing EC2 instances; tasks execute Go backend code compiled for linux/arm64 (Graviton2 for cost optimization).
- **ECS Task Definition**: Configuration specifying Docker image, CPU/memory allocation (0.25 vCPU / 0.5 GB for staging, higher for production), environment variables, and IAM task role.
- **S3 Bucket**: Object storage for frontend static assets (React build artifacts) served via CloudFront, and application data (itinerary attachments, user uploads).
- **CloudFront Distribution**: CDN for serving frontend assets from S3 with HTTPS, caching, and global edge locations.
- **Database Instance**: Managed PostgreSQL database (Amazon RDS) with environment-specific configuration: staging uses db.t4g.micro single-AZ (cheapest), production uses db.t4g.small Multi-AZ (high availability when provisioned).
- **Secret**: Sensitive credential or key stored in AWS Secrets Manager, referenced by ARN at runtime by ECS tasks.
- **Configuration Parameter**: Non-secret environment-specific value stored in AWS Parameter Store or environment variables injected into ECS task definitions.
- **CI/CD Pipeline**: GitHub Actions workflow that builds, tests, and deploys code artifacts; authenticates to AWS via OIDC federation by assuming an environment-specific IAM role.
- **Deployment Artifact**: Immutable Docker image (backend) or static bundle (frontend) versioned and promoted from staging to production.
- **Cost Tag**: Metadata attached to AWS resources to enable cost tracking by environment and service.
- **Monitoring Dashboard**: CloudWatch dashboard displaying infrastructure and application metrics.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Infrastructure for staging can be provisioned from scratch in under 30 minutes using Terraform; production infrastructure definitions are validated in CI but remain unprovisioned until alpha release.
- **SC-002**: 100% of secrets and sensitive credentials are retrieved from AWS Secrets Manager by ECS tasks at runtime — zero hardcoded secrets in code or configuration files.
- **SC-003**: Deployments to staging are automatic and complete within 15 minutes of merging to main (includes Docker image build, push to ECR, ECS task update, health checks).
- **SC-004**: Infrastructure changes pass validation in CI (terraform plan) before being applied, with 100% of changes reviewed and approved via pull request.
- **SC-005**: Staging environment (active MVP) costs under $200/month with cheapest configuration (db.t4g.micro, 0.25 vCPU Fargate tasks, S3 Intelligent-Tiering); production (when provisioned) budgeted for $300-400/month with production-grade resources.
- **SC-006**: Monthly cloud spending is tracked and reported with environment-level granularity (staging vs production when active); budget alerts trigger at 80% of $200 staging threshold.
- **SC-007**: 100% of staging deployments use ECS rolling updates with health checks — unhealthy tasks are automatically replaced. Production (when activated) will use blue-green deployment for zero-downtime updates.
- **SC-008**: Failed ECS deployments trigger automatic rollback within 5 minutes by reverting to the previous task definition and restarting tasks.
- **SC-009**: Infrastructure health metrics are available in real-time on monitoring dashboards with 99.9% uptime for the monitoring system itself.
- **SC-010**: Critical infrastructure alerts (downtime, resource exhaustion) are delivered to the operations team within 2 minutes of detection.

## Assumptions

- The project team has access to a single AWS account with sufficient permissions to create IAM roles, VPCs, OIDC identity providers, ECS clusters, and other infrastructure resources for both staging and production.
- GitHub Actions is available and approved as the CI/CD platform; no corporate restrictions block its use or OIDC federation with AWS.
- Terraform Cloud or Terraform Enterprise is not required — open-source Terraform CLI with S3 remote state is sufficient for the MVP.
- The primary user base is in North America, making us-east-1 an appropriate default region; multi-region deployment is out of scope for MVP.
- The project team has at least one member with experience in AWS, Terraform, ECS/Docker, and CI/CD pipelines to lead infrastructure setup.
- Monthly cloud spending budget for MVP is under $200 USD for staging (active environment); production infrastructure (when provisioned for alpha) will add $200-300/month.
- The application backend is stateless and suitable for containerized deployment on ECS Fargate — all state persisted in RDS or S3, no long-running background processes requiring persistent compute outside of HTTP request handling.
- AI workload characteristics (conversational multi-turn Claude API calls, potentially large responses, < 10 second target latency from spec 002) require compute platform with no execution time limits, no payload size constraints, and persistent HTTP connection pooling to Anthropic API — this requirement rules out Lambda and necessitates ECS Fargate.
- GDPR compliance considerations (documented in spec 002) do not require multi-region data residency or EU-specific infrastructure at MVP; US-based infrastructure is acceptable with data protection policies.
- Production environment will remain unprovisioned (IaC-defined but not deployed) until MVP validation on staging confirms readiness for alpha release; this defers production costs until business value is demonstrated.
