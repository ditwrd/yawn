import { create } from 'zustand'
import { subscribeWithSelector } from 'zustand/middleware'
import type { Project, ProjectMember } from '@/types/api'

interface ProjectState {
  // Current project context
  currentProjectId: string | null
  currentProject: Project | null

  // Project list
  projects: Project[]
  projectsLoading: boolean
  projectsError: string | null

  // Project members
  members: ProjectMember[]
  membersLoading: boolean
  membersError: string | null

  // Recent projects for quick access
  recentProjects: string[]
  maxRecentProjects: number

  // Actions
  setCurrentProject: (projectId: string | null) => void
  updateCurrentProject: (project: Project) => void
  setProjects: (projects: Project[]) => void
  addProject: (project: Project) => void
  updateProject: (projectId: string, updates: Partial<Project>) => void
  removeProject: (projectId: string) => void
  setProjectsLoading: (loading: boolean) => void
  setProjectsError: (error: string | null) => void

  // Members management
  setMembers: (members: ProjectMember[]) => void
  addMember: (member: ProjectMember) => void
  updateMember: (memberId: string, updates: Partial<ProjectMember>) => void
  removeMember: (memberId: string) => void
  setMembersLoading: (loading: boolean) => void
  setMembersError: (error: string | null) => void

  // Recent projects
  addToRecentProjects: (projectId: string) => void
  removeFromRecentProjects: (projectId: string) => void
  clearRecentProjects: () => void

  // Utilities
  clearProjectData: () => void
  getProjectById: (projectId: string) => Project | null
  getMemberById: (memberId: string) => ProjectMember | null
}

export const useProjectStore = create<ProjectState>()(
  subscribeWithSelector((set, get) => ({
    // Initial state
    currentProjectId: null,
    currentProject: null,
    projects: [],
    projectsLoading: false,
    projectsError: null,
    members: [],
    membersLoading: false,
    membersError: null,
    recentProjects: [],
    maxRecentProjects: 5,

    // Actions
    setCurrentProject: (projectId: string | null) => {
      const { projects } = get()
      const project = projectId ? projects.find(p => p.id === projectId) || null : null

      set({
        currentProjectId: projectId,
        currentProject: project,
        members: [], // Reset members when switching projects
      })

      if (projectId) {
        get().addToRecentProjects(projectId)
      }
    },

    updateCurrentProject: (project: Project) => {
      set({ currentProject: project })

      // Also update in projects list
      get().updateProject(project.id, project)
    },

    setProjects: (projects: Project[]) => {
      const { currentProjectId } = get()
      const currentProject = currentProjectId
        ? projects.find(p => p.id === currentProjectId) || null
        : null

      set({
        projects,
        currentProject,
        projectsError: null,
      })
    },

    addProject: (project: Project) => {
      set((state) => ({
        projects: [...state.projects, project],
        projectsError: null,
      }))
    },

    updateProject: (projectId: string, updates: Partial<Project>) => {
      set((state) => ({
        projects: state.projects.map(p =>
          p.id === projectId ? { ...p, ...updates } : p
        ),
        currentProject: state.currentProject?.id === projectId
          ? { ...state.currentProject, ...updates }
          : state.currentProject,
      }))
    },

    removeProject: (projectId: string) => {
      set((state) => ({
        projects: state.projects.filter(p => p.id !== projectId),
        currentProject: state.currentProject?.id === projectId
          ? null
          : state.currentProject,
        currentProjectId: state.currentProjectId === projectId
          ? null
          : state.currentProjectId,
      }))
    },

    setProjectsLoading: (loading: boolean) => {
      set({ projectsLoading: loading })
    },

    setProjectsError: (error: string | null) => {
      set({ projectsError: error })
    },

    // Members management
    setMembers: (members: ProjectMember[]) => {
      set({
        members,
        membersError: null,
      })
    },

    addMember: (member: ProjectMember) => {
      set((state) => ({
        members: [...state.members, member],
        membersError: null,
      }))
    },

    updateMember: (memberId: string, updates: Partial<ProjectMember>) => {
      set((state) => ({
        members: state.members.map(m =>
          m.user_id === memberId ? { ...m, ...updates } : m
        ),
      }))
    },

    removeMember: (memberId: string) => {
      set((state) => ({
        members: state.members.filter(m => m.user_id !== memberId),
      }))
    },

    setMembersLoading: (loading: boolean) => {
      set({ membersLoading: loading })
    },

    setMembersError: (error: string | null) => {
      set({ membersError: error })
    },

    // Recent projects
    addToRecentProjects: (projectId: string) => {
      set((state) => {
        const recentProjects = state.recentProjects.filter(id => id !== projectId)
        recentProjects.unshift(projectId)

        return {
          recentProjects: recentProjects.slice(0, state.maxRecentProjects)
        }
      })
    },

    removeFromRecentProjects: (projectId: string) => {
      set((state) => ({
        recentProjects: state.recentProjects.filter(id => id !== projectId)
      }))
    },

    clearRecentProjects: () => {
      set({ recentProjects: [] })
    },

    // Utilities
    clearProjectData: () => {
      set({
        currentProjectId: null,
        currentProject: null,
        members: [],
        membersError: null,
      })
    },

    getProjectById: (projectId: string) => {
      const { projects } = get()
      return projects.find(p => p.id === projectId) || null
    },

    getMemberById: (memberId: string) => {
      const { members } = get()
      return members.find(m => m.user_id === memberId) || null
    },
  }))
)

// Selectors for optimized re-renders
export const useCurrentProject = () => useProjectStore((state) => state.currentProject)
export const useCurrentProjectId = () => useProjectStore((state) => state.currentProjectId)
export const useProjects = () => useProjectStore((state) => state.projects)
export const useProjectsLoading = () => useProjectStore((state) => state.projectsLoading)
export const useProjectsError = () => useProjectStore((state) => state.projectsError)
export const useMembers = () => useProjectStore((state) => state.members)
export const useMembersLoading = () => useProjectStore((state) => state.membersLoading)
export const useRecentProjects = () => useProjectStore((state) => state.recentProjects)
export const useProjectActions = () => useProjectStore((state) => ({
  setCurrentProject: state.setCurrentProject,
  updateCurrentProject: state.updateCurrentProject,
  setProjects: state.setProjects,
  addProject: state.addProject,
  updateProject: state.updateProject,
  removeProject: state.removeProject,
  setProjectsLoading: state.setProjectsLoading,
  setProjectsError: state.setProjectsError,
  setMembers: state.setMembers,
  addMember: state.addMember,
  updateMember: state.updateMember,
  removeMember: state.removeMember,
  setMembersLoading: state.setMembersLoading,
  setMembersError: state.setMembersError,
  addToRecentProjects: state.addToRecentProjects,
  removeFromRecentProjects: state.removeFromRecentProjects,
  clearRecentProjects: state.clearRecentProjects,
  clearProjectData: state.clearProjectData,
  getProjectById: state.getProjectById,
  getMemberById: state.getMemberById,
}))