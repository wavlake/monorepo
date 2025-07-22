#!/bin/bash

# Pre-commit hook for Wavlake monorepo
# This script runs tests, linting, and type checks before allowing commits

set -e

echo "🔍 Running pre-commit checks..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if we're in the correct directory
if [ ! -f "Taskfile.yml" ]; then
    print_error "Not in monorepo root directory"
    exit 1
fi

# Function to check if there are staged changes for a specific directory
has_staged_changes() {
    git diff --cached --name-only | grep -q "^$1/"
}

# Initialize status flags
FRONTEND_CHANGED=false
BACKEND_CHANGED=false
TYPES_CHANGED=false

# Check what has changed
if has_staged_changes "apps/frontend"; then
    FRONTEND_CHANGED=true
    print_status "Frontend changes detected"
fi

if has_staged_changes "apps/backend"; then
    BACKEND_CHANGED=true
    print_status "Backend changes detected"
fi

if has_staged_changes "packages/shared-types" || has_staged_changes "apps/backend/internal/models"; then
    TYPES_CHANGED=true
    print_status "Type definitions may have changed"
fi

# Regenerate types if needed
if [ "$TYPES_CHANGED" = true ] || [ "$BACKEND_CHANGED" = true ]; then
    echo "🔄 Regenerating TypeScript types..."
    if ! task types:generate; then
        print_error "Type generation failed"
        exit 1
    fi
    print_status "Types regenerated"
    
    # Stage any generated type files
    git add packages/shared-types/api/
fi

# Run linting and formatting
echo "🧹 Running linting and formatting..."

if [ "$FRONTEND_CHANGED" = true ]; then
    echo "📱 Checking frontend code..."
    
    # Run frontend linting
    if ! task lint:frontend; then
        print_error "Frontend linting failed"
        exit 1
    fi
    
    # Run frontend formatting (and stage changes)
    task format:frontend
    git add apps/frontend/
    
    print_status "Frontend linting passed"
fi

if [ "$BACKEND_CHANGED" = true ]; then
    echo "🖥️  Checking backend code..."
    
    # Run backend linting
    if ! task lint:backend; then
        print_error "Backend linting failed"
        exit 1
    fi
    
    # Run backend formatting (and stage changes)
    task format:backend
    git add apps/backend/
    
    print_status "Backend linting passed"
fi

# Run fast tests
echo "🧪 Running fast tests..."

if [ "$FRONTEND_CHANGED" = true ]; then
    echo "Testing frontend..."
    if ! task test:unit:frontend:fast; then
        print_error "Frontend tests failed"
        print_warning "Run 'task test:unit:frontend' for detailed output"
        exit 1
    fi
    print_status "Frontend tests passed"
fi

if [ "$BACKEND_CHANGED" = true ]; then
    echo "Testing backend..."
    if ! task test:unit:backend:fast; then
        print_error "Backend tests failed"
        print_warning "Run 'task test:unit:backend' for detailed output"
        exit 1
    fi
    print_status "Backend tests passed"
fi

# Build check
echo "🏗️  Verifying builds..."

if [ "$FRONTEND_CHANGED" = true ]; then
    echo "Building frontend..."
    if ! task build:frontend; then
        print_error "Frontend build failed"
        exit 1
    fi
    print_status "Frontend build successful"
fi

if [ "$BACKEND_CHANGED" = true ]; then
    echo "Building backend..."
    if ! task build:backend; then
        print_error "Backend build failed"
        exit 1
    fi
    print_status "Backend build successful"
fi

# Check for common issues
echo "🔍 Checking for common issues..."

# Check for TODO/FIXME/HACK comments in staged files
STAGED_FILES=$(git diff --cached --name-only)
if [ -n "$STAGED_FILES" ]; then
    TODO_COUNT=$(echo "$STAGED_FILES" | xargs grep -l "TODO\|FIXME\|HACK" 2>/dev/null | wc -l)
    if [ "$TODO_COUNT" -gt 0 ]; then
        print_warning "Found TODO/FIXME/HACK comments in staged files"
        echo "$STAGED_FILES" | xargs grep -n "TODO\|FIXME\|HACK" 2>/dev/null || true
    fi
fi

# Check for console.log in frontend files (excluding test files)
if [ "$FRONTEND_CHANGED" = true ]; then
    CONSOLE_LOGS=$(git diff --cached --name-only | grep "apps/frontend" | grep -E "\.(ts|tsx|js|jsx)$" | grep -v test | xargs grep -l "console\." 2>/dev/null || true)
    if [ -n "$CONSOLE_LOGS" ]; then
        print_warning "Found console.log statements in frontend files:"
        echo "$CONSOLE_LOGS"
    fi
fi

# Check for sensitive information
SENSITIVE_PATTERNS="password|secret|key|token|api_key"
SENSITIVE_FILES=$(echo "$STAGED_FILES" | xargs grep -il "$SENSITIVE_PATTERNS" 2>/dev/null | grep -v ".env.example" || true)
if [ -n "$SENSITIVE_FILES" ]; then
    print_error "Potential sensitive information found in:"
    echo "$SENSITIVE_FILES"
    print_warning "Please review these files before committing"
fi

# Final status
echo ""
print_status "All pre-commit checks passed! 🎉"
echo "📝 Commit includes:"
[ "$FRONTEND_CHANGED" = true ] && echo "  - Frontend changes"
[ "$BACKEND_CHANGED" = true ] && echo "  - Backend changes"  
[ "$TYPES_CHANGED" = true ] && echo "  - Type definition updates"
echo ""