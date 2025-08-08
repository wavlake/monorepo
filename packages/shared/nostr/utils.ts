// Nostr utility types and helpers

import type { NostrEvent, UnsignedNostrEvent, NostrFilter, EventValidationResult } from './events';

/**
 * Event creation helpers
 */
export interface EventTemplate {
  kind: number;
  tags?: string[][];
  content?: string;
  created_at?: number;
}

/**
 * Key pair for Nostr identity
 */
export interface NostrKeyPair {
  privateKey: string; // 32 bytes hex
  publicKey: string;  // 32 bytes hex
}

/**
 * Secp256k1 signature
 */
export interface NostrSignature {
  r: string; // 32 bytes hex
  s: string; // 32 bytes hex
}

/**
 * Event subscription options
 */
export interface SubscriptionOptions {
  closeOnEose?: boolean;
  timeout?: number; // milliseconds
  maxEvents?: number;
}

/**
 * Relay policy for read/write operations
 */
export interface RelayPolicy {
  read: boolean;
  write: boolean;
  priority: number; // Higher = more preferred
}

/**
 * Connection pool configuration
 */
export interface RelayPoolConfig {
  maxConnections?: number;
  reconnectDelay?: number;    // milliseconds
  connectionTimeout?: number; // milliseconds
  maxReconnectAttempts?: number;
  pingInterval?: number;      // milliseconds
}

/**
 * Event cache configuration
 */
export interface EventCacheConfig {
  maxSize: number;           // Maximum number of events to cache
  ttl: number;              // Time to live in milliseconds
  gcInterval: number;       // Garbage collection interval
}

/**
 * Filtering options for local event processing
 */
export interface LocalFilter extends NostrFilter {
  includeReplies?: boolean;
  includeReposts?: boolean;
  minReactions?: number;
  verified?: boolean;       // Only events with verified signatures
  fromFollowing?: boolean;  // Only events from followed users
}

/**
 * Event enrichment data
 */
export interface EventEnrichment {
  replyCount?: number;
  repostCount?: number;
  likeCount?: number;
  zapAmount?: number;       // Total zap amount in sats
  bookmarked?: boolean;
  muted?: boolean;
  reported?: boolean;
}

/**
 * User profile enrichment
 */
export interface ProfileEnrichment {
  followerCount?: number;
  followingCount?: number;
  noteCount?: number;
  lastSeen?: number;        // Unix timestamp
  nip05Verified?: boolean;
  lightningVerified?: boolean;
  trusted?: boolean;        // Local trust score
}

/**
 * Content warning types
 */
export enum ContentWarning {
  NSFW = 'nsfw',
  VIOLENCE = 'violence',
  LANGUAGE = 'language',
  POLITICAL = 'political',
  SPAM = 'spam',
  SCAM = 'scam'
}

/**
 * Event reporting reasons
 */
export enum ReportReason {
  SPAM = 'spam',
  HARASSMENT = 'harassment',
  COPYRIGHT = 'copyright',
  ILLEGAL = 'illegal',
  NSFW = 'nsfw',
  IMPERSONATION = 'impersonation',
  MALWARE = 'malware',
  OTHER = 'other'
}

/**
 * Trust score calculation
 */
export interface TrustMetrics {
  reputation: number;       // 0-100 reputation score
  verifications: number;    // Number of verifications
  mutualConnections: number; // Mutual followers/following
  activityScore: number;    // Recent activity level
  ageScore: number;         // Account age factor
}

/**
 * Performance metrics
 */
export interface PerformanceMetrics {
  connectionLatency: number;    // ms
  eventFetchTime: number;       // ms per event
  subscriptionLatency: number;  // ms to first event
  errorRate: number;           // 0-1 error rate
  uptime: number;              // 0-1 uptime percentage
}

/**
 * Analytics event types
 */
export enum AnalyticsEventType {
  EVENT_PUBLISHED = 'event_published',
  EVENT_RECEIVED = 'event_received',
  SUBSCRIPTION_CREATED = 'subscription_created',
  SUBSCRIPTION_CLOSED = 'subscription_closed',
  RELAY_CONNECTED = 'relay_connected',
  RELAY_DISCONNECTED = 'relay_disconnected',
  ERROR_OCCURRED = 'error_occurred',
  USER_ACTION = 'user_action'
}

/**
 * Analytics event data
 */
export interface AnalyticsEvent {
  type: AnalyticsEventType;
  timestamp: number;
  data: Record<string, any>;
  userId?: string;
  sessionId?: string;
}

/**
 * Queue configuration for outgoing events
 */
export interface EventQueueConfig {
  maxRetries: number;
  retryDelay: number;      // milliseconds
  batchSize: number;       // Events to send in batch
  flushInterval: number;   // milliseconds
}

/**
 * Rate limiting configuration
 */
export interface RateLimitConfig {
  maxEventsPerSecond: number;
  maxSubscriptionsPerMinute: number;
  burstAllowance: number;
  windowSize: number;      // milliseconds
}

/**
 * Moderation settings
 */
export interface ModerationSettings {
  autoMuteThreshold: number;     // Auto-mute after N reports
  contentWarnings: ContentWarning[];
  blockedWords: string[];
  trustedReporters: string[];    // Pubkeys of trusted reporters
  allowUnverified: boolean;      // Allow events from unverified pubkeys
}

/**
 * Privacy settings
 */
export interface PrivacySettings {
  hideMetadata: boolean;         // Hide metadata in kind 0 events
  privateFollowing: boolean;     // Hide following list
  relayDisclosure: boolean;      // Disclose relay usage
  trackingProtection: boolean;   // Prevent tracking
}

/**
 * Backup and recovery settings
 */
export interface BackupSettings {
  autoBackup: boolean;
  backupInterval: number;        // milliseconds
  maxBackups: number;
  encryptBackups: boolean;
  backupLocation: string;        // File path or URL
}

/**
 * Feature flags for experimental features
 */
export interface FeatureFlags {
  experimentalNips: number[];    // Experimental NIPs to enable
  betaFeatures: string[];        // Beta feature identifiers
  debugMode: boolean;
  analyticsEnabled: boolean;
  crashReporting: boolean;
}