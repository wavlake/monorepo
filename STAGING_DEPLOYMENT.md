# Staging Deployment Guide

This guide covers the staging deployment setup for the Wavlake API migration.

## Overview

The staging environment (`api-staging`) is deployed to Google Cloud Run using Cloud Build for automated building and deployment.

## Automatic Deployment

**Cloud Build triggers automatically deploy to staging** when:
- Code is pushed to `main` or `develop` branch  
- Changes are made to `apps/backend/` directory
- Native GCP integration with no additional authentication required

**Current Setup**: Cloud Build trigger `staging-auto-deploy` is active and configured via `task trigger:create`. See `CLOUD_BUILD_TRIGGERS.md` for detailed configuration and management.

**Benefits**:
- Native GCP integration with no external dependencies
- Direct repository monitoring and instant deployment
- Built-in build caching and optimized performance
- Integrated monitoring and logging through GCP Console

## Infrastructure Files

### Cloud Build Configuration
- **File**: `apps/backend/cloudbuild-staging.yaml`
- **Purpose**: Automated Docker build and Cloud Run deployment
- **Triggers**: Manual execution or GitHub integration

### Deployment Scripts
- **File**: `scripts/deploy-staging.sh` - Full automated deployment
- **File**: `scripts/deploy-staging-simple.sh` - Local testing variant
- **File**: `scripts/deploy-staging-final.sh` - Production-ready variant

### Integration Tests
- **File**: `apps/backend/tests/integration/staging_environment_test.go`
- **Purpose**: End-to-end testing of deployed staging environment

### Task Integration
- **Tasks**: `deploy:staging`, `deploy:staging:build`, `deploy:staging:test`
- **File**: `Taskfile.yml` (lines 423-448)

## Manual Deployment Process

### 1. Prerequisites

Ensure these GCP APIs are enabled in your project:
```bash
# Required APIs
- Cloud Build API
- Cloud Run API
- Container Registry API (or Artifact Registry API)
- Secret Manager API (for Firebase credentials)
```

### 2. Manual Cloud Build Deployment

#### Option A: Using gcloud CLI
```bash
cd /Users/joshremaley/dev/wavlake/monorepo

# Submit build to Cloud Build
gcloud builds submit \
  --config=apps/backend/cloudbuild-staging.yaml \
  --project=wavlake-alpha-alpha \
```

#### Option B: Using Google Cloud Console
1. Go to [Cloud Build Console](https://console.cloud.google.com/cloud-build)
2. Click "Create Trigger" or "Run Trigger"
3. Upload the `cloudbuild-staging.yaml` file
4. Set substitutions:
   - `PROJECT_ID`: `wavlake-alpha`
5. Run the build

### 3. Environment Variables

The staging deployment uses these environment variables:
```bash
DEVELOPMENT=false
GCP_PROJECT=wavlake-alpha  # or your project ID
ENVIRONMENT=staging
BACKEND_PORT=3000
GCS_BUCKET_NAME=wavlake-alpha-staging-storage  # Will be created if needed
```

### 4. Secrets Configuration (Optional)

For Firebase integration, create the secret:
```bash
# Create the secret
gcloud secrets create firebase-service-account-key --project=wavlake-alpha

# Add the service account key JSON
gcloud secrets versions add firebase-service-account-key \
  --data-file=path/to/your-service-account-key.json \
  --project=wavlake-alpha
```

## Testing the Deployment

### Using Task Commands
```bash
# Test the staging deployment
task deploy:staging:test

# Or with custom URL
STAGING_URL=https://your-service-url.run.app task deploy:staging:test
```

### Using Integration Tests
```bash
cd apps/backend

# Run staging-specific tests
STAGING_URL=https://api-staging-wavlake-alpha.run.app \
GCP_PROJECT=wavlake-alpha \
go test -v ./tests/integration -run TestStagingEnvironmentSuite
```

### Manual Testing
```bash
# Test heartbeat endpoint
curl https://api-staging-wavlake-alpha.run.app/heartbeat

# Test various endpoints
curl https://api-staging-wavlake.run.app/v1/heartbeat
curl https://api-staging-wavlake.run.app/dev/status
```

## Cloud Build Pipeline Details

The `cloudbuild-staging.yaml` performs these steps:

1. **Build Docker Image**: Creates optimized Go binary with Alpine Linux
2. **Push to Registry**: Uploads image to `gcr.io/PROJECT_ID/wavlake-api-staging`
3. **Deploy to Cloud Run**: Creates/updates the `api-staging` service
4. **Configure Service**: Sets memory, CPU, scaling, and environment variables

### Build Substitutions
- `$PROJECT_ID`: GCP project ID (automatically provided)
- `_SERVICE_NAME`: Cloud Run service name (default: `api-staging`)
- `_REGION`: GCP region (default: `us-central1`)

## Monitoring and Logs

### View Logs
```bash
# Cloud Run service logs
gcloud logs read \
  'resource.type=cloud_run_revision AND resource.labels.service_name=api-staging' \
  --limit=50 \
  --project=wavlake-alpha

# Cloud Build logs
gcloud builds list --project=wavlake
gcloud builds log BUILD_ID --project=wavlake
```

### Service Management
```bash
# List Cloud Run services
gcloud run services list --project=wavlake

# Describe the staging service
gcloud run services describe api-staging \
  --region=us-central1 \
  --project=wavlake-alpha

# Update service (if needed)
gcloud run services update api-staging \
  --region=us-central1 \
  --project=wavlake-alpha
```

## Integration with API Migration

This staging deployment is part of **Phase 4: Validation & Cutover** of the API migration plan:

- **Purpose**: Environment parity testing between staging and production
- **Validation**: Ensures Docker container works in cloud environment
- **Testing**: Validates API endpoints, authentication, and performance
- **Evidence**: Provides metrics for migration completion assessment

## Next Steps

1. ✅ **CI/CD Setup**: Cloud Build trigger `staging-auto-deploy` is active and working
2. **Production Deployment**: Use similar process for production environment
3. **Monitoring**: Set up Cloud Monitoring and alerting
4. **Load Testing**: Perform load testing against staging environment

## Troubleshooting

### Common Issues

1. **Permission Denied**: Ensure Cloud Build service account has necessary roles
2. **Image Push Failed**: Check Container Registry or Artifact Registry permissions
3. **Deployment Failed**: Review Cloud Run service configuration and logs
4. **Service Not Responding**: Check environment variables and Firebase setup

### Debug Commands
```bash
# Check Cloud Build service account permissions
gcloud projects get-iam-policy wavlake

# Test Docker image locally
docker run -p 8080:8080 \
  -e DEVELOPMENT=true \
  -e SKIP_AUTH=true \
  gcr.io/wavlake/wavlake-api-staging:latest

# Validate Cloud Build configuration
gcloud builds submit --config=apps/backend/cloudbuild-staging.yaml --dry-run
```