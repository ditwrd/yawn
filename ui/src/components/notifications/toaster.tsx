import { Toaster } from "@/components/ui/sonner"
import { useTheme } from "@/components/theme-provider"

export function YAWNToaster() {
  const { theme } = useTheme()

  return (
    <Toaster
      theme={theme === "dark" ? "dark" : theme === "light" ? "light" : "system"}
      richColors
      toastOptions={{
        style: {
          background: "var(--card)",
          color: "var(--card-foreground)",
          border: "1px solid var(--border)",
          fontFamily: "'JetBrains Mono', monospace",
        },
        className: "rounded-lg border-border",
      }}
      position="top-right"
      expand={false}
    />
  )
}