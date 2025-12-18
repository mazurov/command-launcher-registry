# COLA Registry Complete Specification

**Version**: 1.3.0
**Created**: 2025-11-29
**Updated**: 2025-12-18
**Status**: Draft

## Overview

The COLA Registry is a complete solution for hosting and managing Command Launcher remote registries. It consists of two components:

1. **REST API Server** (`cola-registry`): HTTP server that manages registries, packages, and versions
2. **CLI Client** (`cola-regctl`): Command-line interface for managing registries via the REST API

This specification covers both components as an integrated system.

---

## Table of Contents

1. [Core Concepts](#core-concepts)
2. [User Scenarios](#user-scenarios)
3. [Functional Requirements](#functional-requirements)
   - [Server Requirements](#server-requirements)
   - [CLI Client Requirements](#cli-client-requirements)
4. [Non-Functional Requirements](#non-functional-requirements)
5. [Data Model](#data-model)
6. [API Contract](#api-contract)
7. [Configuration](#configuration)
8. [Edge Cases](#edge-cases)
9. [Success Criteria](#success-criteria)
10. [Assumptions](#assumptions)
11. [Command Launcher Compatibility](#command-launcher-compatibility)

---

## Core Concepts

### Key Entities

- **Registry**: A named container for packages. Has name, description, optional admin list, and optional custom_values. Each team can have their own registry.

- **Package**: Metadata for a command bundle within a registry. Has name, description, maintainer list, and custom_values. Contains multiple versions.

- **Package Version**: A specific release with version string, checksum, download URL, and partition range. Immutable once created.

- **Registry Index**: The generated `index.json` file served to Command Launcher clients. Contains array of all package versions in Command Launcher format.

### Permission Model (v1)

- Single authenticated user with full permissions to all operations
- Authentication can be enabled (`auth.type: basic`) or disabled (`auth.type: none`)
- When authentication is enabled, all write operations require valid credentials
- Registry `admins` field is optional and informational only (not enforced)
- Package `maintainers` field is optional and informational only (not enforced)

---

## User Scenarios

### Server Scenarios

#### Scenario 1: Serve Registry Index to Command Launcher Clients (Priority: P1)

As an enterprise platform team, I need to host a registry server that Command Launcher clients can query at `GET /api/v1/registry/:name/index.json` to discover available packages, so that all users can auto-update their commands.

**Success Criteria**:
- Server responds to index requests with HTTP 200 and valid JSON array
- Index contains all published package versions for the registry
- Command Launcher client successfully syncs packages from the registry
- Response time <100ms p50, <200ms p95 under normal load (100 req/min per IP)

**Acceptance Scenarios**:

1. **Given** a registry "build" exists with packages, **When** a client requests `GET /api/v1/registry/build/index.json`, **Then** it receives a JSON array of package versions with fields: name, version, checksum, url, startPartition, endPartition.

2. **Given** multiple registries exist (build, data, infra), **When** clients request different registry indexes, **Then** each returns only packages belonging to that registry.

#### Scenario 2: Manage Registries via Server API (Priority: P2)

As a Server Admin, I need to create, update, list, and delete registries so that different teams can have their own isolated package namespaces.

**Success Criteria**:
- All CRUD operations (create, read, update, delete) work via REST API
- Created registries appear in list endpoint immediately
- Updated registries reflect changes without server restart
- Deleted registries are completely removed including all packages

#### Scenario 3: Manage Packages via Server API (Priority: P3)

As a Registry Admin or Package Provider, I need to add and remove package metadata from a registry so that I can organize which commands are available.

**Success Criteria**:
- Packages can be created within a registry via REST API
- Created packages appear in registry's package list
- Deleted packages are completely removed including all versions
- Package metadata (name, description, maintainers) is preserved correctly

#### Scenario 4: Manage Package Versions via Server API (Priority: P4)

As a Registry Package Provider, I need to add and remove specific versions of packages so that users can receive updates through Command Launcher.

**Success Criteria**:
- Versions can be created with all required fields (version, checksum, url, partitions)
- Created versions appear immediately in registry index.json
- Version immutability is enforced (cannot update, only delete and recreate)
- Partition validation prevents overlapping ranges for the same package
- All checksums use SHA256 format with sha256: prefix

### CLI Client Scenarios

#### Scenario 5: Manage Registries via CLI (Priority: P1)

As a Server Admin, I need to create, update, list, and delete registries using CLI commands, so that I can manage registry infrastructure without writing API calls manually.

**Acceptance Scenarios**:

1. **Given** the server is running, **When** I run `cola-regctl registry create build --description "Build tools"`, **Then** a new registry is created and success message is displayed.

2. **Given** a registry exists, **When** I run `cola-regctl registry list`, **Then** all registries are displayed in human-readable format.

3. **Given** a registry exists, **When** I run `cola-regctl registry update build --description "Updated"`, **Then** the registry metadata is updated.

4. **Given** a registry exists, **When** I run `cola-regctl registry delete build`, **Then** the registry and all its packages are removed with confirmation message.

#### Scenario 6: Manage Packages via CLI (Priority: P2)

As a Package Provider, I need to add and remove packages from registries using CLI commands, so that I can organize which commands are available.

**Acceptance Scenarios**:

1. **Given** a registry "build" exists, **When** I run `cola-regctl package create build hotfix --description "Hotfix tool" --maintainer team`, **Then** the package is created.

2. **Given** a package exists, **When** I run `cola-regctl package list build`, **Then** all packages in that registry are displayed.

3. **Given** a package exists, **When** I run `cola-regctl package delete build hotfix`, **Then** the package and all its versions are removed.

#### Scenario 7: Manage Package Versions via CLI (Priority: P3)

As a Package Provider, I need to publish and remove specific versions of packages, so that users can receive updates through Command Launcher.

**Acceptance Scenarios**:

1. **Given** a package "hotfix" exists, **When** I run:
   ```
   cola-regctl version create build hotfix 1.0.0 \
     --checksum "sha256:abc123..." \
     --url "https://artifacts.example.com/hotfix-1.0.0.zip" \
     --start-partition 0 \
     --end-partition 9
   ```
   **Then** the version is published and appears in the registry index.

2. **Given** versions exist, **When** I run `cola-regctl version list build hotfix`, **Then** all versions for that package are displayed.

3. **Given** a version exists, **When** I run `cola-regctl version delete build hotfix 1.0.0`, **Then** that specific version is removed.

#### Scenario 8: Authentication Session Management (Priority: P1)

As a Registry Administrator, I need to authenticate once interactively and have my credentials reused for subsequent commands, so that I don't have to type credentials for every CLI operation.

**Acceptance Scenarios**:

1. **Given** a server with authentication enabled, **When** I run `cola-regctl login https://registry.example.com`, **Then** I am prompted for credentials interactively (hidden input), credentials are sent to server for validation, and if server accepts them (200 OK), credentials are saved securely.

2. **Given** I am logged in, **When** I run any registry command (e.g., `cola-regctl registry list`), **Then** the CLI sends saved credentials to server, and if server accepts them, command executes successfully without prompting.

3. **Given** I am logged in, **When** I run `cola-regctl logout`, **Then** saved credentials are removed and subsequent commands send no credentials.

4. **Given** I need to authenticate from a service/CI, **When** I set environment variables:
   ```bash
   export COLA_REGISTRY_URL=https://registry.example.com
   export COLA_REGISTRY_SESSION_TOKEN=<token-or-credentials>
   ```
   **Then** commands use these credentials without interactive login.

#### Scenario 9: Machine-Parseable Output (Priority: P4)

As a CI/CD pipeline, I need JSON output from all CLI commands, so that I can parse results programmatically.

**Acceptance Scenarios**:

1. **Given** any CLI command, **When** I add the `--json` flag, **Then** output is valid JSON with schema: `{"success": bool, "data": <result>, "error": <error>}`.

2. **Given** a command fails, **When** I use `--json` flag, **Then** error details are in JSON format with appropriate exit code.

#### Scenario 10: Version Compatibility Check (Priority: P5)

As a CLI user, I need the CLI to verify it's compatible with the server version, so that I don't encounter unexpected errors due to version mismatch.

**Acceptance Scenarios**:

1. **Given** CLI version 2.0.0 and server version 2.0.0, **When** I run any command, **Then** it executes successfully.

2. **Given** CLI version 2.0.0 and server version 1.0.0, **When** I run any command, **Then** it fails with clear error: "Version mismatch: CLI v2.0.0 incompatible with server v1.0.0".

3. **Given** CLI version 2.1.0 and server version 2.0.0, **When** I run any command, **Then** it shows warning but continues (minor version compatibility).

#### Scenario 11: Authentication Status Check (Priority: P6)

As a CLI user, I need to check my current authentication status and verify my credentials are valid, so that I can troubleshoot authentication issues and confirm which server I'm connected to.

**Acceptance Scenarios**:

1. **Given** I am logged in to a server, **When** I run `cola-regctl whoami`, **Then** it displays the server URL, authentication status, and my username.

2. **Given** my stored credentials have expired, **When** I run `cola-regctl whoami`, **Then** it reports authentication failure and suggests re-authenticating.

3. **Given** I am not logged in, **When** I run `cola-regctl whoami`, **Then** it displays error "Not logged in to any server" with exit code 5.

### OCI Storage Backend Scenarios

#### Scenario 12: Configure OCI Storage via URI (Priority: P1)

As an operator, I want to configure the cola-registry server to store its data in an OCI registry (like GitHub Container Registry) using a URI like `oci://ghcr.io/myorg/cola-registry-data`, so that I can leverage existing container registry infrastructure for durable, versioned storage without managing local file systems.

**Success Criteria**:
- Server starts successfully with OCI storage URI and token
- Server connects to OCI registry and can perform basic operations
- Clear error messages when configuration is invalid

**Acceptance Scenarios**:

1. **Given** a server with no configuration, **When** the operator provides `--storage-uri oci://ghcr.io/myorg/cola-registry-data` and `--storage-token <token>`, **Then** the server starts and connects to the OCI registry.

2. **Given** a server configured with OCI storage URI but no token, **When** the server starts, **Then** it fails with a clear error message indicating authentication is required for OCI storage.

3. **Given** a server configured with OCI storage, **When** the OCI registry is unreachable, **Then** the server fails to start with a clear error message indicating the connection failure.

4. **Given** a server configured with OCI storage pointing to a repository that doesn't exist yet, **When** the server starts, **Then** it creates an initial empty registry artifact in the OCI repository.

#### Scenario 13: Persist Registry Data to OCI (Priority: P1)

As an operator, I want every write operation (create/update/delete registry, package, or version) to persist the complete registry JSON to the OCI repository, so that my data is durably stored and can survive server restarts or migrations.

**Success Criteria**:
- All write operations result in a new artifact pushed to OCI registry
- Failed pushes cause operation rollback and error response
- Server can restart and resume with OCI-stored data

**Acceptance Scenarios**:

1. **Given** a server running with OCI storage, **When** a registry is created via REST API, **Then** the updated registry data is pushed to the OCI repository as a new artifact.

2. **Given** a server running with OCI storage, **When** a package version is created, **Then** the updated registry data is pushed to the OCI repository.

3. **Given** a server running with OCI storage, **When** a push to OCI fails (network error, auth expired), **Then** the operation fails, the in-memory state is rolled back, and a storage error is returned to the client.

4. **Given** a server running with OCI storage, **When** the server restarts, **Then** it loads the latest registry data from the OCI repository and resumes normal operation.

#### Scenario 14: Load Registry Data from OCI on Startup (Priority: P2)

As an operator, I want the server to load existing registry data from the OCI repository when it starts, so that I can restart or replace servers without losing data.

**Success Criteria**:
- New server instance loads existing data from OCI repository
- Empty repositories initialize with empty registry data
- Corrupted data causes clear error message

**Acceptance Scenarios**:

1. **Given** an OCI repository with existing registry data, **When** a new server instance starts with that OCI URI, **Then** the server loads the existing data and serves it correctly.

2. **Given** an OCI repository with no registry data (empty or non-existent), **When** the server starts, **Then** the server initializes with empty registry data and creates an initial artifact.

3. **Given** an OCI repository with corrupted or invalid JSON data, **When** the server starts, **Then** the server fails with a clear error message about data corruption.

#### Scenario 15: Support Multiple OCI Registries (Priority: P3)

As an operator, I want to use different OCI registries (GitHub Container Registry, Docker Hub, Azure Container Registry, AWS ECR, etc.) with the same URI scheme, so that I can choose the registry that fits my infrastructure.

**Success Criteria**:
- Server works with any OCI Distribution-compliant registry
- Authentication works with registry-specific tokens
- Same URI scheme for all registries

**Acceptance Scenarios**:

1. **Given** a server configured with `oci://ghcr.io/org/repo`, **When** the server performs storage operations, **Then** it uses GitHub Container Registry authentication and APIs.

2. **Given** a server configured with `oci://docker.io/user/repo`, **When** the server performs storage operations, **Then** it uses Docker Hub authentication and APIs.

3. **Given** a server configured with `oci://myregistry.azurecr.io/repo`, **When** the server performs storage operations, **Then** it uses Azure Container Registry authentication and APIs.

### S3 Storage Backend Scenarios

#### Scenario 16: Configure S3 Storage via URI (Priority: P1)

As an operator, I want to configure the cola-registry server to store its data in an S3-compatible storage (like AWS S3 or MinIO) using a URI like `s3://s3.us-east-1.amazonaws.com/mybucket/registry.json`, so that I can leverage existing S3 infrastructure for durable storage without managing local file systems.

**Success Criteria**:
- Server starts successfully with S3 storage URI and token
- Server connects to S3 and can perform basic operations
- Clear error messages when configuration is invalid

**Acceptance Scenarios**:

1. **Given** a server with no configuration, **When** the operator provides `--storage-uri s3://s3.us-east-1.amazonaws.com/mybucket/registry.json` and `--storage-token ACCESS_KEY:SECRET_KEY`, **Then** the server starts and connects to S3.

2. **Given** a server configured with S3 storage URI but no token and no AWS environment variables, **When** the server starts, **Then** it attempts IAM role authentication or fails with a clear error message.

3. **Given** a server configured with S3 storage, **When** the S3 endpoint is unreachable, **Then** the server fails to start with a clear error message indicating the connection failure.

4. **Given** a server configured with S3 storage pointing to an object that doesn't exist yet, **When** the server starts, **Then** it creates an initial empty registry JSON object.

#### Scenario 17: Persist Registry Data to S3 (Priority: P1)

As an operator, I want every write operation (create/update/delete registry, package, or version) to persist the complete registry JSON to S3, so that my data is durably stored and can survive server restarts or migrations.

**Success Criteria**:
- All write operations result in a new object uploaded to S3
- Failed uploads cause operation rollback and error response
- Server can restart and resume with S3-stored data

**Acceptance Scenarios**:

1. **Given** a server running with S3 storage, **When** a registry is created via REST API, **Then** the updated registry data is uploaded to S3.

2. **Given** a server running with S3 storage, **When** a package version is created, **Then** the updated registry data is uploaded to S3.

3. **Given** a server running with S3 storage, **When** an upload to S3 fails (network error, auth expired), **Then** the operation fails, the in-memory state is rolled back, and a storage error is returned to the client.

4. **Given** a server running with S3 storage, **When** the server restarts, **Then** it loads the latest registry data from S3 and resumes normal operation.

#### Scenario 18: Load Registry Data from S3 on Startup (Priority: P2)

As an operator, I want the server to load existing registry data from S3 when it starts, so that I can restart or replace servers without losing data.

**Success Criteria**:
- New server instance loads existing data from S3
- Empty/missing objects initialize with empty registry data
- Corrupted data causes clear error message

**Acceptance Scenarios**:

1. **Given** an S3 bucket with existing registry JSON data, **When** a new server instance starts with that S3 URI, **Then** the server loads the existing data and serves it correctly.

2. **Given** an S3 bucket with no registry data (object does not exist), **When** the server starts, **Then** the server initializes with empty registry data and creates an initial object.

3. **Given** an S3 bucket with corrupted or invalid JSON data, **When** the server starts, **Then** the server fails with a clear error message about data corruption.

#### Scenario 19: Support Multiple S3-Compatible Providers (Priority: P3)

As an operator, I want to use different S3-compatible storage providers (AWS S3, MinIO, DigitalOcean Spaces, Backblaze B2, etc.) with the same URI scheme, so that I can choose the provider that fits my infrastructure.

**Success Criteria**:
- Server works with any S3-compatible storage
- Authentication works with provider-specific credentials
- Same URI scheme for all providers

**Acceptance Scenarios**:

1. **Given** a server configured with `s3://s3.us-east-1.amazonaws.com/bucket/path`, **When** the server performs storage operations, **Then** it uses AWS S3 authentication and APIs.

2. **Given** a server configured with `s3+http://localhost:9000/bucket/path`, **When** the server performs storage operations, **Then** it uses MinIO with HTTP (non-TLS) connections.

3. **Given** a server configured with `s3://nyc3.digitaloceanspaces.com/bucket/path`, **When** the server performs storage operations, **Then** it uses DigitalOcean Spaces authentication and APIs.

---

## Functional Requirements

### Server Requirements

#### Registry Management

- **FR-S001**: Server MUST expose `GET /api/v1/registry/:name/index.json` returning Command Launcher compatible format (array of package versions with name, version, checksum, url, startPartition, endPartition).

- **FR-S002**: Server MUST expose `POST /api/v1/registry` to create a new registry with name, description, optional admin list, and optional custom_values.

- **FR-S003**: Server MUST expose `PUT /api/v1/registry/:name` to update registry metadata.

- **FR-S004**: Server MUST expose `DELETE /api/v1/registry/:name` to remove a registry and all its packages. This operation MUST be atomic.

- **FR-S005**: Server MUST expose `GET /api/v1/registry` to list all registries with their metadata.

- **FR-S006**: Server MUST expose `GET /api/v1/registry/:name` to retrieve detailed registry information.

#### Package Management

- **FR-S007**: Server MUST expose `POST /api/v1/registry/:name/package` to add package metadata.

- **FR-S008**: Server MUST expose `DELETE /api/v1/registry/:name/package/:package` to remove a package and all its versions. This operation MUST be atomic.

- **FR-S009**: Server MUST expose `GET /api/v1/registry/:name/package` to list all packages in a registry.

- **FR-S010**: Server MUST expose `GET /api/v1/registry/:name/package/:package` to retrieve detailed package information.

#### Version Management

- **FR-S011**: Server MUST expose `POST /api/v1/registry/:name/package/:package/version` to add a version.

- **FR-S012**: Server MUST expose `DELETE /api/v1/registry/:name/package/:package/version/:version` to remove a specific version.

- **FR-S013**: Server MUST expose `GET /api/v1/registry/:name/package/:package/version` to list all versions for a package.

- **FR-S014**: Server MUST expose `GET /api/v1/registry/:name/package/:package/version/:version` to retrieve detailed information for a specific version.

- **FR-S015**: System MUST validate partition fields (startPartition, endPartition) are integers in range 0-9 inclusive, and startPartition <= endPartition.

- **FR-S016**: System MUST reject partition ranges that would create ambiguous version selection (multiple versions with overlapping partition ranges for the same package).

#### Data Integrity

- **FR-S017**: System MUST enforce immutability - once a package version is published, it cannot be overwritten (only deleted and re-created).

- **FR-S018**: System MUST validate required fields on all create operations.

- **FR-S019**: System MUST enforce custom_values constraints on registry and package entities:
  - Maximum 20 key-value pairs per entity
  - Key format: 1-64 characters matching pattern `^[a-zA-Z_][a-zA-Z0-9_-]{0,63}$`
  - Value: Maximum 1024 UTF-8 characters

#### Server Configuration

- **FR-S020**: Server MUST support configuration via CLI flags and environment variables with prefix `COLA_REGISTRY_`, following 12-factor app principles.

- **FR-S021**: Server MUST accept `--port` CLI flag or `COLA_REGISTRY_SERVER_PORT` environment variable for port binding (default: 8080).

- **FR-S022**: Server MUST accept `--storage-uri` CLI flag or `COLA_REGISTRY_STORAGE_URI` environment variable for storage configuration using URI format (default: `file://./data/registry.json`).

- **FR-S022a**: Server MUST automatically prepend `file://` to storage paths that don't include a URI scheme (e.g., `./data/registry.json` becomes `file://./data/registry.json`).

- **FR-S022b**: Server MUST accept `--storage-token` CLI flag or `COLA_REGISTRY_STORAGE_TOKEN` environment variable for storage authentication (passed opaquely to storage backend).

- **FR-S023**: Server MUST accept `--auth-type` CLI flag or `COLA_REGISTRY_AUTH_TYPE` environment variable for authentication type (values: none, basic; default: none).

- **FR-S024**: Configuration precedence MUST be: CLI flags > environment variables > defaults (no config file support).

- **FR-S025**: Server MUST run without any configuration file (using only CLI flags, env vars, and defaults).

- **FR-S025a**: Server MUST provide all configuration via CLI flags: `--storage-uri`, `--storage-token`, `--port`, `--host`, `--log-level`, `--log-format`, `--auth-type`.

- **FR-S025b**: Server MUST log effective configuration at startup with masked sensitive values (storage token shown as `***`).

#### Server Operations

- **FR-S026**: Server MUST expose `GET /api/v1/health` endpoint returning server status and version.

- **FR-S027**: Health endpoint response MUST include: `{"status": "ok", "version": "<semver>"}`.

- **FR-S028**: Server MUST perform graceful shutdown on SIGTERM/SIGINT, ensuring all pending writes are flushed to storage before exit.

- **FR-S029**: All API endpoints MUST be prefixed with `/api/v1` for versioning.

- **FR-S030**: Server MUST return consistent error response format:
  ```json
  {
    "error": {
      "code": "ERROR_CODE",
      "message": "Human-readable error description",
      "details": {}
    }
  }
  ```

#### Storage & Recovery

- **FR-S031**: System MUST use file-based storage on disk for initial version, with pluggable backend interface to allow OCI registry or S3 storage later.

- **FR-S032**: If storage file does not exist on startup, server MUST create an empty registry file with valid JSON structure `{"registries": {}}`.

- **FR-S033**: If storage file is corrupted (invalid JSON), server MUST fail to start with clear error message.

- **FR-S034**: Storage implementation MUST use pessimistic locking (mutex/RWMutex) to prevent race conditions on concurrent write operations.

- **FR-S035**: All storage write operations MUST be atomic using write-to-temp-file-then-rename pattern.

#### Authentication & Authorization

- **FR-S036**: Server MUST support `auth.type: none` mode where all requests are allowed without authentication.

- **FR-S037**: Server MUST support `auth.type: basic` mode where write operations require HTTP Basic Auth.

- **FR-S038**: When authentication is enabled, all write operations (POST, PUT, DELETE) MUST require valid credentials.

- **FR-S039**: When authentication is enabled, read operations (GET) on index endpoints MUST NOT require authentication.

- **FR-S040**: In basic auth mode, server MUST read user credentials from `users.yaml` file with bcrypt-hashed passwords.

- **FR-S041**: Server MUST implement `GET /api/v1/whoami` endpoint for authentication testing:
  - Returns `{"username": "admin"}` on successful authentication (200 OK)
  - Returns 401 Unauthorized if credentials are invalid or missing

#### CLI Commands (Server Binary)

- **FR-S042**: Binary MUST expose `cola-registry server` subcommand to start the HTTP server.

- **FR-S043**: Binary MUST expose `cola-registry auth hash-password` utility command that generates bcrypt hash of a password.

#### Test Data Generation

- **FR-S044**: Project MUST include a shell script (`scripts/populate-test-data.sh`) that uses the REST API to create test registries, packages, and versions.

- **FR-S045**: Project MUST include a cleanup script (`scripts/clean-test-data.sh`) that removes all test data.

#### OCI Storage Backend

- **FR-S046**: System MUST support the `oci://` URI scheme for OCI registry storage configuration.

- **FR-S047**: System MUST parse OCI URIs to extract: registry host, repository path, and optional tag/digest.

- **FR-S048**: System MUST require a storage token when OCI storage is configured (fail startup if missing).

- **FR-S049**: System MUST authenticate to OCI registries using the provided storage token.

- **FR-S050**: System MUST push the complete registry JSON as an OCI artifact after every write operation, always overwriting the `latest` tag.

- **FR-S051**: System MUST pull the latest registry data from the OCI repository on server startup.

- **FR-S052**: System MUST create an initial empty registry artifact if the OCI repository is empty or does not exist.

- **FR-S053**: System MUST roll back in-memory state if OCI push fails.

- **FR-S054**: System MUST use `application/json` as the OCI artifact media type for the registry JSON data.

- **FR-S055**: System MUST support common OCI registries: ghcr.io, docker.io, and registries implementing OCI Distribution spec.

- **FR-S056**: System MUST reuse common storage logic between file and OCI backends (in-memory data model, CRUD operations, validation).

- **FR-S057**: System MUST use pessimistic locking (mutex) to ensure single-writer semantics for concurrent requests.

- **FR-S058**: System MUST log OCI operations (push, pull) with timing and size information.

- **FR-S059**: System MUST provide clear error messages distinguishing between authentication, network, and storage errors.

- **FR-S060**: System MUST reject OCI URIs with unsupported components (query parameters, fragments) with a clear validation error.

- **FR-S061**: System MUST enforce timeout limits: 60 seconds for push operations, 30 seconds for pull operations on startup.

- **FR-S062**: System MUST support OCI registries that implement manifest push/pull and blob upload (OCI Distribution Spec 1.0 minimum).

- **FR-S063**: System MUST gracefully handle registries that don't support artifact annotations (annotations are optional metadata).

- **FR-S064**: System MUST use the existing storage factory pattern to instantiate OCI storage when `oci://` scheme is detected.

#### S3 Storage Backend

- **FR-S065**: System MUST support the `s3://` and `s3+http://` URI schemes for S3 storage configuration.

- **FR-S066**: System MUST parse S3 URIs to extract: endpoint host, bucket name, object key path, and optional region query parameter.

- **FR-S067**: System MUST support token format `ACCESS_KEY:SECRET_KEY` for S3 authentication.

- **FR-S068**: System MUST fall back to `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables if no token is provided.

- **FR-S069**: System MUST support IAM role authentication when no explicit credentials are provided (for AWS deployments).

- **FR-S070**: System MUST upload the complete registry JSON to S3 after every write operation, overwriting the existing object.

- **FR-S071**: System MUST download the latest registry data from S3 on server startup.

- **FR-S072**: System MUST create an initial empty registry JSON object if the S3 object does not exist.

- **FR-S073**: System MUST roll back in-memory state if S3 upload fails.

- **FR-S074**: System MUST use `application/json` as the content type for the registry JSON object.

- **FR-S075**: System MUST support common S3-compatible providers: AWS S3, MinIO, DigitalOcean Spaces, Backblaze B2.

- **FR-S076**: System MUST reuse common storage logic between file, OCI, and S3 backends (in-memory data model, CRUD operations, validation).

- **FR-S077**: System MUST use pessimistic locking (mutex) to ensure single-writer semantics for concurrent requests.

- **FR-S078**: System MUST log S3 operations (upload, download) with timing and size information.

- **FR-S079**: System MUST provide clear error messages distinguishing between authentication, network, and storage errors.

- **FR-S080**: System MUST reject S3 URIs with fragments or unsupported query parameters with a clear validation error.

- **FR-S081**: System MUST enforce timeout limits: 60 seconds for upload operations, 30 seconds for download operations.

- **FR-S082**: System MUST validate that the S3 bucket exists and is accessible on startup.

- **FR-S083**: System MUST use `s3://` for HTTPS connections and `s3+http://` for HTTP connections (useful for local MinIO).

- **FR-S084**: System MUST auto-detect AWS region from endpoint URLs like `s3.REGION.amazonaws.com` or `s3-REGION.amazonaws.com`.

- **FR-S085**: System MUST use the existing storage factory pattern to instantiate S3 storage when `s3://` or `s3+http://` scheme is detected.

### CLI Client Requirements

#### Core CLI Operations

- **FR-C001**: CLI MUST provide commands for all registry operations:
  - `registry create <name>` - Create registry
  - `registry list` - List all registries
  - `registry get <name>` - Get registry details
  - `registry update <name>` - Update registry metadata
  - `registry delete <name>` - Delete registry

- **FR-C002**: CLI MUST provide commands for all package operations:
  - `package create <registry> <package>` - Create package
  - `package list <registry>` - List packages in registry
  - `package get <registry> <package>` - Get package details
  - `package update <registry> <package>` - Update package metadata
  - `package delete <registry> <package>` - Delete package

- **FR-C003**: CLI MUST provide commands for all version operations:
  - `version create <registry> <package> <version>` - Publish version
  - `version list <registry> <package>` - List versions of package
  - `version get <registry> <package> <version>` - Get version details
  - `version delete <registry> <package> <version>` - Delete version

- **FR-C004**: CLI MUST provide `whoami` command to show authentication status and server information.

#### Command Flags

- **FR-C005**: Registry create/update commands MUST support flags:
  - `--description <text>` - Optional description
  - `--admin <email>` - Admin email (repeatable for multiple admins)
  - `--custom-value <key=value>` - Custom metadata (repeatable)
  - `--clear-admins` - Remove all admins by sending empty array (update only)
  - `--clear-custom-values` - Remove all custom values by sending empty array (update only)

- **FR-C006**: Package create/update commands MUST support flags:
  - `--description <text>` - Optional description
  - `--maintainer <email>` - Maintainer email (repeatable)
  - `--custom-value <key=value>` - Custom metadata (repeatable)
  - `--clear-maintainers` - Remove all maintainers by sending empty array (update only)
  - `--clear-custom-values` - Remove all custom values by sending empty array (update only)

- **FR-C007**: Version create command MUST support flags:
  - `--checksum <sha256:hash>` - Required checksum in format "sha256:hash"
  - `--url <download-url>` - Required download URL
  - `--start-partition <0-9>` - Required partition start (0-9 inclusive, integer)
  - `--end-partition <0-9>` - Required partition end (0-9 inclusive, integer)
  - `--custom-value <key=value>` - Optional custom metadata (repeatable)

- **FR-C008**: Update commands MUST use **complete replacement** semantics for list fields.

#### Configuration & Connection

- **FR-C009**: CLI MUST support `--url` flag to specify server URL.

- **FR-C010**: CLI MUST read server URL from environment variable `COLA_REGISTRY_URL` if `--url` flag is not provided.

- **FR-C011**: CLI MUST remove trailing slash from server URLs before storing them.

- **FR-C012**: URL precedence order MUST be: `--url` flag > `COLA_REGISTRY_URL` env var > stored URL (from credentials file).

- **FR-C013**: CLI MUST store credentials in `~/.config/cola-registry/credentials.yaml` with `0600` permissions.

- **FR-C014**: CLI MUST support `--timeout` flag with default value of 30 seconds for all HTTP requests.

#### Authentication - Session Management

- **FR-C015**: CLI MUST provide `login [server-url]` command for interactive authentication.

- **FR-C016**: CLI `login` command MUST prompt for username and password interactively with hidden password input.

- **FR-C017**: CLI MUST store credentials for only one server at a time (most recent login).

- **FR-C018**: CLI MUST provide `logout` command that removes stored credentials.

- **FR-C019**: CLI MUST use stored credentials automatically for all operations after login.

#### Authentication - Command Line & Environment Variables

- **FR-C020**: CLI MUST support `--token` flag to provide authentication token for a single command.

- **FR-C021**: CLI MUST read authentication token from environment variable `COLA_REGISTRY_SESSION_TOKEN`.

- **FR-C022**: Authentication precedence order MUST be: `--token` flag > `COLA_REGISTRY_SESSION_TOKEN` env var > stored credentials.

- **FR-C023**: CLI MUST prevent credential exposure:
  - MUST NOT log credentials in any output
  - MUST NOT display credentials in error messages
  - MUST NOT include credentials in process listings
  - Interactive password input MUST use terminal no-echo mode

#### Output Formatting

- **FR-C024**: CLI MUST support `--json` flag for machine-parseable output on all commands.

- **FR-C025**: JSON output MUST follow schema:
  ```json
  {
    "success": true,
    "data": <command-specific-result>,
    "error": null
  }
  ```

- **FR-C026**: CLI MUST support `--verbose` flag for detailed operation logging.

- **FR-C027**: Default output (without `--json`) MUST be human-readable with clear formatting.

#### Exit Codes

- **FR-C028**: CLI MUST return specific exit codes:
  - 0: Success
  - 1: General error
  - 2: Invalid arguments/usage
  - 3: Resource not found (404)
  - 4: Conflict (409)
  - 5: Authentication error (401)
  - 6: Permission denied (403)

#### Version Compatibility

- **FR-C029**: CLI MUST send its version in `User-Agent` header: `cola-regctl/<version>`.

- **FR-C030**: CLI MUST call `GET /api/v1/health` on first operation to get server version.

- **FR-C031**: CLI MUST fail if server major version differs from CLI major version.

- **FR-C032**: CLI SHOULD warn (but continue) if server minor version differs from CLI minor version.

#### Shell Integration

- **FR-C033**: CLI MUST provide `completion` command to generate shell completions for bash, zsh, and fish.

#### Deletion Confirmation

- **FR-C034**: CLI MUST require interactive confirmation before executing any delete operation.

- **FR-C035**: Confirmation prompt MUST display what will be deleted and any cascade effects.

- **FR-C036**: CLI MUST support `--yes` or `-y` flag to skip confirmation for automation/scripting scenarios.

---

## Non-Functional Requirements

### Performance

- **NFR-001**: Server MUST respond to `GET /api/v1/registry/:name/index.json` requests in less than **100ms** under normal load (100 requests per minute per IP, 10 concurrent connections).

- **NFR-002**: CLI startup time (without server communication) MUST be under 100ms.

### Capacity

- **NFR-003**: System MUST support up to **100 registries**, **100 packages per registry**, and **100 versions per package** without degradation.

- **NFR-004**: Description fields (registry, package) MUST be limited to **4096 UTF-8 characters** maximum.

- **NFR-005**: Storage file (`registry.json`) MUST NOT exceed **100MB**.

### Security

- **NFR-006**: Server MUST implement rate limiting of **100 requests per minute per IP** for ALL endpoints.

- **NFR-007**: Server MUST log security-relevant events (authentication failures, authorization failures, deletions).

- **NFR-008**: Server MUST set `Access-Control-Allow-Origin: *` header for `GET` requests to index.json endpoint.

- **NFR-009**: Credential storage MUST meet platform-specific security requirements:
  - macOS: Use Keychain Services API
  - Windows: Use Credential Manager
  - Linux: File-based storage with 0600 permissions

### Observability

- **NFR-010**: Server MUST emit structured logs in JSON format to stdout/stderr.

- **NFR-011**: Server MUST expose basic HTTP metrics at `/api/v1/metrics` endpoint:
  - Request count by endpoint and status code
  - Request error rate (4xx, 5xx responses)
  - Request latency (p50, p95, p99) by endpoint

- **NFR-012**: Log entries MUST include: timestamp (ISO8601), level, message, request_id, endpoint, status_code, duration_ms.

### Usability

- **NFR-013**: All CLI commands MUST provide `--help` flag showing usage examples.

- **NFR-014**: Error messages MUST be clear and actionable.

- **NFR-015**: CLI MUST validate inputs locally before making API calls where possible.

- **NFR-016**: Long-running operations (>2 seconds) MUST show progress indication.

### Terminal Output

- **NFR-017**: CLI MUST support color and unicode output with auto-detection:
  - Use ANSI color codes: Green for success, Red for errors, Yellow for warnings
  - Use unicode symbols if terminal supports UTF-8
  - Fallback to ASCII if terminal doesn't support unicode
  - Support NO_COLOR environment variable
  - No colors or unicode in pipes, redirects, or non-TTY environments

---

## Data Model

### Registry

```json
{
  "name": "string (required, unique, 1-64 chars, alphanumeric/hyphen/underscore)",
  "description": "string (optional, max 4096 chars)",
  "admins": ["string (email/identifier)"],
  "custom_values": {"key": "value"}
}
```

### Package

```json
{
  "name": "string (required, unique within registry)",
  "description": "string (optional, max 4096 chars)",
  "maintainers": ["string (email/identifier)"],
  "custom_values": {"key": "value"}
}
```

### Package Version

```json
{
  "name": "string (package name)",
  "version": "string (semver)",
  "checksum": "string (sha256:<64-hex-chars>)",
  "url": "string (download URL, max 2048 chars)",
  "startPartition": "integer (0-9)",
  "endPartition": "integer (0-9)"
}
```

### CLI Credentials File

Location: `~/.config/cola-registry/credentials.yaml`

```yaml
url: "https://registry.example.com"
token: "user:password"  # Linux only; macOS/Windows store in OS keychain
```

### OCI Storage Entities

- **OCI Artifact**: A container image-like object stored in an OCI registry containing the registry JSON data as a blob with `application/json` media type. Always tagged as `latest` (overwritten on each push, no version history).

- **Registry Data**: The same JSON structure used by file storage (`{"registries": {...}}`) stored as the artifact's content.

- **Storage Token**: An authentication token (e.g., GitHub PAT, Docker Hub token) used to authenticate with the OCI registry. Token format is registry-specific.

---

## API Contract

### Endpoints Overview

#### Operational
- `GET /api/v1/health` - Health check and version info
- `GET /api/v1/metrics` - Server metrics
- `GET /api/v1/whoami` - Authentication status (requires auth)

#### Registries
- `GET /api/v1/registry` - List all registries
- `POST /api/v1/registry` - Create registry (auth required)
- `GET /api/v1/registry/:name` - Get registry details
- `PUT /api/v1/registry/:name` - Update registry (auth required)
- `DELETE /api/v1/registry/:name` - Delete registry (auth required)
- `GET /api/v1/registry/:name/index.json` - Get registry index (CDT format, public)

#### Packages
- `GET /api/v1/registry/:name/package` - List packages
- `POST /api/v1/registry/:name/package` - Create package (auth required)
- `GET /api/v1/registry/:name/package/:package` - Get package details
- `PUT /api/v1/registry/:name/package/:package` - Update package (auth required)
- `DELETE /api/v1/registry/:name/package/:package` - Delete package (auth required)

#### Versions
- `GET /api/v1/registry/:name/package/:package/version` - List versions
- `POST /api/v1/registry/:name/package/:package/version` - Create version (auth required)
- `GET /api/v1/registry/:name/package/:package/version/:version` - Get version details
- `DELETE /api/v1/registry/:name/package/:package/version/:version` - Delete version (auth required)

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| REGISTRY_NOT_FOUND | 404 | Registry does not exist |
| REGISTRY_ALREADY_EXISTS | 409 | Registry name already taken |
| PACKAGE_NOT_FOUND | 404 | Package does not exist |
| PACKAGE_ALREADY_EXISTS | 409 | Package name already exists in registry |
| VERSION_NOT_FOUND | 404 | Version does not exist |
| VERSION_ALREADY_EXISTS | 409 | Version already published (immutability) |
| VALIDATION_ERROR | 400 | Request validation failed |
| INVALID_PARTITION | 400 | Partition values out of range |
| PARTITION_OVERLAP | 400 | Partition overlap detected |
| STORAGE_UNAVAILABLE | 503 | Storage backend unavailable |
| OCI_AUTH_ERROR | 503 | OCI registry authentication failed |
| OCI_CONNECTION_ERROR | 503 | OCI registry connection failed |
| S3_AUTH_ERROR | 503 | S3 storage authentication failed |
| S3_CONNECTION_ERROR | 503 | S3 storage connection failed |
| S3_STORAGE_ERROR | 503 | S3 storage operation failed |
| UNAUTHORIZED | 401 | Authentication required/failed |

---

## Configuration

Configuration follows [12-factor app](https://12factor.net/) principles. No configuration file is required.

### Server CLI Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--storage-uri` | Storage URI (`file://` for local, `oci://` for OCI registry, `s3://` for S3) | `file://./data/registry.json` |
| `--storage-token` | Storage authentication token (required for OCI, optional for S3) | (empty) |
| `--port` | Server port | 8080 |
| `--host` | Bind address | 0.0.0.0 |
| `--log-level` | Log level (debug\|info\|warn\|error) | info |
| `--log-format` | Log format (json\|text) | json |
| `--auth-type` | Authentication type (none\|basic) | none |

### Server Environment Variables

All CLI flags have corresponding environment variables with `COLA_REGISTRY_` prefix:

| Variable | Description | Default |
|----------|-------------|---------|
| `COLA_REGISTRY_STORAGE_URI` | Storage URI | `file://./data/registry.json` |
| `COLA_REGISTRY_STORAGE_TOKEN` | Storage authentication token | (empty) |
| `COLA_REGISTRY_SERVER_PORT` | HTTP port | 8080 |
| `COLA_REGISTRY_SERVER_HOST` | Bind address | 0.0.0.0 |
| `COLA_REGISTRY_LOGGING_LEVEL` | Log level | info |
| `COLA_REGISTRY_LOGGING_FORMAT` | Log format | json |
| `COLA_REGISTRY_AUTH_TYPE` | Auth type (none/basic) | none |
| `COLA_REGISTRY_AUTH_USERS_FILE` | Users file path (env-only, no CLI flag) | ./users.yaml |

### Storage URI

The storage backend is configured via URI:

```bash
# File storage
--storage-uri file://./data/registry.json       # Relative path
--storage-uri file:///var/data/registry.json    # Absolute path (Unix)
--storage-uri ./data/registry.json              # Auto-prefixed with file://

# OCI storage (GitHub Container Registry)
--storage-uri oci://ghcr.io/myorg/cola-registry-data
--storage-token ghp_xxxxxxxxxxxxxxxxxxxx

# OCI storage (Docker Hub)
--storage-uri oci://docker.io/myuser/cola-registry-data
--storage-token dckr_pat_xxxxxxxxx

# OCI storage (Azure Container Registry)
--storage-uri oci://myregistry.azurecr.io/cola-registry-data
--storage-token <acr-token>
```

**OCI Storage Notes**:
- OCI storage requires `--storage-token` or `COLA_REGISTRY_STORAGE_TOKEN` environment variable
- The registry data is stored as an OCI artifact with `latest` tag (overwritten on each write)
- Supports any OCI Distribution-compliant registry
- Token format is registry-specific (e.g., GitHub PAT for ghcr.io)

```bash
# S3 storage (AWS S3)
--storage-uri s3://s3.us-east-1.amazonaws.com/mybucket/registry.json
--storage-token ACCESS_KEY:SECRET_KEY

# S3 storage with explicit region
--storage-uri s3://s3.amazonaws.com/mybucket/registry.json?region=us-east-1
--storage-token ACCESS_KEY:SECRET_KEY

# S3 storage (MinIO - local development)
--storage-uri s3+http://localhost:9000/mybucket/registry.json
--storage-token minioadmin:minioadmin

# S3 storage (DigitalOcean Spaces)
--storage-uri s3://nyc3.digitaloceanspaces.com/mybucket/registry.json
--storage-token ACCESS_KEY:SECRET_KEY

# S3 storage (Backblaze B2)
--storage-uri s3://s3.us-west-004.backblazeb2.com/mybucket/registry.json
--storage-token ACCESS_KEY:SECRET_KEY
```

**S3 Storage Notes**:
- S3 storage uses `s3://` for HTTPS or `s3+http://` for HTTP connections
- Token format: `ACCESS_KEY:SECRET_KEY`
- Falls back to `AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY` environment variables if no token provided
- Supports IAM role authentication (leave token empty)
- Region is auto-detected from AWS endpoints or can be specified via `?region=` query parameter
- Compatible with any S3-compatible storage: AWS S3, MinIO, DigitalOcean Spaces, Backblaze B2, Wasabi, etc.

### CLI Client Environment Variables

| Variable | Description |
|----------|-------------|
| `COLA_REGISTRY_URL` | Server base URL |
| `COLA_REGISTRY_SESSION_TOKEN` | Authentication token (format: user:password) |

### Precedence Order

1. **Server**: CLI flags > Environment variables > Defaults
2. **CLI (URL)**: `--url` flag > `COLA_REGISTRY_URL` env var > stored credentials
3. **CLI (Token)**: `--token` flag > `COLA_REGISTRY_SESSION_TOKEN` env var > stored credentials

### Server Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | Invalid configuration |
| 2 | Storage initialization failure |
| 3 | Server startup failure |

---

## Edge Cases

### Server Edge Cases

- **Duplicate registry name**: Return 409 Conflict with REGISTRY_ALREADY_EXISTS
- **Package in non-existent registry**: Return 404 Not Found with REGISTRY_NOT_FOUND
- **Duplicate version**: Return 409 Conflict with VERSION_ALREADY_EXISTS (immutability)
- **Storage unavailable**: Return 503 Service Unavailable with STORAGE_UNAVAILABLE
- **Malformed JSON request**: Return 400 Bad Request with VALIDATION_ERROR
- **Concurrent writes**: Pessimistic locking with last-write-wins semantics
- **Invalid partition range**: Return 400 Bad Request with INVALID_PARTITION

### OCI Storage Edge Cases

- **OCI auth required but no token**: Server fails to start with error: "OCI storage requires authentication token via --storage-token or COLA_REGISTRY_STORAGE_TOKEN"
- **OCI push network timeout**: Operation fails, in-memory state is rolled back, client receives 503 Storage Unavailable error
- **OCI pull on startup with invalid token**: Server fails to start with clear error indicating authentication failure
- **Concurrent OCI modifications**: Single-writer assumption: in-memory mutex ensures sequential writes; each write pushes a new artifact
- **OCI repository quota exceeded**: Push fails, operation is rolled back, client receives storage error with details from OCI registry
- **OCI artifact size exceeds limit**: Push fails, operation is rolled back, client receives error indicating size limit exceeded
- **OCI repository does not exist**: Server creates initial empty artifact on first startup
- **OCI artifact corrupted/invalid JSON**: Server fails to start with clear error about data corruption

### S3 Storage Edge Cases

- **S3 credentials invalid**: Server fails to start with clear error indicating authentication failure (categorized as auth error)
- **S3 bucket does not exist**: Server fails to start with error: "S3 bucket validation failed: bucket does not exist"
- **S3 upload network timeout**: Operation fails, in-memory state is rolled back, client receives 503 Storage Unavailable error
- **S3 download on startup with invalid credentials**: Server fails to start with clear error indicating authentication failure
- **Concurrent S3 modifications**: Single-writer assumption: in-memory mutex ensures sequential writes; each write uploads a new object
- **S3 object does not exist**: Server creates initial empty JSON object on first startup
- **S3 object corrupted/invalid JSON**: Server fails to start with clear error about data corruption
- **S3 URI with fragment**: Server fails validation with error: "S3 URI does not support fragments"
- **S3 URI with unknown query parameter**: Server fails validation with error: "S3 URI does not support query parameter: [param]"
- **S3 token format invalid**: Server fails to start with error indicating expected format "ACCESS_KEY:SECRET_KEY"

### CLI Edge Cases

- **Server unreachable**: Exit code 1 with error: "Failed to connect to server at <url>"
- **Authentication failure**: Exit code 5 with error: "Authentication failed (401). Please run 'cola-regctl login <server-url>' to re-authenticate."
- **No server configured**: Exit code 2 with error: "No server configured. Run 'cola-regctl login <server-url>' first."
- **Version mismatch**: Exit code 1 with clear error message
- **Invalid custom-value format**: Exit code 2 with error: "Invalid --custom-value format. Expected 'key=value', got: '<input>'"
- **Invalid checksum format**: Exit code 2 with specific format error
- **Invalid partition range**: Exit code 2 with error: "Invalid partition range: start cannot be greater than end"
- **Delete cancelled by user**: Exit code 0, no action taken

---

## Success Criteria

### Server Success Criteria

- **SC-S001**: Command Launcher clients can fetch registry index and use it for command synchronization.
- **SC-S002**: Teams can create and manage their own registries via REST API (self-service).
- **SC-S003**: All registry operations available via REST API.
- **SC-S004**: 100% of package versions include checksum for integrity verification.
- **SC-S005**: Server runs with zero configuration files (env vars only) for containerized deployments.
- **SC-S006**: Server exposes version in health endpoint for client compatibility checks.

### CLI Success Criteria

- **SC-C001**: All registry, package, and version operations available via CLI commands.
- **SC-C002**: CLI can operate against remote server without any configuration files (env vars only).
- **SC-C003**: JSON output mode enables full automation in CI/CD pipelines.
- **SC-C004**: Version compatibility checks prevent API mismatch errors.
- **SC-C005**: Shell completion works in bash, zsh, and fish for all commands and flags.

### OCI Storage Success Criteria

- **SC-O001**: Operators can start the server with OCI storage using only `--storage-uri oci://<registry>/<repo>` and `--storage-token` (no additional configuration required).
- **SC-O002**: All CRUD operations (registry, package, version) work identically with OCI storage as with file storage from the client's perspective.
- **SC-O003**: Server can be restarted and all previously stored data is available from the OCI repository.
- **SC-O004**: Push operations complete within 60 seconds for registry data up to 10MB under normal network conditions.
- **SC-O005**: Error messages clearly indicate whether failures are due to authentication, network, or storage issues.
- **SC-O006**: Storage backend code is structured to allow adding S3 support with minimal duplication.

### S3 Storage Success Criteria

- **SC-S3-001**: Operators can start the server with S3 storage using `--storage-uri s3://<endpoint>/<bucket>/<path>` and `--storage-token ACCESS_KEY:SECRET_KEY`.
- **SC-S3-002**: All CRUD operations (registry, package, version) work identically with S3 storage as with file storage from the client's perspective.
- **SC-S3-003**: Server can be restarted and all previously stored data is available from S3.
- **SC-S3-004**: Upload operations complete within 60 seconds for registry data up to 10MB under normal network conditions.
- **SC-S3-005**: Error messages clearly indicate whether failures are due to authentication, network, or storage issues.
- **SC-S3-006**: Server works with AWS S3, MinIO, DigitalOcean Spaces, and other S3-compatible providers without code changes.
- **SC-S3-007**: Operators can use `s3+http://` scheme for non-TLS connections to local MinIO instances.

---

## Assumptions

- Partition logic (which user gets which version) is handled by Command Launcher client, not this registry server.
- Package files (ZIP archives) are hosted externally; registry stores URLs only.
- HTTPS termination handled by reverse proxy in production.
- Initial version uses file-based storage on disk; OCI storage backend can be added later via URI scheme extension.
- CLI binary is named `cola-regctl` (distinct from server binary `cola-registry`).
- CLI and server are versioned together (same version number).
- Network connectivity exists between CLI and server.
- User has necessary OS permissions to access credential storage.
- Server configuration follows 12-factor app principles: no configuration file support, only CLI flags and environment variables.
- Storage paths without a URI scheme (e.g., `./data/registry.json`) are treated as file paths and automatically get `file://` prepended.
- Relative file paths are resolved relative to the current working directory where the server is started.

### OCI Storage Assumptions

- Only one server instance writes to the OCI registry at a time (single-writer assumption); multi-writer scenarios are out of scope.
- The OCI registry implements the OCI Distribution specification (most major registries do).
- Storage token format and authentication mechanism are registry-specific; the token is passed as-is to the OCI client library.
- Network connectivity between server and OCI registry is generally reliable; transient failures result in operation failures (no automatic retry).
- Registry data size will remain under 100MB (consistent with file storage limits).
- The OCI artifact uses a simple single-blob layout rather than multi-layer images.
- Pull operations on startup can tolerate higher latency (up to 30 seconds) compared to push operations during runtime.

### S3 Storage Assumptions

- Only one server instance writes to the S3 bucket at a time (single-writer assumption); multi-writer scenarios are out of scope.
- The S3 provider implements the S3 API (AWS S3, MinIO, DigitalOcean Spaces, Backblaze B2, etc.).
- Token format is `ACCESS_KEY:SECRET_KEY` for explicit credentials, or credentials can be provided via AWS environment variables or IAM roles.
- Network connectivity between server and S3 is generally reliable; transient failures result in operation failures (no automatic retry).
- Registry data size will remain under 100MB (consistent with file storage limits).
- The S3 bucket must already exist; the server does not create buckets automatically.
- Download operations on startup can tolerate higher latency (up to 30 seconds) compared to upload operations during runtime.

---

## Command Launcher Compatibility

This registry server is designed to be compatible with **Command Launcher (cdt) v1.x** remote registry protocol.

### Index Format Contract

The `GET /api/v1/registry/:name/index.json` endpoint returns a JSON array where each entry contains exactly these fields:

| Field | Type | Description |
|-------|------|-------------|
| name | string | Package name |
| version | string | Semantic version string |
| checksum | string | SHA256 checksum with `sha256:` prefix |
| url | string | Download URL for the package archive |
| startPartition | integer | Start of partition range (0-9) |
| endPartition | integer | End of partition range (0-9) |

### Client Configuration

Command Launcher clients configure this registry using:
```bash
cdt remote add <name> <server-url>/api/v1/registry/<registry-name>
```

### Compatibility Notes

- Index format matches the existing Command Launcher remote registry protocol
- Partition logic (user assignment to partitions 0-9) is handled client-side
- No additional fields are added to index entries to ensure backward compatibility
