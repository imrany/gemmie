import type { Theme } from "vue-sonner/src/packages/types.js";
import type { FunctionalComponent } from "vue";

export type ModeOption = {
  mode: "light-response" | "web-search" | "deep-search";
  label: string;
  description: string;
  icon: FunctionalComponent<any>;
  title: string;
};

export type ContextReference = {
  preview: string;
  fullText: string;
};

export interface Message {
  id: string;
  chat_id: string;
  prompt: string;
  response: string;
  created_at: string;
  model: string;
  requestId?: string;
  references: string[];
}

export interface Chat {
  id: string;
  user_id: string;
  title: string;
  created_at: string;
  updated_at: string;
  is_archived: boolean;
  message_count: number;
  last_message_at: string;
  messages: Message[];
  is_private: boolean;
  is_read_only?: boolean;
}

export interface PlatformError {
  message: string;
  description?: string;
  id: string;
  action: string;
  createdAt: string;
  status?: number | string;
  userId?: string;
  context?: Record<string, any>;
  severity?: "low" | "medium" | "high" | "critical";
}

export interface LinkPreview {
  url: string;
  title?: string;
  description?: string;
  images?: string[];
  previewImage?: string;
  domain?: string;
  favicon?: string;
  links?: string[];
  video?: string;
  videoType?: "youtube" | "vimeo" | "direct" | "twitter" | "tiktok";
  videoDuration?: string;
  videoThumbnail?: string;
  embedHtml?: string;
  loading?: boolean;
  error?: boolean;
}

export type ConfirmDialogOptions = {
  visible?: boolean;
  title: string;
  message: string;
  isLoading: boolean;
  type: "danger" | "warning" | "info";
  confirmText: string;
  cancelText: string;
  onConfirm: () => void;
  onCancel?: () => void;
};

export type RequestCount = {
  count: number;
  timestamp: number;
};

export interface Transaction {
  id: string;
  ExternalReference?: string;
  MpesaReceiptNumber?: string;
  CheckoutRequestID?: string;
  MerchantRequestID?: string;
  Amount: number;
  Phone: string;
  ResultCode: number;
  ResultDesc: string;
  Status: string;
  CreatedAt: Date;
  UpdatedAt: Date;
}

export interface UserDetails {
  userId: string;
  username: string;
  email: string;
  createdAt: Date;
  preferences?: string;
  workFunction?: string;
  theme?: Theme;
  syncEnabled: boolean;
  plan?: string;
  planName?: string;
  amount?: number;
  duration?: string;
  phoneNumber?: string;
  expiryTimestamp?: number;
  expireDuration?: number;
  price?: string;
  responseMode?: string;
  agreeToTerms?: boolean;
  requestCount?: RequestCount;
  emailVerified: boolean;
  emailSubscribed: boolean;
  sessionId?: string;
  userTransactions?: Transaction[];
}

// API response wrapper
export interface ApiResponse<T = any> {
  success: boolean;
  message: string;
  data?: T;
}

export interface Arcade {
  ID?: string;
  UserId: string;
  Code: string;
  Label: string;
  Description: string;
  CodeType: string;
  CreatedAt: string;
  UpdatedAt?: Date;
}

export interface RawArcade {
  id?: string;
  user_id: string;
  code: string;
  label: string;
  description: string;
  code_type: string;
  created_at: string;
  updated_at?: Date;
}

export interface CustomPayload {
  title: string;
  body: string;
  url?: string;
  icon?: string;
  tag?: "default-tag" | "notification-tag" | string;
  requireInteraction?: boolean;
}
