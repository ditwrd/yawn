import * as React from "react"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import { useNotification } from "./use-notification"

interface ConfirmDialogProps {
  title: string
  description: string
  onConfirm: () => void | Promise<void>
  onCancel?: () => void
  confirmText?: string
  cancelText?: string
  variant?: "default" | "destructive"
  children: React.ReactNode
  disabled?: boolean
}

export function ConfirmDialog({
  title,
  description,
  onConfirm,
  onCancel,
  confirmText = "Confirm",
  cancelText = "Cancel",
  variant = "default",
  children,
  disabled = false,
}: ConfirmDialogProps) {
  const { success, error } = useNotification()
  const [isLoading, setIsLoading] = React.useState(false)

  const handleConfirm = async () => {
    setIsLoading(true)
    try {
      await onConfirm()
      success("Action completed successfully")
    } catch (err) {
      error("Failed to complete action. Please try again.")
      console.error("Confirm dialog error:", err)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <AlertDialog>
      <AlertDialogTrigger asChild disabled={disabled}>
        {children}
      </AlertDialogTrigger>
      <AlertDialogContent className="bg-card border-border">
        <AlertDialogHeader>
          <AlertDialogTitle className="text-card-foreground">
            {title}
          </AlertDialogTitle>
          <AlertDialogDescription className="text-muted-foreground">
            {description}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel onClick={onCancel} disabled={isLoading}>
            {cancelText}
          </AlertDialogCancel>
          <AlertDialogAction
            onClick={handleConfirm}
            disabled={isLoading}
            className={variant === "destructive" ? "bg-destructive hover:bg-destructive/90" : ""}
          >
            {isLoading ? "Loading..." : confirmText}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}

interface ConfirmDeleteDialogProps {
  itemName: string
  itemType: string
  onConfirm: () => void | Promise<void>
  onCancel?: () => void
  children: React.ReactNode
  disabled?: boolean
}

export function ConfirmDeleteDialog({
  itemName,
  itemType,
  onConfirm,
  onCancel,
  children,
  disabled = false,
}: ConfirmDeleteDialogProps) {
  return (
    <ConfirmDialog
      title={`Delete ${itemType}`}
      description={`Are you sure you want to delete "${itemName}"? This action cannot be undone.`}
      onConfirm={onConfirm}
      onCancel={onCancel}
      confirmText="Delete"
      variant="destructive"
      disabled={disabled}
    >
      {children}
    </ConfirmDialog>
  )
}