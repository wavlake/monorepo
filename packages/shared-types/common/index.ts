// Common utility types used across the application

/**
 * Standard API response wrapper
 */
export interface ApiResponse<T = any> {
  data?: T;
  error?: string;
  message?: string;
  success: boolean;
  timestamp: string;
}

/**
 * Paginated response wrapper
 */
export interface PaginatedResponse<T = any> {
  data: T[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
    hasNext: boolean;
    hasPrev: boolean;
  };
  success: boolean;
  timestamp: string;
}

/**
 * Error response structure
 */
export interface ErrorResponse {
  error: string;
  code?: string;
  details?: Record<string, any>;
  success: false;
  timestamp: string;
}

/**
 * File upload response
 */
export interface FileUploadResponse {
  url: string;
  filename: string;
  size: number;
  mimetype: string;
  uploadedAt: string;
}

/**
 * Health check response
 */
export interface HealthCheckResponse {
  status: 'healthy' | 'degraded' | 'unhealthy';
  version: string;
  timestamp: string;
  services: {
    database: 'healthy' | 'unhealthy';
    storage: 'healthy' | 'unhealthy';
    relay: 'healthy' | 'unhealthy';
    auth: 'healthy' | 'unhealthy';
  };
  uptime: number; // seconds
}

/**
 * Generic ID types
 */
export type ID = string;
export type UUID = string;
export type Timestamp = number; // Unix timestamp
export type NostrPubkey = string; // 64-char hex
export type NostrEventId = string; // 64-char hex
export type LightningInvoice = string; // bolt11 invoice
export type LightningAddress = string; // user@domain.com

/**
 * Currency types
 */
export type Satoshis = number;
export type Millisatoshis = number;

/**
 * Audio/music related types
 */
export type Duration = number; // seconds
export type BPM = number;
export type AudioFormat = 'mp3' | 'wav' | 'flac' | 'ogg' | 'aac';
export type ImageFormat = 'jpg' | 'jpeg' | 'png' | 'webp' | 'svg';

/**
 * User roles and permissions
 */
export enum UserRole {
  USER = 'user',
  ARTIST = 'artist',
  MODERATOR = 'moderator',
  ADMIN = 'admin'
}

export enum Permission {
  READ = 'read',
  WRITE = 'write',
  DELETE = 'delete',
  MODERATE = 'moderate',
  ADMIN = 'admin'
}

/**
 * Content status types
 */
export enum ContentStatus {
  DRAFT = 'draft',
  PENDING = 'pending',
  PUBLISHED = 'published',
  ARCHIVED = 'archived',
  DELETED = 'deleted',
  BANNED = 'banned'
}

/**
 * Payment status types
 */
export enum PaymentStatus {
  PENDING = 'pending',
  PROCESSING = 'processing',
  COMPLETED = 'completed',
  FAILED = 'failed',
  CANCELLED = 'cancelled',
  REFUNDED = 'refunded'
}

/**
 * Notification types
 */
export enum NotificationType {
  LIKE = 'like',
  COMMENT = 'comment',
  FOLLOW = 'follow',
  MENTION = 'mention',
  REPOST = 'repost',
  ZAP = 'zap',
  PURCHASE = 'purchase',
  RELEASE = 'release'
}

/**
 * Sort options
 */
export type SortOrder = 'asc' | 'desc';

export interface SortOptions {
  field: string;
  order: SortOrder;
}

/**
 * Filter options
 */
export interface FilterOptions {
  search?: string;
  tags?: string[];
  dateFrom?: string;
  dateTo?: string;
  status?: ContentStatus;
  genre?: string;
  [key: string]: any;
}

/**
 * Pagination options
 */
export interface PaginationOptions {
  page: number;
  limit: number;
  sort?: SortOptions;
  filter?: FilterOptions;
}

/**
 * Location/geography types
 */
export interface Location {
  latitude: number;
  longitude: number;
  city?: string;
  country?: string;
  timezone?: string;
}

/**
 * Social media link types
 */
export interface SocialLinks {
  twitter?: string;
  instagram?: string;
  youtube?: string;
  bandcamp?: string;
  spotify?: string;
  soundcloud?: string;
  website?: string;
  [platform: string]: string | undefined;
}

/**
 * Color theme types
 */
export type ColorTheme = 'light' | 'dark' | 'auto';

/**
 * Language/locale types
 */
export type Language = 'en' | 'es' | 'fr' | 'de' | 'ja' | 'zh' | 'pt';
export type Locale = `${Language}-${Uppercase<string>}`;

/**
 * Configuration types
 */
export interface AppConfig {
  apiUrl: string;
  relayUrls: string[];
  environment: 'development' | 'staging' | 'production';
  features: {
    payments: boolean;
    social: boolean;
    analytics: boolean;
    beta: boolean;
  };
  limits: {
    maxFileSize: number;     // bytes
    maxPlaylistSize: number; // tracks
    maxBio: number;         // characters
    maxTrackTitle: number;  // characters
  };
}

/**
 * Utility types
 */
export type Optional<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>;
export type RequiredFields<T, K extends keyof T> = T & Required<Pick<T, K>>;
export type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P];
};

/**
 * Event emitter types
 */
export type EventHandler<T = any> = (data: T) => void | Promise<void>;
export type EventMap = Record<string, any>;

/**
 * WebSocket message types
 */
export interface WebSocketMessage<T = any> {
  type: string;
  data: T;
  id?: string;
  timestamp: number;
}

/**
 * Cache types
 */
export interface CacheEntry<T = any> {
  value: T;
  expiry: number;
  size?: number;
}

export interface CacheStats {
  hits: number;
  misses: number;
  size: number;
  maxSize: number;
}

/**
 * Validation types
 */
export interface ValidationError {
  field: string;
  message: string;
  code?: string;
}

export interface ValidationResult {
  valid: boolean;
  errors: ValidationError[];
  warnings?: ValidationError[];
}

/**
 * Analytics types
 */
export interface AnalyticsEvent {
  event: string;
  properties?: Record<string, any>;
  userId?: string;
  timestamp: number;
}

export interface AnalyticsSession {
  sessionId: string;
  userId?: string;
  startTime: number;
  endTime?: number;
  events: AnalyticsEvent[];
  metadata: Record<string, any>;
}