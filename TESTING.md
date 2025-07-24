# Wavlake Monorepo Testing Strategy

## 📊 Overall Testing Status

### Coverage Summary
| **Application** | **Target** | **Current** | **Status** | **Tests** |
|-----------------|------------|-------------|------------|-----------|
| **Backend** | 80%+ | **22.4%** | ❌ **CRITICAL** | 225 passing |
| **Frontend** | 75%+ | **~85%** | ✅ **GOOD** | 1 passing |
| **Integration** | 60%+ | **~70%** | ✅ **GOOD** | 7 suites |
| **E2E** | 90%+ | **TBD** | ⚠️ **PLANNED** | Playwright ready |

### Test Infrastructure Status
- **Framework**: ✅ Ginkgo + Gomega (Backend), Vitest + RTL (Frontend)
- **Mocking**: ✅ GoMock (Backend), MSW (Frontend)
- **E2E**: ✅ Playwright configured
- **Services**: ✅ Firebase emulators, Docker, Nostr relay
- **CI/CD**: ✅ Pre-commit hooks, automated coverage

---

## 🎯 Testing Philosophy & Standards

### Quality Targets by Application
- **Backend**: 80%+ coverage (business logic critical)
- **Frontend**: 75%+ coverage (component + integration)
- **Integration**: 60%+ coverage (critical user journeys)
- **E2E**: 90%+ coverage (key user workflows)

### Test Categories
1. **Unit Tests**: Isolated component/function testing
2. **Integration Tests**: Service interaction testing
3. **Contract Tests**: API interface validation
4. **E2E Tests**: Full user workflow testing

---

## 🚀 Development Workflow

### TDD Commands (Cross-App)
```bash
# Start TDD environment (all apps + test watchers)
task dev:tdd

# Red-Green-Refactor cycle
task red          # Create failing test
task green        # Run tests & implement  
task refactor     # Improve code while tests pass

# Fast feedback loops
task test:unit:fast    # Quick unit tests (no coverage)
task test:unit         # Full unit tests with coverage
```

### Testing by Application
```bash
# Backend testing
task test:unit:backend           # Go unit tests
task test:integration           # Integration tests
task coverage:backend          # Coverage report

# Frontend testing  
task test:unit:frontend        # React component tests
task test:e2e                  # Playwright E2E tests
task coverage:frontend         # Coverage report

# Full test suite
task test                      # All tests
task quality:check            # Comprehensive validation
```

---

## 🔧 Backend Testing (Current Focus)

### ⚠️ CRITICAL: Backend Coverage Gap Analysis
**Current**: 22.4% | **Target**: 80%+ | **Gap**: -57.6 percentage points

#### Services Layer Crisis (4.2% vs 85% target)
- `UserService`: Business logic core (minimal coverage)
- `NostrTrackService`: Track management (interface-only)
- `ProcessingService`: Audio pipeline (completely uncovered)

#### Immediate Actions Required
1. **Activate Mock Generation**:
   ```bash
   task mocks:generate
   cd apps/backend/internal/services && go generate ./...
   ```

2. **Fix Disabled Tests**:
   - User Service integration (Firebase emulator)
   - Audio pipeline tests (install ffmpeg)
   - Enhanced legacy handler (implement or remove)

3. **Create Missing Infrastructure**:
   ```bash
   mkdir -p apps/backend/tests/{mocks,testutil}
   ```

#### Backend Test Organization
```
apps/backend/
├── internal/
│   ├── handlers/*_test.go     # HTTP handler tests (48.1% coverage)
│   ├── services/*_test.go     # Business logic tests (4.2% coverage)
│   ├── auth/*_test.go         # Auth middleware tests (planned)
│   └── utils/*_test.go        # Utility function tests (partial)
├── tests/
│   ├── integration/           # API integration tests (7 suites)
│   ├── mocks/                 # Generated mocks (missing)
│   └── testutil/              # Test fixtures (missing)
```

### Backend Critical Timeline
- **Week 1**: Services layer emergency (4.2% → 40%)
- **Week 2**: Handler enhancement (48.1% → 65%)  
- **Week 3**: Infrastructure & auth (65% → 80%+)

---

## 🎨 Frontend Testing

### Current Status
- **Framework**: Vitest + React Testing Library + MSW
- **Coverage**: ~85% (single component test exists)
- **E2E**: Playwright configured but minimal tests

### Frontend Test Structure
```
apps/frontend/
├── src/
│   ├── components/**/*.test.tsx    # Component unit tests
│   ├── hooks/**/*.test.ts          # Custom hook tests
│   ├── pages/**/*.test.tsx         # Page component tests
│   ├── services/**/*.test.ts       # API service tests
│   └── test/
│       ├── setup.ts                # Test configuration
│       ├── mocks/                  # MSW mocks
│       └── utils/                  # Test utilities
```

### Frontend Testing Priorities
1. **Component Testing**: React Testing Library patterns
2. **Hook Testing**: Custom hooks with realistic scenarios
3. **API Integration**: MSW for API mocking
4. **User Workflows**: Critical user journeys

### Frontend Commands
```bash
# Development
task tdd:frontend              # Watch mode testing
task test:unit:frontend        # Run all component tests
task test:coverage:frontend    # Coverage report

# E2E Testing
task test:e2e                  # Playwright tests
```

---

## 🔗 Integration & E2E Testing

### Integration Test Infrastructure
- **Services**: Firebase emulators, PostgreSQL, Nostr relay
- **Environment**: Docker Compose managed
- **Scope**: API endpoints, authentication flows, data persistence

### E2E Test Strategy
- **Tool**: Playwright (multi-browser)
- **Focus**: Critical user journeys
- **Coverage**: 90%+ of key workflows

### Integration Commands
```bash
# Service management
task firebase:emulators        # Start Firebase emulators
task dev:services             # Start all development services

# Integration testing
task test:integration         # Backend integration tests
task test:contract           # API contract tests
```

---

## 📈 Monitoring & Quality Gates

### Daily Monitoring
```bash
# Coverage tracking
task coverage                 # Generate all coverage reports
task health                   # Service health checks
```

### Pre-Commit Quality Gates
- ✅ Unit tests pass
- ✅ Linting passes  
- ✅ Type checking passes
- ✅ Coverage thresholds met
- ✅ No sensitive data

### CI/CD Pipeline
- **Pre-commit**: Fast tests + linting
- **PR**: Full test suite + coverage reports
- **Deploy**: Integration tests + smoke tests

---

## 🎯 Success Metrics & Timeline

### Overall Monorepo Health
- **Backend**: 22.4% → 80%+ (3 weeks)
- **Frontend**: ~85% → 75%+ (maintain)
- **Integration**: ~70% → 60%+ (maintain)
- **E2E**: 0% → 90%+ (4 weeks)

### Weekly Targets
- **Week 1**: Backend services crisis (40% backend coverage)
- **Week 2**: Backend handlers + Frontend expansion (65% backend)
- **Week 3**: Backend infrastructure + E2E foundation (80% backend)
- **Week 4**: E2E completion + polish (90% E2E coverage)

### Quality Indicators
- **Test Execution**: <60s for full unit test suite
- **Coverage Trends**: +5 percentage points per week minimum
- **Failure Rate**: <1% flaky tests
- **Review Time**: Automated coverage reporting

---

## 📋 Application-Specific Details

### Backend Deep Dive
**Current Critical Issues**: See `apps/backend/TESTING.md` for comprehensive backend-specific analysis, including:
- Detailed coverage gaps by service
- Mock generation strategies  
- Disabled test fixes
- Service-by-service implementation plan

### Frontend Development
**Testing Patterns**:
- Component isolation with React Testing Library
- User event simulation
- MSW for API mocking
- Accessibility testing

### Infrastructure Dependencies
**Required Services**:
- Firebase emulators (Auth, Firestore, Storage)
- PostgreSQL test database
- Local Nostr relay (nak serve)
- FFmpeg for audio processing tests

---

**Status**: 🚨 **BACKEND CRITICAL** - Immediate action required on backend coverage gaps while maintaining strong frontend and integration test foundations.