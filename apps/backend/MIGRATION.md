# API Migration Documentation: Original vs Monorepo Backend

This document outlines the functionality gaps between the original `/dev/wavlake/api` backend and the current monorepo backend implementation.

## Executive Summary

The monorepo backend implements core functionality but is missing several advanced features from the original API. **Key missing areas include advanced audio processing, file management, webhook integration, and enhanced legacy endpoints.**

## Feature Comparison Matrix

| Feature Category | Original API | Monorepo Backend | Status |
|------------------|--------------|------------------|---------|
| **Basic Authentication** | ✅ Firebase + NIP-98 | ✅ Firebase + NIP-98 | ✅ Complete |
| **Core Track CRUD** | ✅ Full CRUD | ✅ Full CRUD | ✅ Complete |
| **Basic Legacy Endpoints** | ✅ 4 endpoints | ✅ 4 endpoints | ✅ Complete |
| **Advanced Audio Processing** | ✅ Multi-format | ❌ Single format | ❌ Missing |
| **Track Processing Webhooks** | ✅ Cloud Functions | ❌ None | ❌ Missing |
| **File Server** | ✅ Full featured | ❌ None | ❌ Missing |
| **Mock Storage** | ✅ Development | ❌ GCS only | ❌ Missing |
| **Enhanced Legacy** | ✅ Artist/Album tracks | ❌ Basic only | ❌ Missing |

## Missing Functionality Details

### 1. Advanced Audio Processing & Compression System

**Missing from Monorepo:**
```
POST /v1/tracks/:id/compress          # Request custom compression
PUT  /v1/tracks/:id/compression-visibility  # Manage version visibility  
GET  /v1/tracks/:id/public-versions   # Get public versions for Nostr
```

**Original Features:**
- Multiple compression formats (mp3, aac, ogg)
- Custom bitrate selection (128, 256, 320 kbps)
- Quality levels (low, medium, high)
- Per-version visibility controls (public/private)
- Metadata tracking for each compression

**Impact:** Users cannot customize audio quality or manage multiple versions

### 2. Track Processing & Status Management

**Missing from Monorepo:**
```
POST /v1/tracks/webhook/process       # Cloud Function webhook
GET  /v1/tracks/:id/status           # Processing status
POST /v1/tracks/:id/process          # Manual processing trigger
```

**Original Features:**
- Automated processing via GCS triggers
- Manual processing restart capability
- Real-time status monitoring
- Processing error reporting and recovery

**Impact:** No automated processing pipeline or status tracking

### 3. File Server & Storage Operations

**Missing Entire File Server:**
```
POST /upload                         # File upload
PUT  /upload                         # Alternative upload
GET  /file/*filepath                 # File download
GET  /status                         # Server status
GET  /list                           # File listing
DELETE /file/*filepath               # File deletion
```

**Original Features:**
- Token-based upload authentication
- Direct file download URLs
- File management operations
- Storage status monitoring

**Impact:** No direct file operations, limited development workflow

### 4. Enhanced Legacy PostgreSQL Integration

**Missing from Monorepo:**
```
GET /v1/legacy/artists/:artist_id/tracks    # Tracks by artist
GET /v1/legacy/albums/:album_id/tracks      # Tracks by album
```

**Original Features:**
- Artist-specific track listings
- Album-specific track listings
- Complex JOIN operations with filtering

**Impact:** Limited legacy data relationship queries

### 5. Development & Infrastructure Features

**Missing from Monorepo:**
```
GET  /dev/storage/list               # Mock storage listing
DELETE /dev/storage/clear            # Mock storage cleanup
```

**Original Features:**
- Complete mock storage system
- Local file server for development
- Offline development capabilities

**Impact:** Development requires GCS connectivity

## Architecture Differences

### Audio Processing Pipeline

**Original API:**
```
Upload → GCS Trigger → Cloud Function → FFmpeg Processing → 
Multiple Formats → Compression Versions → Visibility Management
```

**Monorepo Backend:**
```
Upload → Basic Processing → Single MP3 (128kbps) → Storage
```

### File Management

**Original API:**
- Dedicated file server with token authentication
- Mock storage for development
- Direct file operations API

**Monorepo Backend:**
- GCS-only storage
- No mock storage system
- Limited file management

### Legacy Data Access

**Original API:**
- Complex relationship queries
- Artist/album track relationships
- Advanced PostgreSQL operations

**Monorepo Backend:**
- Basic user-centric queries
- Limited relationship traversal

## Implementation Roadmap

### Phase 1: Critical Audio Processing (Weeks 1-3)
**Priority: HIGH** - Core functionality gap

1. **Advanced Compression System**
   - Implement multiple format support (mp3, aac, ogg)
   - Add bitrate and quality options
   - Create compression version management
   - Add visibility controls

2. **Processing Pipeline**
   - Implement webhook endpoint for Cloud Functions
   - Add manual processing triggers
   - Create status monitoring system
   - Add error handling and recovery

3. **FFmpeg Integration**
   - Add audio processing utilities
   - Implement format conversion
   - Add metadata extraction
   - Create quality validation

### Phase 2: File Server Integration (Weeks 4-5)
**Priority: HIGH** - Development workflow improvement

1. **Standalone File Server**
   - Implement cmd/fileserver binary
   - Add upload/download endpoints
   - Create token-based authentication
   - Add file management operations

2. **Mock Storage System**
   - Implement local storage interface
   - Add development configuration
   - Create GCS-compatible API
   - Add storage utilities

### Phase 3: Enhanced Legacy Support (Week 6)
**Priority: MEDIUM** - API completeness

1. **Additional Endpoints**
   - Add artist track listing endpoint
   - Add album track listing endpoint
   - Enhance PostgreSQL queries
   - Add relationship filtering

2. **Error Handling**
   - Improve database error handling
   - Add connection resilience
   - Create fallback mechanisms

### Phase 4: Infrastructure & DevOps (Week 7)
**Priority: LOW** - Operational excellence

1. **Cloud Function Integration**
   - Set up GCS triggers
   - Implement webhook processing
   - Add automated deployment
   - Create monitoring

2. **Development Tools**
   - Complete dev endpoint implementations
   - Add enhanced logging
   - Create debugging utilities

## Risk Assessment

### High Risk (Immediate Impact)
- **Audio Processing Gap**: Core feature missing affects user experience
- **No Processing Pipeline**: Manual operations limit scalability
- **Development Workflow**: GCS dependency complicates local development

### Medium Risk (Future Impact)
- **File Management**: Limited file operations affect admin workflows
- **Legacy Completeness**: Missing endpoints may affect migration
- **Automation Gap**: No Cloud Function integration limits processing efficiency

### Low Risk (Quality of Life)
- **Development Tools**: Missing dev endpoints slow debugging
- **Infrastructure**: Manual deployment processes

## Dependencies & Prerequisites

### External Dependencies
- **Cloud Functions**: For automated processing pipeline
- **FFmpeg**: For advanced audio processing
- **PostgreSQL**: For enhanced legacy endpoints

### Infrastructure Requirements
- **GCS Buckets**: For file storage and triggers
- **IAM Permissions**: For Cloud Function integration
- **Secrets Management**: For database connections

### Development Dependencies
- **Mock Storage**: For offline development
- **File Server**: For local testing
- **Token Generation**: For upload authentication

## Migration Strategy

### Backwards Compatibility
- All existing endpoints remain functional
- No breaking changes to current API
- Gradual feature addition approach

### Data Migration
- No data migration required
- Existing tracks compatible with new system
- Legacy data remains accessible

### Deployment Strategy
- Feature flags for new functionality
- Gradual rollout of advanced features
- Fallback to current implementation

## Success Metrics

### Phase 1 Success Criteria
- [ ] Multiple compression formats supported
- [ ] Webhook processing functional
- [ ] Processing status monitoring available
- [ ] Error handling comprehensive

### Phase 2 Success Criteria
- [ ] File server operational
- [ ] Mock storage working
- [ ] Development workflow improved
- [ ] Upload/download functional

### Phase 3 Success Criteria
- [ ] All legacy endpoints implemented
- [ ] Artist/album relationships working
- [ ] PostgreSQL operations enhanced

### Overall Success
- [ ] Feature parity with original API
- [ ] Performance maintained or improved
- [ ] Development workflow streamlined
- [ ] Production deployment automated

## Conclusion

The monorepo backend provides a solid foundation with core functionality implemented. The missing features primarily affect advanced use cases and development workflow. **Implementing Phase 1 (audio processing) should be the immediate priority** as it addresses the most significant functionality gap.

The migration strategy allows for gradual implementation while maintaining current functionality, reducing risk and allowing for iterative improvement.