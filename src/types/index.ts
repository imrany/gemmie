
export type Res = {
  response: string,
  prompt?: string,
  status?: number,
}

export type Chat = {
  id: string,
  title: string,
  messages: Res[],
  createdAt: string,
  updatedAt: string
}

export type LinkPreview = {
  url: string,
  title?: string,
  description?: string,
  image?: string,
  domain?: string,
  loading?: boolean,
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
    visible: boolean
    title: string
    message: string
    type?: 'danger' | 'warning' | 'info'
    confirmText?: string
    cancelText?: string
    onConfirm: () => void
}