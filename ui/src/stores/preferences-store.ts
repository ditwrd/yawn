import { create } from 'zustand'
import { persist, subscribeWithSelector } from 'zustand/middleware'
import type { Theme, ThemeMode, ColorScheme } from '@/types'

interface PreferencesState {
  // Theme preferences
  theme: Theme

  // UI preferences
  sidebarCollapsed: boolean
  fontSize: 'sm' | 'base' | 'lg'

  // Navigation preferences
  showBreadcrumbs: boolean
  autoCollapseSidebar: boolean

  // Notification preferences
  notifications: {
    desktop: boolean
    email: boolean
    sound: boolean
  }

  // Performance preferences
  reduceAnimations: boolean
  enableDevTools: boolean

  // Actions
  setTheme: (theme: Partial<Theme>) => void
  setThemeMode: (mode: ThemeMode) => void
  setColorScheme: (scheme: ColorScheme) => void
  setFontSize: (size: 'sm' | 'base' | 'lg') => void
  toggleSidebar: () => void
  setSidebarCollapsed: (collapsed: boolean) => void
  setShowBreadcrumbs: (show: boolean) => void
  setAutoCollapseSidebar: (auto: boolean) => void
  updateNotifications: (notifications: Partial<PreferencesState['notifications']>) => void
  setReduceAnimations: (reduce: boolean) => void
  setEnableDevTools: (enable: boolean) => void
  resetPreferences: () => void
}

const defaultTheme: Theme = {
  mode: 'dark',
  colorScheme: 'purple',
  fontSize: 'base',
}

const defaultNotifications = {
  desktop: true,
  email: false,
  sound: false,
}

export const usePreferencesStore = create<PreferencesState>()(
  subscribeWithSelector(
    persist(
      (set) => ({
        // Initial state
        theme: defaultTheme,
        sidebarCollapsed: false,
        fontSize: 'base',
        showBreadcrumbs: true,
        autoCollapseSidebar: false,
        notifications: defaultNotifications,
        reduceAnimations: false,
        enableDevTools: import.meta.env.DEV,

        // Actions
        setTheme: (themeUpdate: Partial<Theme>) => {
          set((state) => ({
            theme: { ...state.theme, ...themeUpdate }
          }))
        },

        setThemeMode: (mode: ThemeMode) => {
          set((state) => ({
            theme: { ...state.theme, mode }
          }))
        },

        setColorScheme: (scheme: ColorScheme) => {
          set((state) => ({
            theme: { ...state.theme, colorScheme: scheme }
          }))
        },

        setFontSize: (size: 'sm' | 'base' | 'lg') => {
          set((state) => ({
            theme: { ...state.theme, fontSize: size },
            fontSize: size
          }))
        },

        toggleSidebar: () => {
          set((state) => ({
            sidebarCollapsed: !state.sidebarCollapsed
          }))
        },

        setSidebarCollapsed: (collapsed: boolean) => {
          set({ sidebarCollapsed: collapsed })
        },

        setShowBreadcrumbs: (show: boolean) => {
          set({ showBreadcrumbs: show })
        },

        setAutoCollapseSidebar: (auto: boolean) => {
          set({ autoCollapseSidebar: auto })
        },

        updateNotifications: (notificationsUpdate: Partial<PreferencesState['notifications']>) => {
          set((state) => ({
            notifications: { ...state.notifications, ...notificationsUpdate }
          }))
        },

        setReduceAnimations: (reduce: boolean) => {
          set({ reduceAnimations: reduce })
        },

        setEnableDevTools: (enable: boolean) => {
          set({ enableDevTools: enable })
        },

        resetPreferences: () => {
          set({
            theme: defaultTheme,
            sidebarCollapsed: false,
            fontSize: 'base',
            showBreadcrumbs: true,
            autoCollapseSidebar: false,
            notifications: defaultNotifications,
            reduceAnimations: false,
            enableDevTools: import.meta.env.DEV,
          })
        },
      }),
      {
        name: 'preferences-store',
        partialize: (state) => ({
          theme: state.theme,
          sidebarCollapsed: state.sidebarCollapsed,
          fontSize: state.fontSize,
          showBreadcrumbs: state.showBreadcrumbs,
          autoCollapseSidebar: state.autoCollapseSidebar,
          notifications: state.notifications,
          reduceAnimations: state.reduceAnimations,
          enableDevTools: state.enableDevTools,
        }),
      }
    )
  )
)

// Selectors for optimized re-renders
export const useTheme = () => usePreferencesStore((state) => state.theme)
export const useThemeMode = () => usePreferencesStore((state) => state.theme.mode)
export const useColorScheme = () => usePreferencesStore((state) => state.theme.colorScheme)
export const useSidebarCollapsed = () => usePreferencesStore((state) => state.sidebarCollapsed)
export const useFontSize = () => usePreferencesStore((state) => state.fontSize)
export const useShowBreadcrumbs = () => usePreferencesStore((state) => state.showBreadcrumbs)
export const useNotifications = () => usePreferencesStore((state) => state.notifications)
export const useReduceAnimations = () => usePreferencesStore((state) => state.reduceAnimations)
export const usePreferencesActions = () => usePreferencesStore((state) => ({
  setTheme: state.setTheme,
  setThemeMode: state.setThemeMode,
  setColorScheme: state.setColorScheme,
  setFontSize: state.setFontSize,
  toggleSidebar: state.toggleSidebar,
  setSidebarCollapsed: state.setSidebarCollapsed,
  setShowBreadcrumbs: state.setShowBreadcrumbs,
  updateNotifications: state.updateNotifications,
  setReduceAnimations: state.setReduceAnimations,
  resetPreferences: state.resetPreferences,
}))