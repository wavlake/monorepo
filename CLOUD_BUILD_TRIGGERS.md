# Cloud Build Triggers for Staging Deployment

This document explains how to set up Cloud Build triggers that automatically watch your GitHub repository and deploy to staging when backend changes are pushed.

## Overview

**Cloud Build Triggers** provide native GCP integration that eliminates the need for GitHub Actions to trigger deployments. The trigger watches your repository directly and initiates builds when specific conditions are met.

## Benefits vs GitHub Actions

| Aspect | Cloud Build Triggers | GitHub Actions |
|--------|---------------------|----------------|
| **Integration** | Native GCP, no auth setup | Requires service account JSON |
| **Monitoring** | Built-in GCP console | Separate GitHub Actions tab |
| **Billing** | GCP compute pricing | GitHub Actions minutes |
| **Latency** | Direct trigger | GitHub → GCP handoff |
| **Configuration** | GCP console/CLI | YAML workflows |

## Setup Methods

### Method 1: Connect Repository (Required First Step)

**Repository connection is a one-time setup through GCP Console:**

1. Go to [Cloud Build Triggers](https://console.cloud.google.com/cloud-build/triggers)
2. Click "Connect Repository" 
3. Choose "GitHub (Cloud Build GitHub App)"
4. Select `wavlake/monorepo`
5. Grant permissions

### Method 2: Create Trigger (After Repository Connected)

```bash
# Create the trigger automatically
task trigger:create
```

This creates a trigger with these settings:
- **Name**: `staging-auto-deploy`
- **Repository**: `wavlake/monorepo`
- **Branches**: `main`, `develop`
- **Path Filter**: `apps/backend/**`
- **Build Config**: `apps/backend/cloudbuild-staging.yaml`

### Method 3: Using GCP Console (Alternative)

1. Go to [Cloud Build Triggers](https://console.cloud.google.com/cloud-build/triggers)
2. Click "Create Trigger"
3. Connect your GitHub repository (if not already connected)
4. Configure the trigger:
   - **Name**: `staging-auto-deploy`
   - **Event**: Push to a branch
   - **Source**: `wavlake/monorepo`
   - **Branch**: `^(main|develop)$`
   - **Build Configuration**: Cloud Build configuration file
   - **Location**: `apps/backend/cloudbuild-staging.yaml`
   - **Included files filter**: `apps/backend/**`


## Trigger Configuration Details

### Path Filtering

The trigger only runs when files change in:
- `apps/backend/**` - Any backend code changes

It ignores changes in:
- `apps/frontend/**` - Frontend-only changes
- `docs/**`, `*.md` - Documentation changes  
- `apps/backend/tests/**` - Test-only changes (optional)

### Branch Patterns

- **main**: Production-ready code
- **develop**: Development branch testing

The regex `^(main|develop)$` ensures only these branches trigger builds.

### Build Process

When triggered, the build follows `apps/backend/cloudbuild-staging.yaml`:

1. **Docker Build**: Creates optimized Go binary with Alpine Linux
2. **Push to Artifact Registry**: Stores image at `us-central1-docker.pkg.dev/wavlake-alpha/api-repo/api-staging`
3. **Deploy to Cloud Run**: Updates the `api-staging` service
4. **Health Check**: Validates deployment success

## Managing Triggers

### List Triggers
```bash
task trigger:list
# or
gcloud builds triggers list --project=wavlake-alpha
```

### View Trigger Details  
```bash
gcloud builds triggers describe staging-auto-deploy --project=wavlake-alpha
```

### Delete Trigger
```bash
task trigger:delete
# or  
gcloud builds triggers delete staging-auto-deploy --project=wavlake-alpha
```

### Edit Trigger
```bash
# Edit via console or delete/recreate
gcloud builds triggers delete staging-auto-deploy --project=wavlake-alpha
task trigger:create
```

## Repository Connection

### First-Time Setup

If your GitHub repository isn't connected to Cloud Build:

1. Go to [Cloud Build Settings](https://console.cloud.google.com/cloud-build/settings)
2. Enable the Cloud Build API (if not enabled)
3. Connect your GitHub account:
   - Click "Connect Repository"
   - Authenticate with GitHub
   - Select `wavlake/monorepo`
   - Grant necessary permissions

### Required Permissions

The Cloud Build service account needs:
- **Source Repository Administrator** (to read repository)
- **Cloud Build Editor** (to create/run builds) 
- **Cloud Run Admin** (to deploy services)
- **Artifact Registry Writer** (to push images)

## Testing the Trigger

### Manual Test
```bash
# Make a small change to backend code
echo "// Trigger test" >> apps/backend/main.go

# Commit and push
git add apps/backend/main.go
git commit -m "test: trigger staging deployment"
git push origin main

# Monitor the build
gcloud builds list --project=wavlake-alpha --ongoing
```

### Verify Trigger Execution
```bash
# Check recent builds
gcloud builds list --project=wavlake-alpha --limit=5

# Watch build logs
gcloud builds log BUILD_ID --project=wavlake-alpha

# Test deployed service
curl https://api-staging-cgi4gylh7q-uc.a.run.app/heartbeat
```

## Monitoring and Troubleshooting

### Build Status
- **Console**: [Cloud Build History](https://console.cloud.google.com/cloud-build/builds)
- **CLI**: `gcloud builds list --project=wavlake-alpha`

### Common Issues

**Trigger Not Firing:**
- Check repository connection in Cloud Build settings
- Verify branch name matches pattern `^(main|develop)$`
- Ensure changes are in `apps/backend/` directory
- Check GitHub webhook delivery in repo settings

**Build Failures:**
- Review build logs: `gcloud builds log BUILD_ID`
- Check Docker build context and Dockerfile
- Verify service account permissions
- Ensure Artifact Registry repository exists

**Deployment Issues:**
- Check Cloud Run service configuration
- Verify environment variables in `cloudbuild-staging.yaml`
- Test image locally before deployment
- Review Cloud Run logs

### Debug Commands
```bash
# Check trigger configuration
gcloud builds triggers describe staging-auto-deploy --project=wavlake-alpha

# Test build configuration locally
gcloud builds submit --config=apps/backend/cloudbuild-staging.yaml --dry-run

# Check service account permissions
gcloud projects get-iam-policy wavlake-alpha

# View recent webhook deliveries (requires GitHub Admin)
# Go to: GitHub repo → Settings → Webhooks → Recent Deliveries
```

## Migration from GitHub Actions

Migration from GitHub Actions deployment is **complete**:

1. ✅ **Created the trigger**: `staging-auto-deploy` trigger is active
2. ✅ **Tested the trigger**: Confirmed working deployment on backend changes
3. ✅ **Removed GitHub Actions**: Deleted `.github/workflows/staging-deploy.yml`
4. ✅ **Updated documentation**: All docs now reference trigger-based approach

### Hybrid Approach (Optional)

You can keep both systems:
- **Cloud Build Trigger**: Automatic deployment on push
- **GitHub Actions**: Manual testing and PR validation

Just ensure they don't conflict by using different trigger conditions or manual-only GitHub Actions.

## Integration with Development Workflow

### Development Process

1. **Feature Development**: Create feature branch from `develop`
2. **Local Testing**: Use `task dev:tdd` for test-driven development
3. **Push to Feature Branch**: No automatic deployment (not main/develop)
4. **Create PR to develop**: Manual testing using `task deploy:staging`
5. **Merge to develop**: Automatic staging deployment via Cloud Build trigger
6. **Create PR to main**: Final review and testing
7. **Merge to main**: Production-ready staging deployment

### Team Coordination

- **Develop Branch**: Shared staging environment for team testing
- **Main Branch**: Production-ready staging for final validation
- **Feature Branches**: No automatic deployment (use manual deployment for testing)

This approach provides continuous staging updates while maintaining control over the deployment process.