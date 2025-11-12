// Export all stores
export * from './auth-store'
export * from './preferences-store'
export * from './project-store'

import { useAuthStore } from './auth-store'
import { usePreferencesStore } from './preferences-store'
import { useProjectStore } from './project-store'

// Global store actions for cross-store operations
export const useGlobalStoreActions = () => {
  const { logout } = useAuthStore()
  const { resetPreferences } = usePreferencesStore()
  const { clearProjectData } = useProjectStore()

  return {
    // Global logout that clears all stores
    globalLogout: () => {
      logout()
      // Note: preferences are not cleared on logout
      clearProjectData()
    },

    // Global reset for development
    resetAllStores: () => {
      logout()
      resetPreferences()
      clearProjectData()
    },
  }
}