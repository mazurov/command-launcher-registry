# COLA Registry Quickstart Guide

Get the COLA Registry server running in minutes with your preferred storage backend.

## Prerequisites

- Go 1.24+ (for building from source)
- Docker (for MinIO or containerized deployment)

## Build

```bash
git clone <repository-url>
cd command-launcher-registry
make build        # Build server
make build-cli    # Build CLI client
```

---

## Option 1: GitHub Container Registry (ghcr.io)

Store registry data in GitHub Container Registry using OCI storage.

### Step 1: Create a GitHub Personal Access Token

1. Go to GitHub Settings > Developer settings > Personal access tokens > Tokens (classic)
2. Click "Generate new token (classic)"
3. Select scopes:
   - `write:packages` - Upload packages
   - `read:packages` - Download packages
   - `delete:packages` - Delete packages (optional)
4. Copy the token (starts with `ghp_`)

### Step 2: Start the Server

```bash
# Set your GitHub username and token
export GITHUB_USER="your-username"
export GITHUB_TOKEN="ghp_xxxxxxxxxxxxxxxxxxxx"

# Start server with OCI storage
./bin/cola-registry server \
  --storage-uri oci://ghcr.io/${GITHUB_USER}/cola-registry-data \
  --storage-token ${GITHUB_TOKEN} \
  --port 8080
```

Or using environment variables:

```bash
export COLA_REGISTRY_STORAGE_URI=oci://ghcr.io/${GITHUB_USER}/cola-registry-data
export COLA_REGISTRY_STORAGE_TOKEN=${GITHUB_TOKEN}
./bin/cola-registry server
```

### Step 3: Verify

```bash
curl http://localhost:8080/api/v1/health
# {"status":"ok","version":"1.0.0"}
```

---

## Option 2: AWS S3

Store registry data in Amazon S3 - ideal for production deployments.

### Step 1: Create an S3 Bucket

1. Go to AWS Console > S3
2. Click "Create bucket"
3. Choose a bucket name (e.g., `my-cola-registry`)
4. Select your preferred region (e.g., `us-east-1`)
5. Keep default settings and create

### Step 2: Create IAM Credentials

1. Go to AWS Console > IAM > Users
2. Create a new user or select existing
3. Attach policy with S3 permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:HeadObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::my-cola-registry",
        "arn:aws:s3:::my-cola-registry/*"
      ]
    }
  ]
}
```

4. Create Access Key and save the credentials

### Step 3: Start the Server

```bash
# Set your AWS credentials
export AWS_ACCESS_KEY="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export AWS_REGION="us-east-1"
export S3_BUCKET="my-cola-registry"

# Start server with S3 storage
./bin/cola-registry server \
  --storage-uri "s3://s3.${AWS_REGION}.amazonaws.com/${S3_BUCKET}/registry.json" \
  --storage-token "${AWS_ACCESS_KEY}:${AWS_SECRET_KEY}" \
  --port 8080
```

Or using environment variables:

```bash
export COLA_REGISTRY_STORAGE_URI="s3://s3.us-east-1.amazonaws.com/my-cola-registry/registry.json"
export COLA_REGISTRY_STORAGE_TOKEN="AKIAIOSFODNN7EXAMPLE:wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
./bin/cola-registry server
```

Alternative: Use AWS environment variables (no `--storage-token` needed):

```bash
export AWS_ACCESS_KEY_ID="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export COLA_REGISTRY_STORAGE_URI="s3://s3.us-east-1.amazonaws.com/my-cola-registry/registry.json"
./bin/cola-registry server
```

### Step 4: Verify

```bash
curl http://localhost:8080/api/v1/health
# {"status":"ok","version":"1.0.0"}

# Check S3 for the registry file (using AWS CLI)
aws s3 ls s3://my-cola-registry/
# 2025-12-18 10:00:00        123 registry.json
```

### IAM Role Authentication (EC2/ECS/EKS)

For AWS deployments, you can use IAM roles instead of access keys:

```bash
# No token needed - uses instance/task role automatically
export COLA_REGISTRY_STORAGE_URI="s3://s3.us-east-1.amazonaws.com/my-cola-registry/registry.json"
./bin/cola-registry server
```

---

## Option 3: MinIO (Docker)

Store registry data in a local MinIO instance - perfect for development and testing.

### Step 1: Start MinIO

```bash
# Create data directory
mkdir -p ~/minio-data

# Start MinIO container
docker run -d \
  --name minio \
  -p 9000:9000 \
  -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  -v ~/minio-data:/data \
  minio/minio server /data --console-address ":9001"
```

### Step 2: Create a Bucket

Option A: Using MinIO Console (Web UI)
1. Open http://localhost:9001
2. Login with `minioadmin` / `minioadmin`
3. Click "Create Bucket"
4. Name it `cola-registry`

Option B: Using MinIO Client (mc)
```bash
# Install mc if needed
brew install minio/stable/mc  # macOS
# or download from https://min.io/docs/minio/linux/reference/minio-mc.html

# Configure mc
mc alias set local http://localhost:9000 minioadmin minioadmin

# Create bucket
mc mb local/cola-registry
```

### Step 3: Start the Server

```bash
# Start server with S3 storage (note: s3+http for non-TLS)
./bin/cola-registry server \
  --storage-uri s3+http://localhost:9000/cola-registry/registry.json \
  --storage-token minioadmin:minioadmin \
  --port 8080
```

Or using environment variables:

```bash
export COLA_REGISTRY_STORAGE_URI=s3+http://localhost:9000/cola-registry/registry.json
export COLA_REGISTRY_STORAGE_TOKEN=minioadmin:minioadmin
./bin/cola-registry server
```

### Step 4: Verify

```bash
curl http://localhost:8080/api/v1/health
# {"status":"ok","version":"1.0.0"}

# Check MinIO for the registry file
mc ls local/cola-registry/
# [2025-12-18 10:00:00 UTC]   123B registry.json
```

---

## Using the CLI Client

Once the server is running, use the CLI client to manage registries.

### Setup

```bash
# Set server URL and credentials
export COLA_REGISTRY_URL=http://localhost:8080
export COLA_REGISTRY_SESSION_TOKEN=admin:admin  # If auth enabled
```

### Create a Registry

```bash
./bin/cola-regctl registry create my-tools \
  --description "My team's tools"

# Output:
# Registry 'my-tools' created successfully
```

### Create a Package

```bash
./bin/cola-regctl package create my-tools deployment-cli \
  --description "Deployment automation tool" \
  --maintainer "team@example.com"
```

### Publish a Version

```bash
./bin/cola-regctl version create my-tools deployment-cli 1.0.0 \
  --checksum "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" \
  --url "https://releases.example.com/deployment-cli-1.0.0.tar.gz" \
  --start-partition 0 \
  --end-partition 9
```

### Verify the Registry Index

```bash
curl http://localhost:8080/api/v1/registry/my-tools/index.json | jq
```

Output:
```json
[
  {
    "name": "deployment-cli",
    "version": "1.0.0",
    "checksum": "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    "url": "https://releases.example.com/deployment-cli-1.0.0.tar.gz",
    "startPartition": 0,
    "endPartition": 9
  }
]
```

---

## Docker Compose (MinIO + COLA Registry)

For a complete local setup, use Docker Compose:

```yaml
# docker-compose.yml
version: '3.8'

services:
  minio:
    image: minio/minio
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    command: server /data --console-address ":9001"
    volumes:
      - minio-data:/data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3

  minio-init:
    image: minio/mc
    depends_on:
      minio:
        condition: service_healthy
    entrypoint: >
      /bin/sh -c "
      mc alias set local http://minio:9000 minioadmin minioadmin;
      mc mb local/cola-registry --ignore-existing;
      exit 0;
      "

  cola-registry:
    build:
      context: .
      dockerfile: docker/Dockerfile
    ports:
      - "8080:8080"
    environment:
      COLA_REGISTRY_STORAGE_URI: s3+http://minio:9000/cola-registry/registry.json
      COLA_REGISTRY_STORAGE_TOKEN: minioadmin:minioadmin
    depends_on:
      minio-init:
        condition: service_completed_successfully

volumes:
  minio-data:
```

Run:
```bash
docker-compose up -d
```

---

## Troubleshooting

### OCI Storage (ghcr.io)

**Error: "authentication required"**
- Verify your GitHub token has `write:packages` and `read:packages` scopes
- Check token hasn't expired

**Error: "repository not found"**
- The repository is created automatically on first push
- Ensure your token has permissions to create packages

### S3 Storage (AWS)

**Error: "Access Denied"**
- Verify IAM policy includes `s3:GetObject`, `s3:PutObject`, `s3:HeadObject`
- Check bucket name in policy matches your bucket
- Ensure credentials haven't expired

**Error: "bucket does not exist"**
- Verify bucket name is correct
- Check bucket exists in the correct region

**Error: "invalid access key"**
- Verify `AWS_ACCESS_KEY_ID` or token format `ACCESS_KEY:SECRET_KEY`
- Check for typos in credentials

**Using IAM Roles**
- On EC2/ECS/EKS, leave `--storage-token` empty to use instance role
- Ensure instance role has required S3 permissions

### S3 Storage (MinIO)

**Error: "bucket does not exist"**
- Create the bucket first using MinIO Console or `mc mb`

**Error: "Access Denied"**
- Verify credentials in `--storage-token` match MinIO root user
- Format must be `ACCESS_KEY:SECRET_KEY`

**Error: "connection refused"**
- Ensure MinIO is running: `docker ps | grep minio`
- Check port 9000 is accessible

**Using HTTPS with MinIO**
- For production MinIO with TLS, use `s3://` instead of `s3+http://`

---

## Next Steps

- [Full Configuration Reference](../README.md#configuration)
- [API Documentation](./spec.md#api-contract)
- [CLI Command Reference](../README.md#cli-client-cola-regctl)
