# API Migration Plan: api/ → monorepo/apps/backend/

## Executive Summary

This document outlines the strategic migration of the Wavlake Go API from the standalone `api/` directory to the monorepo structure at `monorepo/apps/backend/`. The migration preserves all existing functionality while integrating with the monorepo's TDD-focused development workflow and type generation system.

## Current State Analysis

### API Structure (Source: `api/`)
- **Go Version**: 1.24.1 (latest)
- **Module**: `github.com/wavlake/api`
- **Key Dependencies**:
  - Firebase Admin SDK v4.16.1
  - Cloud Firestore & Storage
  - Gin web framework v1.10.1
  - Nostr protocol support (nbd-wtf/go-nostr)
  - PostgreSQL integration
  - Audio processing (FFmpeg)

### Architecture Components
- **Dual Authentication**: Firebase JWT + NIP-98 Nostr signatures
- **Storage**: Google Cloud Storage with presigned URLs
- **Database**: Firestore (primary) + PostgreSQL (legacy read-only)
- **Audio Processing**: FFmpeg-based compression pipeline
- **Deployment**: Docker containers to Cloud Run

### Monorepo Structure (Target: `monorepo/apps/backend/`)
- **Go Version**: 1.23.0 (needs update to match api/)
- **Module**: `github.com/wavlake/monorepo`
- **Testing Framework**: Ginkgo/Gomega (TDD-focused)
- **Type Generation**: Go structs → TypeScript interfaces
- **Build System**: Task runner (Taskfile.yml)

## Migration Strategy

### Phase 1: Foundation Setup (Days 1-2)
**Objective**: Establish monorepo backend structure without breaking existing API

#### 1.1 Module & Dependency Migration
```bash
# Update monorepo/apps/backend/go.mod
- Update Go version from 1.23.0 → 1.24.1
- Migrate all dependencies from api/go.mod
- Add missing dependencies:
  - gin-contrib/cors v1.7.2
  - lib/pq v1.10.9 (PostgreSQL)
  - nbd-wtf/go-nostr v0.51.12
  - google/uuid v1.6.0
  - stretchr/testify v1.10.0
```

#### 1.2 Directory Structure Setup
```
monorepo/apps/backend/
├── cmd/
│   ├── api/          # Renamed from server/
│   │   └── main.go   # Updated module paths
│   └── fileserver/   # Migrate from api/cmd/fileserver/
│       └── main.go
├── internal/
│   ├── auth/         # Complete auth system migration
│   ├── config/       # Service configuration
│   ├── handlers/     # HTTP handlers (expand existing)
│   ├── middleware/   # Logging & CORS middleware
│   ├── models/       # Expand existing user.go
│   ├── services/     # Business logic layer
│   └── utils/        # Audio & storage utilities
├── pkg/
│   └── nostr/        # Nostr protocol support
├── tests/
│   ├── integration/  # New integration tests
│   └── mocks/        # Generated mocks
└── tools/
    └── typegen/      # Existing tool (update for new types)
```

#### 1.3 Configuration Updates
- Update Dockerfile for new structure
- Migrate Makefile targets to Taskfile.yml
- Update Cloud Build configuration
- Migrate docker-compose.dev.yml

### Phase 2: Core Migration (Days 3-5)
**Objective**: Migrate core application code with full functionality

#### 2.1 Authentication System Migration
```bash
# Priority: High - Critical for all endpoints
internal/auth/
├── firebase.go              # Firebase JWT validation
├── firebase_link_guard.go   # Firebase UID linking
├── nip98.go                # NIP-98 signature validation
├── dual.go                 # Dual authentication middleware
└── flexible.go             # Flexible auth patterns
```

#### 2.2 Service Layer Migration
```bash
# Migrate with interface-first approach for testability
internal/services/
├── interfaces.go           # Service contracts
├── user_service.go         # User & pubkey linking
├── nostr_track.go         # Track lifecycle management
├── storage.go             # GCS integration
├── postgres_service.go    # Legacy database access
└── processing.go          # Audio processing pipeline
```

#### 2.3 HTTP Handlers Migration
```bash
# Expand existing handlers/ directory
internal/handlers/
├── auth.go               # Authentication endpoints
├── tracks.go            # Track CRUD operations
├── legacy_handler.go    # Legacy PostgreSQL endpoints
├── heartbeat.go         # Health check
└── responses.go         # Existing (expand)
```

#### 2.4 Models & Types Migration
```bash
# Critical: These generate TypeScript interfaces
internal/models/
├── user.go              # Existing (expand)
├── track.go             # Track metadata
├── auth.go              # Authentication types
└── legacy.go            # Legacy database types

internal/types/
├── requests.go          # API request types
├── responses.go         # API response types
└── errors.go           # Error types
```

### Phase 3: Integration & Testing (Days 6-7)
**Objective**: Ensure full compatibility and testing coverage

#### 3.1 Testing Infrastructure Setup
```bash
# Migrate to Ginkgo/Gomega framework
tests/
├── integration/
│   ├── auth_test.go         # Authentication flow tests
│   ├── tracks_test.go       # Track operations
│   └── legacy_test.go       # Legacy endpoint tests
├── mocks/
│   ├── user_service_mock.go # Existing (update)
│   ├── storage_mock.go      # Storage service mock
│   └── postgres_mock.go     # PostgreSQL service mock
└── setup/
    └── db_setup.go          # Existing (expand)
```

#### 3.2 Type Generation Integration
- Update `tools/typegen/main.go` for new model structures
- Generate TypeScript interfaces for all API types
- Ensure frontend type safety for all endpoints

#### 3.3 Docker & Deployment Migration
```bash
# Update monorepo deployment configuration
monorepo/
├── Dockerfile.backend      # New backend-specific Dockerfile
├── docker-compose.yml      # Update for monorepo structure
└── .github/
    └── workflows/
        └── deploy-backend.yml  # Backend deployment pipeline
```

### Phase 4: Validation & Cutover (Days 8-9)
**Objective**: Validate migration and prepare for cutover

#### 4.1 Environment Parity Testing
- Deploy monorepo backend to staging environment
- Run comprehensive integration tests
- Validate all authentication flows
- Test legacy endpoint compatibility
- Verify audio processing pipeline

#### 4.2 Performance & Load Testing
- Compare response times: api/ vs monorepo/apps/backend/
- Test concurrent authentication flows
- Validate file upload/processing performance
- Monitor memory usage and goroutine management

#### 4.3 Documentation Updates
- Update CLAUDE.md with monorepo-specific guidance
- Document new testing patterns
- Update deployment procedures
- Create migration verification checklist

### Phase 5: Production Deployment (Day 10)
**Objective**: Execute production cutover with rollback plan

#### 5.1 Pre-Deployment Checklist
- [ ] All tests passing (unit, integration, e2e)
- [ ] TypeScript interfaces generated and validated
- [ ] Docker images built and tested
- [ ] Environment variables configured
- [ ] Database migrations (if any) applied
- [ ] Monitoring and alerting configured

#### 5.2 Deployment Strategy
1. **Blue-Green Deployment**: Deploy monorepo backend alongside existing API
2. **Traffic Splitting**: Gradually shift traffic to new backend
3. **Monitoring**: Watch metrics, logs, and error rates
4. **Rollback Plan**: Quick revert to api/ if issues detected

#### 5.3 Post-Deployment Validation
- Verify all endpoints responding correctly
- Check authentication flows working
- Validate file upload/processing
- Monitor performance metrics
- Test legacy PostgreSQL integration

## Risk Assessment & Mitigation

### High-Risk Areas
1. **Authentication System**: Complex dual auth with Nostr signatures
   - **Mitigation**: Extensive integration testing, gradual rollout
2. **Legacy PostgreSQL Integration**: VPC connector and connection pooling
   - **Mitigation**: Test database connectivity, monitor connection limits
3. **Audio Processing Pipeline**: FFmpeg and file handling
   - **Mitigation**: Test with various audio formats, monitor processing timeouts
4. **Type Generation**: Frontend dependency on generated interfaces
   - **Mitigation**: Validate type generation before deployment

### Medium-Risk Areas
1. **Docker Configuration**: New build process and dependencies
   - **Mitigation**: Test docker builds in CI/CD pipeline
2. **Environment Variables**: Configuration drift
   - **Mitigation**: Document all environment variables, use validation

### Low-Risk Areas
1. **HTTP Routing**: Gin framework stays the same
2. **Business Logic**: No functional changes
3. **Cloud Storage**: Same GCS integration patterns

## Success Criteria

### Functional Requirements
- [ ] All existing API endpoints respond correctly
- [ ] Authentication flows work (Firebase + NIP-98)
- [ ] File upload and processing functional
- [ ] Legacy PostgreSQL endpoints working
- [ ] TypeScript interfaces generated correctly

### Performance Requirements
- [ ] Response times ≤ existing API performance
- [ ] Memory usage within acceptable limits
- [ ] Audio processing times unchanged
- [ ] Database connection pooling efficient

### Quality Requirements
- [ ] Test coverage ≥ 80% (unit + integration)
- [ ] All linting and formatting checks pass
- [ ] Documentation updated and accurate
- [ ] Deployment pipeline automated

## Rollback Plan

### Immediate Rollback (< 5 minutes)
1. Revert traffic routing to original `api/` deployment
2. Update DNS/load balancer configuration
3. Monitor for stability

### Database Rollback
- No database schema changes planned
- Firestore data compatible between versions
- PostgreSQL remains read-only

### Code Rollback
- Revert Git commit to pre-migration state
- Redeploy original Docker containers
- Restore original CI/CD pipeline

## Post-Migration Cleanup

### After 30 Days of Stable Operation
1. Archive original `api/` directory
2. Update documentation references
3. Clean up old Docker images and deployments
4. Remove staging environments
5. Update monitoring dashboards

## Timeline Summary

| Phase | Duration | Key Deliverables |
|-------|----------|------------------|
| 1. Foundation | 2 days | Module setup, directory structure |
| 2. Core Migration | 3 days | All application code migrated |
| 3. Integration | 2 days | Testing, type generation |
| 4. Validation | 2 days | Environment parity, performance |
| 5. Deployment | 1 day | Production cutover |

**Total Duration**: 10 days

**Critical Path**: Authentication system → Service layer → Type generation → Integration testing

## Resource Requirements

### Development Team
- 1 Backend Go developer (full-time)
- 1 DevOps engineer (50% allocation)
- 1 QA engineer (25% allocation)

### Infrastructure
- Staging environment for parallel testing
- Additional Cloud Run instances for blue-green deployment
- Monitoring and alerting setup

## Conclusion

This migration plan provides a comprehensive, low-risk approach to moving the Wavlake API into the monorepo structure. The phased approach ensures functionality is preserved while gaining the benefits of the monorepo's TDD workflow, type generation system, and unified build process.

The plan prioritizes critical components (authentication, core services) and includes robust testing and rollback strategies to ensure smooth production deployment.