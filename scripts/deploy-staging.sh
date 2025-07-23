#!/bin/bash

# Staging Deployment Script using Google Cloud Build
# Usage: ./scripts/deploy-staging.sh [PROJECT_ID]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ID=${1:-wavlake-alpha}
REGION="us-central1"
SERVICE_NAME="api-staging"

# Use the full path for gcloud if available
GCLOUD_CMD="gcloud"
if [ -f "/Users/joshremaley/google-cloud-sdk/bin/gcloud" ]; then
    GCLOUD_CMD="/Users/joshremaley/google-cloud-sdk/bin/gcloud"
fi

echo -e "${BLUE}🚀 Staging Deployment via Cloud Build${NC}"
echo -e "${BLUE}====================================${NC}"
echo "Project ID: $PROJECT_ID"
echo "Service: $SERVICE_NAME"
echo "Region: $REGION"
echo ""

# Check if we're in the right directory
if [ ! -f "apps/backend/cloudbuild-staging.yaml" ]; then
    echo -e "${RED}❌ Error: Please run this script from the monorepo root directory${NC}"
    echo "Expected file: apps/backend/cloudbuild-staging.yaml"
    exit 1
fi

# Check if authenticated
if ! $GCLOUD_CMD auth list --filter=status:ACTIVE --format="value(account)" | grep -q .; then
    echo -e "${RED}❌ Error: Not authenticated with gcloud${NC}"
    echo "Please run: $GCLOUD_CMD auth login"
    exit 1
fi

# Set the project
echo -e "${YELLOW}📋 Setting GCP project...${NC}"
$GCLOUD_CMD config set project $PROJECT_ID

# Check Artifact Registry repository exists (use same repo as production)
echo -e "${YELLOW}🏗️ Checking Artifact Registry...${NC}"
if ! $GCLOUD_CMD artifacts repositories describe api-repo \
    --location=us-central1 \
    --project=$PROJECT_ID &>/dev/null; then
    echo -e "${RED}❌ Error: Artifact Registry repository 'api-repo' not found${NC}"
    echo "Please create it first:"
    echo "$GCLOUD_CMD artifacts repositories create api-repo --repository-format=docker --location=us-central1 --project=$PROJECT_ID"
    exit 1
else
    echo "✅ Artifact Registry repository 'api-repo' exists"
fi

# Submit build to Cloud Build
echo -e "${YELLOW}🏗️ Submitting build to Cloud Build...${NC}"
echo "This will build and deploy the Docker image automatically."
echo ""

BUILD_ID=$($GCLOUD_CMD builds submit \
    --config=apps/backend/cloudbuild-staging.yaml \
    --project=$PROJECT_ID \
    --format="value(id)")

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Build submitted successfully!${NC}"
    echo -e "${BLUE}Build ID: $BUILD_ID${NC}"
    echo ""
    
    # Get the service URL (might take a moment for deployment to complete)
    echo -e "${YELLOW}⏳ Waiting for deployment to complete...${NC}"
    sleep 30
    
    SERVICE_URL=$($GCLOUD_CMD run services describe $SERVICE_NAME \
        --region=$REGION \
        --project=$PROJECT_ID \
        --format="value(status.url)" 2>/dev/null || echo "")
    
    if [ -n "$SERVICE_URL" ]; then
        echo -e "${GREEN}🎉 Deployment completed successfully!${NC}"
        echo -e "${GREEN}Service URL: $SERVICE_URL${NC}"
        echo ""
        
        # Test the deployment
        echo -e "${YELLOW}🧪 Testing deployment...${NC}"
        if curl -f -s "$SERVICE_URL/heartbeat" > /dev/null; then
            echo -e "${GREEN}✅ Staging deployment is responding${NC}"
            RESPONSE=$(curl -s "$SERVICE_URL/heartbeat" 2>/dev/null || echo "No response")
            echo -e "${GREEN}Response: $RESPONSE${NC}"
        else
            echo -e "${RED}❌ Staging deployment is not responding yet${NC}"
            echo "Give it a few more minutes for the service to fully start up."
        fi
        
        echo ""
        echo -e "${BLUE}📊 Next steps:${NC}"
        echo "1. Test endpoints: curl $SERVICE_URL/heartbeat"
        echo "2. Run integration tests: STAGING_URL=$SERVICE_URL GCP_PROJECT=$PROJECT_ID task deploy:staging:test"
        echo "3. View logs: $GCLOUD_CMD logs read 'resource.type=cloud_run_revision AND resource.labels.service_name=$SERVICE_NAME' --limit=50 --project=$PROJECT_ID"
    else
        echo -e "${YELLOW}⚠️ Service URL not available yet. Check Cloud Console for deployment status.${NC}"
    fi
else
    echo -e "${RED}❌ Build submission failed${NC}"
    echo "Check the Cloud Build console for details: https://console.cloud.google.com/cloud-build/builds"
    exit 1
fi

echo ""
echo -e "${BLUE}🔗 Useful links:${NC}"
echo "Cloud Build: https://console.cloud.google.com/cloud-build/builds/$BUILD_ID?project=$PROJECT_ID"
echo "Cloud Run: https://console.cloud.google.com/run/detail/$REGION/$SERVICE_NAME?project=$PROJECT_ID"
echo ""
echo -e "${GREEN}🎯 Staging deployment complete! Use 'task deploy:staging:test' to run integration tests.${NC}"