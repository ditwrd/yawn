import { toast } from "sonner"

type NotificationType = "success" | "error" | "warning" | "info"

interface NotificationOptions {
  duration?: number
  action?: {
    label: string
    onClick: () => void
  }
  dismissible?: boolean
}

export function useNotification() {
  const showNotification = (
    type: NotificationType,
    message: string,
    options?: NotificationOptions
  ) => {
    switch (type) {
      case "success":
        return toast.success(message, {
          duration: options?.duration || 4000,
          action: options?.action,
          dismissible: options?.dismissible !== false,
        })
      case "error":
        return toast.error(message, {
          duration: options?.duration || 6000,
          action: options?.action,
          dismissible: options?.dismissible !== false,
        })
      case "warning":
        return toast.warning(message, {
          duration: options?.duration || 5000,
          action: options?.action,
          dismissible: options?.dismissible !== false,
        })
      case "info":
        return toast.info(message, {
          duration: options?.duration || 4000,
          action: options?.action,
          dismissible: options?.dismissible !== false,
        })
      default:
        return toast(message, {
          duration: options?.duration || 4000,
          action: options?.action,
          dismissible: options?.dismissible !== false,
        })
    }
  }

  const success = (message: string, options?: NotificationOptions) =>
    showNotification("success", message, options)

  const error = (message: string, options?: NotificationOptions) =>
    showNotification("error", message, options)

  const warning = (message: string, options?: NotificationOptions) =>
    showNotification("warning", message, options)

  const info = (message: string, options?: NotificationOptions) =>
    showNotification("info", message, options)

  const promise = <T,>(
    promise: Promise<T>,
    messages: {
      loading: string
      success: string | ((data: T) => string)
      error: string | ((error: any) => string)
    }
  ) => {
    return toast.promise(promise, {
      loading: messages.loading,
      success: messages.success,
      error: messages.error,
    })
  }

  return {
    showNotification,
    success,
    error,
    warning,
    info,
    promise,
  }
}

export function notify(
  type: NotificationType,
  message: string,
  options?: NotificationOptions
) {
  const { showNotification } = useNotification()
  return showNotification(type, message, options)
}

export const notification = {
  success: (message: string, options?: NotificationOptions) =>
    notify("success", message, options),
  error: (message: string, options?: NotificationOptions) =>
    notify("error", message, options),
  warning: (message: string, options?: NotificationOptions) =>
    notify("warning", message, options),
  info: (message: string, options?: NotificationOptions) =>
    notify("info", message, options),
  promise: <T,>(
    promise: Promise<T>,
    messages: {
      loading: string
      success: string | ((data: T) => string)
      error: string | ((error: any) => string)
    }
  ) => {
    const { promise: promiseFn } = useNotification()
    return promiseFn(promise, messages)
  },
}