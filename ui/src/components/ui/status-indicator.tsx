import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

const statusIndicatorVariants = cva(
  "inline-flex items-center gap-2 rounded-full px-2.5 py-1 text-xs font-medium",
  {
    variants: {
      status: {
        success: "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20",
        warning: "bg-amber-500/10 text-amber-400 border border-amber-500/20",
        error: "bg-red-500/10 text-red-400 border border-red-500/20",
        info: "bg-blue-500/10 text-blue-400 border border-blue-500/20",
        pending: "bg-gray-500/10 text-gray-400 border border-gray-500/20",
        running: "bg-purple-500/10 text-purple-400 border border-purple-500/20",
      },
    },
    defaultVariants: {
      status: "info",
    },
  }
)

export interface StatusIndicatorProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof statusIndicatorVariants> {
  label: string
  showIcon?: boolean
}

export function StatusIndicator({
  status,
  label,
  showIcon = true,
  className,
  ...props
}: StatusIndicatorProps) {
  const getStatusIcon = () => {
    switch (status) {
      case "success":
        return "●"
      case "warning":
        return "▲"
      case "error":
        return "●"
      case "info":
        return "◆"
      case "pending":
        return "○"
      case "running":
        return "◍"
      default:
        return "●"
    }
  }

  return (
    <div
      className={cn(statusIndicatorVariants({ status }), className)}
      {...props}
    >
      {showIcon && <span className="text-[10px]">{getStatusIcon()}</span>}
      <span>{label}</span>
    </div>
  )
}