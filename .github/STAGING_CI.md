# Staging CI/CD Documentation

## GitHub Actions for Staging Environment

### Automatic Deployment (`staging-deploy.yml`)

**Triggers:**
- Push to `main` or `develop` branch with changes in `apps/backend/`
- Pull requests to `main` with changes in `apps/backend/`

**Workflow:**
1. **Setup**: Checkout, Go setup, GCP authentication
2. **Pre-deploy**: Run unit tests
3. **Deploy**: Execute Cloud Build staging pipeline
4. **Validate**: Run integration tests against deployed service
5. **Notify**: Comment on PR with staging URL (for PRs)

**Required Secrets:**
- `GCP_SA_KEY`: Google Cloud service account JSON key with permissions:
  - Cloud Build Editor
  - Cloud Run Admin
  - Artifact Registry Writer
  - Service Account User

### Manual Testing (`manual-staging-test.yml`)

**Triggers:**
- Manual dispatch via GitHub Actions UI

**Options:**
- **Staging URL**: Override default staging URL
- **Test Suite**: Choose specific tests to run:
  - `all`: Run all staging tests
  - `staging-environment`: Environment validation tests
  - `staging-api`: API functionality tests
  - `health-check`: Basic health check only

## Setting up GCP Service Account

1. **Create Service Account:**
```bash
gcloud iam service-accounts create github-actions-staging \
  --description="GitHub Actions staging deployment" \
  --display-name="GitHub Actions Staging"
```

2. **Grant Required Roles:**
```bash
PROJECT_ID="wavlake-alpha"
SA_EMAIL="github-actions-staging@${PROJECT_ID}.iam.gserviceaccount.com"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/cloudbuild.builds.editor"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/artifactregistry.writer"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/iam.serviceAccountUser"
```

3. **Create and Download Key:**
```bash
gcloud iam service-accounts keys create github-actions-key.json \
  --iam-account="${SA_EMAIL}"
```

4. **Add to GitHub Secrets:**
   - Go to repository Settings → Secrets and variables → Actions
   - Add new secret `GCP_SA_KEY` with the contents of `github-actions-key.json`

## Workflow Outputs

### Successful Deployment
- ✅ Unit tests passed
- ✅ Cloud Build deployment succeeded
- ✅ Integration tests passed
- 💬 PR comment with staging URL (for PRs)

### Failed Deployment
- ❌ Clear error messages at each step
- 🔍 Links to Cloud Build logs
- 📋 Summary of what failed

## Monitoring

**View Workflow Status:**
- Repository → Actions tab
- Look for "Deploy to Staging" workflow runs

**View Staging Logs:**
```bash
gcloud logs read 'resource.type=cloud_run_revision AND resource.labels.service_name=api-staging' \
  --limit=50 --project=wavlake-alpha
```

**Manual Deployment:**
```bash
# If GitHub Actions fails, deploy manually
task deploy:staging
```

## Environment Variables Used

- `PROJECT_ID`: `wavlake-alpha`
- `SERVICE_NAME`: `api-staging`
- `REGION`: `us-central1`
- `STAGING_URL`: Auto-detected from deployed service
- `GCP_PROJECT`: Set to `PROJECT_ID` for tests

## Integration with Development Workflow

1. **Feature Development**: Work on backend changes in feature branch
2. **Pull Request**: Create PR → automatic staging deployment + tests
3. **Review**: Use staging URL from PR comment for manual testing
4. **Merge**: Merge to main → production-ready validation
5. **Production**: Deploy to production with confidence

## Troubleshooting

**Common Issues:**
- **Authentication**: Check `GCP_SA_KEY` secret is correctly set
- **Permissions**: Ensure service account has all required roles
- **API Enabled**: Verify Cloud Build, Cloud Run, Artifact Registry APIs are enabled
- **Timeout**: Increase timeout if builds are slow

**Debug Commands:**
```bash
# Check service account permissions
gcloud projects get-iam-policy wavlake-alpha

# View recent builds
gcloud builds list --project=wavlake-alpha --limit=5

# Test staging URL manually
curl https://api-staging-cgi4gylh7q-uc.a.run.app/heartbeat
```