import type { Theme } from "vue-sonner/src/packages/types.js"

export interface Res {
  prompt?: string
  response: string
  status?: number
  requestId?: string 
}

export type Chat = {
  id: string,
  title: string,
  messages: Res[],
  createdAt: string,
  updatedAt: string
}

export interface LinkPreview {
  url: string
  title?: string
  description?: string
  images?: string[]
  previewImage?: string
  domain?: string
  favicon?: string
  links?: string[]
  video?: string
  videoType?: 'youtube' | 'vimeo' | 'direct' | 'twitter' | 'tiktok'
  videoDuration?: string
  videoThumbnail?: string
  embedHtml?: string
  loading?: boolean
  error?: boolean
}

export type CurrentChat ={
    id: string;
    title: string;
    messages: Res[];
    createdAt: string;
    updatedAt: string
}

export type ConfirmDialogOptions = {
    visible?: boolean
    title: string
    message: string
    type?: 'danger' | 'warning' | 'info'
    confirmText?: string
    cancelText?: string
    onConfirm: () => void
    onCancel?: () => void
}

export type RequestCount ={
	count: number,
	timestamp: number
}

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
  sessionId?: string,
  userTransactions?: Transaction[];
}