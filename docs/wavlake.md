# Wavlake Platform - Technical Product Requirements Document

## Executive Summary

Wavlake is a Nostr-native music streaming platform that combines decentralized content discovery with robust monetization tools for artists. The platform uses Nostr for content metadata and social features, while leveraging traditional backend services for reliable content delivery and payment processing. Artists maintain sovereign control over their content and pricing, while listeners enjoy a seamless streaming experience powered by prepaid credits.

The system architecture consists of a React web client, Golang backend API, PostgreSQL database, streaming credits mint (wrapping cashu-mint-go), forked Nutstash wallet for credit management, and automated payment splitting service. Authentication is primarily through Nostr keypairs with Firebase as a backup recovery method (for artist accounts).

## System Architecture

### Monorepo Structure
```
wavlake-streaming-monorepo/
├── packages/
│   ├── mint/          # Streaming credits mint (cashu-mint-go wrapper)
│   └── shared/        # Shared types, utils, constants
├── apps/
│   ├── backend/           # Backend API written in go
│   ├── payments/          # Backend payment processor handling splits and other payment side effects
│   └── frontend/          # Main web application
└── docs/             # Documentation
```

### Core Components

#### 1. Web Client (React/Vite)
- Nostr event subscription and publishing via Wavlake relay
- Integration with Nostr browser extensions (NIP-07)
- Progressive Web App capabilities for mobile experience
- Real-time updates via WebSocket connections
- Integration with streaming wallet for credit management

#### 2. Golang Backend API (non-payments api)
- RESTful API serving clients
- Handles authentication (Nostr NIP-98 and Firebase)
- Manages presigned URLs for content upload to GCS
- Processes audio files and generates multiple quality tiers
- Maintains user and content metadata in PostgreSQL
- Integrates with streaming credits mint for credit operations
- Generates presigned URLs for secure content upload

#### 3. PostgreSQL Database
- Primary datastore for platform operations
- Stores user profiles and authentication mappings
- Maintains content metadata and relationships
- Tracks streaming credit quotes and delivery status
- Stores payment split configurations
- Provides analytics and reporting data

#### 4. Streaming Credits Mint Service (cashu-mint-go wrapper)
- Wraps cashu-mint-go library with streaming-specific restrictions
- Issues non-withdrawable credits with custom unit "streaming_credits"
- Disables NUT-05 (melting/withdrawals) completely
- Delivers credits via Nostr events (NIP-60 compatible)
- Integration with Lightning for direct purchases
- Gift card legal model for regulatory compliance

#### 5. Payment Splitting Service (payments api)
- Processes incoming Lightning payments
- Reads split metadata from database
- Calculates and executes payment distributions
- Handles failed payment retries
- Maintains settlement records

#### 6. Nostr Relay Infrastructure (completed, hosted at wss://relay.wavlake.com)
- Wavlake relay as primary event store
- Stores artist profiles, albums, and tracks as Nostr events

### Authentication Architecture

#### Primary Authentication: Nostr
- Users authenticate with Nostr keypair via NIP-07 browser extension or entering their nsec into the client
- Public key serves as primary user identifier
- NIP-98 HTTP authentication for API requests
- No central password database required

#### Backup Authentication: Firebase
- Optional Firebase account linking for recovery
- Enables account recovery if Nostr private key is lost
- Email/password, social login, or passwordless login options
- JWT tokens for session management

#### Account Linking Flow
1. User signs in with Nostr keypair
2. Option to link Firebase account appears in settings
3. User authenticates with Firebase via passwordless email
4. Backend creates mapping: `nostr_pubkey <-> firebase_uid`
5. User can now authenticate with either method, certain nostr actions will not be available in the app if there is no way to sign nostr events

## Data Models

### Nostr Events Architecture

All content metadata is stored as Nostr events, making the platform interoperable with other Nostr clients. Events use standard and custom kinds:

TODO - see nostr event architecture document

### Database Schema

#### Users Table
```sql
CREATE TABLE users (
  id UUID PRIMARY KEY,
  nostr_pubkey VARCHAR(64) UNIQUE NOT NULL,
  firebase_uid VARCHAR(128) UNIQUE,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  email VARCHAR(255),
  display_name VARCHAR(100)
);
```

#### Artists Table
```sql
CREATE TABLE artists (
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES users(id),
  nostr_event_id VARCHAR(64) UNIQUE,
  name VARCHAR(255) NOT NULL,
  bio TEXT,
  verified BOOLEAN DEFAULT false,
  created_at TIMESTAMP NOT NULL
);
```

#### Tracks Table
```sql
CREATE TABLE tracks (
  id UUID PRIMARY KEY,
  artist_id UUID REFERENCES artists(id),
  nostr_event_id VARCHAR(64) UNIQUE,
  title VARCHAR(255) NOT NULL,
  duration_seconds INTEGER,
  file_hash VARCHAR(64),
  price_per_stream INTEGER DEFAULT 10,
  play_count INTEGER DEFAULT 0,
  created_at TIMESTAMP NOT NULL
);
```

#### Payment Splits Table
```sql
CREATE TABLE payment_splits (
  id UUID PRIMARY KEY,
  track_id UUID REFERENCES tracks(id),
  recipient_pubkey VARCHAR(64) NOT NULL,
  percentage DECIMAL(5,2) NOT NULL,
  role VARCHAR(50),
  created_at TIMESTAMP NOT NULL,
  CONSTRAINT percentage_sum CHECK (
    SELECT SUM(percentage) FROM payment_splits 
    WHERE track_id = track_id
  ) = 100
);
```

#### Streaming Credits Tables (Mint-specific)
```sql
-- Quotes for credit purchases
CREATE TABLE quotes (
  id VARCHAR(64) PRIMARY KEY,
  pubkey VARCHAR(64) NOT NULL,
  amount_credits BIGINT NOT NULL,
  amount_sats BIGINT NOT NULL,
  payment_hash VARCHAR(64) UNIQUE NOT NULL,
  payment_request TEXT NOT NULL,
  paid BOOLEAN DEFAULT FALSE,
  delivered BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT NOW(),
  expires_at TIMESTAMP NOT NULL,
  
  INDEX idx_payment_hash (payment_hash),
  INDEX idx_pubkey_created (pubkey, created_at DESC)
);

-- Nostr delivery tracking
CREATE TABLE delivery_queue (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  quote_id VARCHAR(64) REFERENCES quotes(id),
  pubkey VARCHAR(64) NOT NULL,
  relay_urls TEXT[],
  event_id VARCHAR(64),
  delivered BOOLEAN DEFAULT FALSE,
  retry_count INTEGER DEFAULT 0,
  created_at TIMESTAMP DEFAULT NOW(),
  
  INDEX idx_pending_delivery (delivered, retry_count) WHERE delivered = FALSE
);

-- Daily statistics
CREATE TABLE daily_stats (
  date DATE NOT NULL,
  pubkey VARCHAR(64) NOT NULL,
  credits_purchased BIGINT DEFAULT 0,
  sats_spent BIGINT DEFAULT 0,
  purchase_count INTEGER DEFAULT 0,
  PRIMARY KEY (date, pubkey)
);
```

## Web Client MVP

### Core Features

#### User Authentication & Onboarding
1. **Nostr Login**
   - Detect NIP-07 compatible extensions
   - Request pubkey permission
   - Create/retrieve user profile
   - Store session in localStorage

2. **Firebase Backup Setup**
   - Optional flow during onboarding
   - Link existing account or create new
   - Verify email if provided
   - Store mapping in backend

3. **Profile Management**
   - Display Nostr pubkey
   - Edit display name and avatar
   - Manage linked accounts
   - View streaming credit balance

#### Music Discovery & Playback

1. **Browse Interface**
   - Homepage with featured content
   - Search by artist, album, track
   - Genre and mood categorization
   - Recently played history
   - Personalized recommendations

2. **Player Interface**
   - Persistent bottom player bar
   - Play/pause, skip, seek controls
   - Volume and quality settings
   - Queue management
   - Now playing information
   - Integrated credit deduction

3. **Streaming Credit Management**
   - Current balance display (from wallet)
   - Purchase credits via Lightning
   - Transaction history
   - Low balance warnings
   - Auto-top-up options
   - Gift card model explanation

#### Artist Tools

1. **Artist Profile Creation**
   - Claim artist name (unique check)
   - Upload profile and banner images
   - Add bio and links
   - Set Lightning address
   - Verification request

2. **Content Upload**
   - Create album with metadata
   - Batch upload tracks
   - Set cover art
   - Configure pricing per track
   - Preview before publishing

3. **Payment Configuration**
   - Add collaborators by pubkey
   - Set split percentages
   - Define roles (producer, featured, etc.)
   - Test split calculations
   - Save as templates

4. **Analytics Dashboard**
   - Play count metrics
   - Revenue reports
   - Listener geography
   - Top tracks performance
   - Payout history

### User Flows

#### New User Onboarding
```
1. Land on homepage → Browse as guest
2. Click "Sign In" → Detect Nostr extension
3. Approve pubkey access → Create user profile
4. Optional: Link Firebase → Email verification
5. Purchase credits → Lightning invoice (21 sats per 1000 credits)
6. Credits delivered via Nostr → Auto-detected by wallet
7. Start streaming → Credits deducted seamlessly
```

#### Credit Purchase Flow
```
1. Client requests invoice via /v1/credits/invoice with amount and pubkey
2. Mint creates Lightning invoice (21 sats per 1000 credits rate)
3. User pays invoice via WebLN or Lightning wallet
4. Mint monitors payment and marks quote as paid
5. Credits automatically delivered via Nostr (NIP-60 event)
6. Wallet detects new tokens and updates balance
7. No manual token claiming required
```

#### Artist Track Upload
```
1. Switch to Artist Mode → Verify profile
2. Create New Album → Enter metadata
3. Add Tracks → Get presigned URLs
4. Upload Files → Processing queue
5. Configure Splits → Add recipients
6. Set Pricing → Per-stream rate
7. Publish → Broadcast Nostr events
```

#### Listening Session
```
1. Search/Browse → Find content
2. Click Play → Check credit balance in wallet
3. Sufficient Credits → Begin streaming immediately
4. Deduct Credits → Update balance in background
5. Track Completion → Log analytics
6. Low Balance → Prompt top-up
```

## Backend Services

### API Endpoints (from main.go)

#### Authentication Endpoints
- `GET /v1/auth/get-linked-pubkeys` - Retrieve linked Nostr pubkeys for Firebase user
- `POST /v1/auth/link-pubkey` - Link Nostr pubkey to Firebase account
- `POST /v1/auth/unlink-pubkey` - Remove Nostr pubkey link
- `POST /v1/auth/check-pubkey-link` - Verify pubkey ownership via NIP-98

#### Track Management
- `GET /v1/tracks/:trackId` - Retrieve track metadata
- `POST /v1/tracks/nostr` - Create track from Nostr event
- `GET /v1/tracks/my` - List user's uploaded tracks
- `DELETE /v1/tracks/:trackId` - Remove track

#### Streaming Credits Endpoints
- `GET /v1/info` - Mint capabilities (shows melting disabled)
- `POST /v1/credits/invoice` - Create Lightning invoice for credits
- `GET /v1/credits/quote/{quote_id}` - Check payment status
- `POST /v1/mint` - Exchange paid quote for tokens (Cashu standard)
- `POST /v1/swap` - Token swap for change-making
- `POST /v1/checkstate` - Verify token spent status
- `POST /v1/melt` - Returns error (disabled for streaming credits)

#### Legacy Support
- `GET /v1/legacy/metadata` - User metadata for migration
- `GET /v1/legacy/tracks` - User track library
- `GET /v1/legacy/artists` - User artist follows
- `GET /v1/legacy/albums` - User album collection

#### Content Upload Flow
1. Client requests presigned URL from backend
2. Backend generates GCS presigned URL (15-minute expiry)
3. Client uploads directly to GCS
4. Backend receives completion webhook
5. Audio processing job triggered
6. Multiple quality tiers generated
7. Metadata updated in database
8. Nostr event published to relay

### Streaming Credits Mint Service

#### Credit Issuance
```
POST /v1/credits/invoice
{
  "amount": 10000,
  "pubkey": "02abc123...",
  "relays": ["wss://relay.wavlake.com", "wss://nos.lol"]
}
Response: {
  "quote_id": "550e8400-e29b-41d4-a716-446655440000",
  "payment_request": "lnbc210n1pjk8u9tpp5...",
  "payment_hash": "f7c42a24...",
  "amount_sats": 210,
  "amount_credits": 10000,
  "rate": "21 sats = 1000 credits",
  "expires_at": 1706234567
}
```

#### Credit Validation
```
POST /v1/checkstate
{
  "Ys": ["02abc...", "03def..."]
}
Response: {
  "states": [
    {"Y": "02abc...", "state": "UNSPENT"},
    {"Y": "03def...", "state": "SPENT"}
  ]
}
```

#### Integration Points
- Lightning invoice creation for purchases
- Webhook on payment confirmation
- Atomic credit operations via cashu-mint-go
- Balance queries for UI
- Expiration management (1 year)
- Non-withdrawable enforcement at mint level

### Payment Splitting Service

#### Split Processing Flow
1. **Payment Received**
   - Lightning payment to track address
   - Payment hash logged
   - Amount after platform fee calculated

2. **Split Retrieval**
   - Query payment_splits table
   - Validate percentages sum to 100%
   - Check recipient Lightning addresses

3. **Distribution Execution**
   ```
   For each recipient:
     amount = total * (percentage / 100)
     create_invoice(recipient_address, amount)
     send_payment(invoice)
     log_settlement(payment_id, recipient, amount)
   ```

4. **Error Handling**
   - Retry failed payments (3 attempts)
   - Queue for manual review if persistent failure
   - Notify artist of any issues
   - Hold funds in escrow until resolved

#### Settlement Rules
- Immediate settlement for amounts > 1000 sats
- Daily batch for smaller amounts
- Weekly minimum payout threshold
- Platform fee deducted upfront
- Transparent fee structure displayed

## Implementation Phases

### Phase 1: Core Platform (Months 1-2)

#### Authentication & User Management
- Implement NIP-98 authentication in Golang backend
- Firebase Auth integration and account linking
- User profile creation and management
- Basic PostgreSQL schema setup

#### Content Infrastructure
- GCS integration for file storage
- Presigned URL generation
- Basic audio file validation
- Single quality tier (256kbps)

#### Nostr Integration
- Connect to Wavlake relay
- Publish artist and track events
- Subscribe to content updates
- Basic event validation

#### Web Client Foundation
- Next.js application setup
- Nostr extension detection
- Basic browse and search
- Simple audio player

### Phase 2: Monetization (Months 3-4)

#### Streaming Credits Mint
- Fork and wrap cashu-mint-go library
- Implement streaming-specific restrictions
- Lightning invoice generation for credits
- Nostr delivery via NIP-60 events
- Non-withdrawable token enforcement

#### Streaming Wallet
- Fork Nutstash wallet
- Remove withdrawal features
- Add music player integration
- Implement kind:37560 wallet events
- Single-mint restriction

#### Payment Splitting
- Split configuration UI
- Payment processing service
- Settlement automation
- Basic reporting

#### Enhanced Player
- Credit balance display from wallet
- Seamless streaming deduction
- Purchase flow integration
- Transaction history

#### Artist Tools
- Upload interface
- Metadata editing
- Split management
- Revenue dashboard

### Phase 3: Scale & Optimize (Months 5-6)

#### Performance Optimization
- CDN integration for content delivery
- Multiple audio quality tiers
- Caching layer (Redis)
- Database query optimization

#### Advanced Features
- Playlist creation and sharing
- Social features (comments, likes)
- Recommendation algorithm
- Advanced analytics

#### Mobile Experience
- Progressive Web App enhancements
- Offline playback capability
- Push notifications
- Native app development planning

#### Platform Expansion
- Additional payment methods
- Multi-language support
- Third-party client SDK
- API rate limiting and quotas

## Technical Considerations

### Scalability
- Horizontal scaling of Golang API servers
- PostgreSQL read replicas for query distribution
- GCS multi-region replication
- Relay federation for Nostr events
- CDN edge caching for global performance

### Security
- NIP-98 signature verification for all API calls
- Rate limiting per pubkey
- Content hash verification
- SQL injection prevention
- XSS protection in web client
- Secure payment handling
- Non-withdrawable token enforcement at mint level
- Streaming credits cannot be melted back to sats
- Clear unit separation ("streaming_credits" not "sat")
- Gift card legal model for regulatory compliance

### Reliability
- 99.9% uptime SLA target
- Automated backup systems
- Disaster recovery plan
- Circuit breakers for external services
- Graceful degradation strategies

### Monitoring & Analytics
- Application performance monitoring (APM)
- Error tracking and alerting
- User behavior analytics
- Payment success rates
- Content delivery metrics
- Real-time dashboard

## Success Metrics

### Platform Health
- Monthly active users
- Total tracks uploaded
- Streaming credits in circulation
- Payment success rate
- Average session duration

### Artist Success
- Revenue per artist
- Average streams per track
- Split payment accuracy
- Upload to publish time
- Artist retention rate

### User Experience
- Time to first stream
- Credit purchase conversion
- Playback quality score
- Client load time
- Error rate

### Business Metrics
- Total payment volume
- Platform fee revenue
- Cost per stream
- Customer acquisition cost
- Lifetime value per user

## Risk Mitigation

### Technical Risks
- **Nostr relay availability**: Implement relay redundancy and local caching
- **Payment failures**: Automatic retries and manual intervention queue
- **Scalability bottlenecks**: Load testing and capacity planning
- **Data loss**: Regular backups and point-in-time recovery
- **Credit loss**: Multiple backup strategies via Nostr, recovery tools

### Business Risks
- **Credit fraud**: Implement velocity checks and anomaly detection
- **Content piracy**: Content fingerprinting and DMCA process
- **Regulatory compliance**: Gift card model for streaming credits
- **Platform abuse**: Community moderation and automated detection

### User Experience Risks
- **Complexity barrier**: Progressive disclosure and guided onboarding
- **Wallet friction**: WebLN integration and clear instructions
- **Credit confusion**: Clear pricing display and gift card metaphor
- **Technical issues**: Comprehensive error handling and support

## Future Roadmap

### Near-term (6-12 months)
- Native mobile applications
- Podcast support
- Live streaming capabilities
- Fan engagement tools
- Merchandise integration
- Artist-specific credit bundles
- Time-based passes

### Long-term (12-24 months)
- Decentralized content hosting
- Cross-platform identity
- Advanced monetization models
- AI-powered recommendations
- Global payment network integration
- Multi-mint federation for redundancy

## Conclusion

The Wavlake platform represents a fundamental shift in music streaming architecture, combining the openness of Nostr with the reliability of traditional backend services. By prioritizing artist sovereignty and user ownership while maintaining excellent user experience, we create a sustainable ecosystem for digital music consumption.

The implementation leverages proven technologies - wrapping cashu-mint-go for the mint and forking Nutstash for the wallet - to accelerate development while ensuring security and reliability. The streaming credits model, with its gift card semantics and non-withdrawable design, provides a familiar and compliant payment mechanism.

The phased implementation approach allows us to validate core assumptions early while building toward a comprehensive platform. With careful attention to scalability, security, and user experience, Wavlake can become the preferred platform for artists seeking fair compensation and listeners wanting to directly support creators.