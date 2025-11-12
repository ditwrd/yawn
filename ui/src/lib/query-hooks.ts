import {
  useQuery,
  useMutation,
  useQueryClient,
  useInfiniteQuery,
  QueryOptions,
  MutationOptions,
  InfiniteQueryOptions,
} from '@tanstack/react-query'
import { apiClient } from './api-client'
import { isRecoverableError } from './error-handling'
import type {
  AssetResponse,
  AssetListResponse,
  CreateAssetRequest,
  UpdateAssetRequest,
  ProjectResponse,
  ProjectListResponse,
  CreateProjectRequest,
  UpdateProjectRequest,
  PipelineResponse,
  PipelineListResponse,
  CreatePipelineRequest,
  UpdatePipelineRequest,
} from '@shared/dto'

// Query key factory for consistent cache keys
export const queryKeys = {
  // Assets
  assets: ['assets'] as const,
  asset: (id: string) => ['assets', id] as const,
  assetsByProject: (projectId: string) => ['assets', 'project', projectId] as const,
  assetsSearch: (query: string) => ['assets', 'search', query] as const,

  // Projects
  projects: ['projects'] as const,
  project: (id: string) => ['projects', id] as const,
  projectsSearch: (query: string) => ['projects', 'search', query] as const,

  // Pipelines
  pipelines: ['pipelines'] as const,
  pipeline: (id: string) => ['pipelines', id] as const,
  pipelinesByProject: (projectId: string) => ['pipelines', 'project', projectId] as const,
  pipelineExecutions: (pipelineId: string) => ['pipelines', pipelineId, 'executions'] as const,

  // Repositories
  repositories: ['repositories'] as const,
  repository: (id: string) => ['repositories', id] as const,
  repositoriesByProject: (projectId: string) => ['repositories', 'project', projectId] as const,

  // Users
  users: ['users'] as const,
  user: (id: string) => ['users', id] as const,
} as const

// Default query options with retry logic
const defaultQueryOptions = {
  retry: (failureCount: number, error: any) => {
    // Don't retry on 4xx errors
    if (error?.status >= 400 && error?.status < 500) {
      return false
    }
    // Don't retry non-recoverable errors
    if (!isRecoverableError(error)) {
      return false
    }
    // Retry up to 3 times for other errors
    return failureCount < 3
  },
  retryDelay: (attemptIndex: number) => Math.min(1000 * 2 ** attemptIndex, 30000),
  staleTime: 5 * 60 * 1000, // 5 minutes
  gcTime: 10 * 60 * 1000, // 10 minutes
  refetchOnWindowFocus: false,
  refetchOnReconnect: true,
  refetchOnMount: true,
}

// Default mutation options
const defaultMutationOptions = {
  retry: (failureCount: number, error: any) => {
    // Don't retry mutations on 4xx errors
    if (error?.status >= 400 && error?.status < 500) {
      return false
    }
    // Don't retry non-recoverable errors
    if (!isRecoverableError(error)) {
      return false
    }
    return failureCount < 2 // Retry mutations up to 2 times
  },
  retryDelay: (attemptIndex: number) => Math.min(1000 * 2 ** attemptIndex, 10000),
}

// Asset hooks
export function useAssets(params?: { page?: number; limit?: number; project_id?: string }) {
  return useQuery({
    queryKey: [...queryKeys.assets, params],
    queryFn: () => apiClient.getAssets(params).then(res => res.data),
    ...defaultQueryOptions,
  })
}

export function useAsset(id: string, enabled: boolean = true) {
  return useQuery({
    queryKey: queryKeys.asset(id),
    queryFn: () => apiClient.getAsset(id).then(res => res.data),
    enabled: enabled && !!id,
    ...defaultQueryOptions,
  })
}

export function useAssetsByProject(projectId: string) {
  return useQuery({
    queryKey: queryKeys.assetsByProject(projectId),
    queryFn: () => apiClient.getAssets({ project_id: projectId }).then(res => res.data),
    enabled: !!projectId,
    ...defaultQueryOptions,
  })
}

export function useCreateAsset() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateAssetRequest) => apiClient.createAsset(data).then(res => res.data),
    onSuccess: (newAsset) => {
      // Invalidate and refetch assets list
      queryClient.invalidateQueries({ queryKey: queryKeys.assets })

      // If asset has a project, invalidate project-specific assets
      if (newAsset.project_id) {
        queryClient.invalidateQueries({ queryKey: queryKeys.assetsByProject(newAsset.project_id) })
      }

      // Add the new asset to the cache
      queryClient.setQueryData(queryKeys.asset(newAsset.id), newAsset)
    },
    ...defaultMutationOptions,
  })
}

export function useUpdateAsset() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateAssetRequest }) =>
      apiClient.updateAsset(id, data).then(res => res.data),
    onSuccess: (updatedAsset) => {
      // Update the asset in cache
      queryClient.setQueryData(queryKeys.asset(updatedAsset.id), updatedAsset)

      // Invalidate related queries
      queryClient.invalidateQueries({ queryKey: queryKeys.assets })

      if (updatedAsset.project_id) {
        queryClient.invalidateQueries({ queryKey: queryKeys.assetsByProject(updatedAsset.project_id) })
      }
    },
    ...defaultMutationOptions,
  })
}

export function useDeleteAsset() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => apiClient.deleteAsset(id).then(res => res.data),
    onSuccess: (_, deletedId) => {
      // Remove the asset from cache
      queryClient.removeQueries({ queryKey: queryKeys.asset(deletedId) })

      // Invalidate related queries
      queryClient.invalidateQueries({ queryKey: queryKeys.assets })
    },
    ...defaultMutationOptions,
  })
}

// Project hooks
export function useProjects(params?: { page?: number; limit?: number }) {
  return useQuery({
    queryKey: [...queryKeys.projects, params],
    queryFn: () => apiClient.getProjects(params).then(res => res.data),
    ...defaultQueryOptions,
  })
}

export function useProject(id: string, enabled: boolean = true) {
  return useQuery({
    queryKey: queryKeys.project(id),
    queryFn: () => apiClient.getProject(id).then(res => res.data),
    enabled: enabled && !!id,
    ...defaultQueryOptions,
  })
}

export function useCreateProject() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateProjectRequest) => apiClient.createProject(data).then(res => res.data),
    onSuccess: (newProject) => {
      // Invalidate projects list
      queryClient.invalidateQueries({ queryKey: queryKeys.projects })

      // Add the new project to cache
      queryClient.setQueryData(queryKeys.project(newProject.id), newProject)
    },
    ...defaultMutationOptions,
  })
}

export function useUpdateProject() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateProjectRequest }) =>
      apiClient.updateProject(id, data).then(res => res.data),
    onSuccess: (updatedProject) => {
      // Update the project in cache
      queryClient.setQueryData(queryKeys.project(updatedProject.id), updatedProject)

      // Invalidate projects list
      queryClient.invalidateQueries({ queryKey: queryKeys.projects })
    },
    ...defaultMutationOptions,
  })
}

export function useDeleteProject() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => apiClient.deleteProject(id).then(res => res.data),
    onSuccess: (_, deletedId) => {
      // Remove the project from cache
      queryClient.removeQueries({ queryKey: queryKeys.project(deletedId) })

      // Invalidate projects list
      queryClient.invalidateQueries({ queryKey: queryKeys.projects })

      // Also invalidate assets and pipelines for this project
      queryClient.invalidateQueries({ queryKey: ['assets', 'project', deletedId] })
      queryClient.invalidateQueries({ queryKey: ['pipelines', 'project', deletedId] })
    },
    ...defaultMutationOptions,
  })
}

// Pipeline hooks
export function usePipelines(params?: { page?: number; limit?: number; project_id?: string }) {
  return useQuery({
    queryKey: [...queryKeys.pipelines, params],
    queryFn: () => apiClient.getPipelines(params).then(res => res.data),
    ...defaultQueryOptions,
  })
}

export function usePipeline(id: string, enabled: boolean = true) {
  return useQuery({
    queryKey: queryKeys.pipeline(id),
    queryFn: () => apiClient.getPipeline(id).then(res => res.data),
    enabled: enabled && !!id,
    ...defaultQueryOptions,
  })
}

export function usePipelinesByProject(projectId: string) {
  return useQuery({
    queryKey: queryKeys.pipelinesByProject(projectId),
    queryFn: () => apiClient.getPipelines({ project_id: projectId }).then(res => res.data),
    enabled: !!projectId,
    ...defaultQueryOptions,
  })
}

export function useCreatePipeline() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreatePipelineRequest) => apiClient.createPipeline(data).then(res => res.data),
    onSuccess: (newPipeline) => {
      // Invalidate pipelines list
      queryClient.invalidateQueries({ queryKey: queryKeys.pipelines })

      // If pipeline has a project, invalidate project-specific pipelines
      if (newPipeline.project_id) {
        queryClient.invalidateQueries({ queryKey: queryKeys.pipelinesByProject(newPipeline.project_id) })
      }

      // Add the new pipeline to the cache
      queryClient.setQueryData(queryKeys.pipeline(newPipeline.id), newPipeline)
    },
    ...defaultMutationOptions,
  })
}

export function useUpdatePipeline() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdatePipelineRequest }) =>
      apiClient.updatePipeline(id, data).then(res => res.data),
    onSuccess: (updatedPipeline) => {
      // Update the pipeline in cache
      queryClient.setQueryData(queryKeys.pipeline(updatedPipeline.id), updatedPipeline)

      // Invalidate related queries
      queryClient.invalidateQueries({ queryKey: queryKeys.pipelines })

      if (updatedPipeline.project_id) {
        queryClient.invalidateQueries({ queryKey: queryKeys.pipelinesByProject(updatedPipeline.project_id) })
      }
    },
    ...defaultMutationOptions,
  })
}

export function useDeletePipeline() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => apiClient.deletePipeline(id).then(res => res.data),
    onSuccess: (_, deletedId) => {
      // Remove the pipeline from cache
      queryClient.removeQueries({ queryKey: queryKeys.pipeline(deletedId) })

      // Invalidate pipelines list
      queryClient.invalidateQueries({ queryKey: queryKeys.pipelines })
    },
    ...defaultMutationOptions,
  })
}

export function useTriggerPipeline() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data?: any }) => apiClient.triggerPipeline(id, data).then(res => res.data),
    onSuccess: (_, { id }) => {
      // Invalidate pipeline data and executions
      queryClient.invalidateQueries({ queryKey: queryKeys.pipeline(id) })
      queryClient.invalidateQueries({ queryKey: queryKeys.pipelineExecutions(id) })
    },
    ...defaultMutationOptions,
  })
}

// Prefetching utilities
export function usePrefetchAsset() {
  const queryClient = useQueryClient()

  return (id: string) => {
    queryClient.prefetchQuery({
      queryKey: queryKeys.asset(id),
      queryFn: () => apiClient.getAsset(id).then(res => res.data),
      ...defaultQueryOptions,
    })
  }
}

export function usePrefetchProject() {
  const queryClient = useQueryClient()

  return (id: string) => {
    queryClient.prefetchQuery({
      queryKey: queryKeys.project(id),
      queryFn: () => apiClient.getProject(id).then(res => res.data),
      ...defaultQueryOptions,
    })
  }
}

// Cache management utilities
export function useInvalidateAssets() {
  const queryClient = useQueryClient()

  return () => {
    queryClient.invalidateQueries({ queryKey: queryKeys.assets })
  }
}

export function useInvalidateProjects() {
  const queryClient = useQueryClient()

  return () => {
    queryClient.invalidateQueries({ queryKey: queryKeys.projects })
  }
}

export function useInvalidatePipelines() {
  const queryClient = useQueryClient()

  return () => {
    queryClient.invalidateQueries({ queryKey: queryKeys.pipelines })
  }
}

// Optimistic update utilities
export function useOptimisticAssetUpdate() {
  const queryClient = useQueryClient()

  return (id: string, optimisticData: Partial<AssetResponse>) => {
    queryClient.setQueryData(queryKeys.asset(id), (old: AssetResponse | undefined) =>
      old ? { ...old, ...optimisticData } : optimisticData
    )
  }
}

export function useOptimisticProjectUpdate() {
  const queryClient = useQueryClient()

  return (id: string, optimisticData: Partial<ProjectResponse>) => {
    queryClient.setQueryData(queryKeys.project(id), (old: ProjectResponse | undefined) =>
      old ? { ...old, ...optimisticData } : optimisticData
    )
  }
}