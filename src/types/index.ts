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