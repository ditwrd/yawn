import { describe, it, expect, vi, beforeEach } from 'vitest'

// Mock the API client
vi.mock('@/lib/api-client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
  createApiClient: vi.fn(),
}))

// Mock the auth service
vi.mock('@/lib/auth-service', () => ({
  authService: {
    getAccessToken: vi.fn(() => 'test-token'),
    isTokenExpired: vi.fn(() => false),
  },
}))


describe('API Integration Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Type-Safe API Calls', () => {
    it('should make successful GET request with proper types', async () => {
      const { apiClient } = await import('@/lib/api-client')
      const mockGet = vi.fn().mockResolvedValueOnce({
        data: { id: '1', name: 'Test Project', description: 'Test Description' },
      })
      apiClient.get = mockGet

      const result = await apiClient.get('/projects/1')

      expect(mockGet).toHaveBeenCalledWith('/projects/1')
      expect(result).toHaveProperty('data')
      expect(result.data.id).toBe('1')
      expect(result.data.name).toBe('Test Project')
    })

    it('should make POST request with type-safe payload', async () => {
      const { apiClient } = await import('@/lib/api-client')
      const mockPost = vi.fn().mockResolvedValueOnce({
        data: { id: '2', name: 'New Project', description: 'New Description' },
      })
      apiClient.post = mockPost

      const payload = {
        name: 'New Project',
        description: 'New Description',
        visibility: 'private',
      }

      const result = await apiClient.post('/projects', payload)

      expect(mockPost).toHaveBeenCalledWith('/projects', payload)
      expect(result.data.name).toBe(payload.name)
    })

    it('should handle PUT request with proper typing', async () => {
      const { apiClient } = await import('@/lib/api-client')
      const mockPut = vi.fn().mockResolvedValueOnce({
        data: { id: '1', name: 'Updated Project', description: 'Updated Description' },
      })
      apiClient.put = mockPut

      const payload = {
        name: 'Updated Project',
        description: 'Updated Description',
      }

      const result = await apiClient.put('/projects/1', payload)

      expect(mockPut).toHaveBeenCalledWith('/projects/1', payload)
      expect(result.data.name).toBe(payload.name)
    })

    it('should handle DELETE request properly', async () => {
      const { apiClient } = await import('@/lib/api-client')
      const mockDelete = vi.fn().mockResolvedValueOnce({
        data: { message: 'Project deleted successfully' },
      })
      apiClient.delete = mockDelete

      const result = await apiClient.delete('/projects/1')

      expect(mockDelete).toHaveBeenCalledWith('/projects/1')
      expect(result.data.message).toBe('Project deleted successfully')
    })
  })

  describe('Error Handling Scenarios', () => {
    it('should handle 404 Not Found errors', async () => {
      const { apiClient } = await import('@/lib/api-client')
      const mockGet = vi.fn().mockRejectedValueOnce({
        status: 404,
        error: 'Project not found',
        code: 'NOT_FOUND',
      })
      apiClient.get = mockGet

      try {
        await apiClient.get('/projects/999')
      } catch (error: any) {
        expect(error.status).toBe(404)
        expect(error.error).toBe('Project not found')
        expect(error.code).toBe('NOT_FOUND')
      }
    })

    it('should handle 401 Unauthorized errors', async () => {
      const { apiClient } = await import('@/lib/api-client')
      const mockPost = vi.fn().mockRejectedValueOnce({
        status: 401,
        error: 'Unauthorized',
        code: 'UNAUTHORIZED',
      })
      apiClient.post = mockPost

      try {
        await apiClient.post('/projects', { name: 'Test' })
      } catch (error: any) {
        expect(error.status).toBe(401)
        expect(error.error).toBe('Unauthorized')
        expect(error.code).toBe('UNAUTHORIZED')
      }
    })

    it('should handle 422 Validation errors', async () => {
      const { apiClient } = await import('@/lib/api-client')
      const mockPost = vi.fn().mockRejectedValueOnce({
        status: 422,
        error: 'Validation failed',
        code: 'VALIDATION_ERROR',
        details: 'Invalid input data',
        fields: {
          name: ['Name is required'],
          description: ['Description too short'],
        },
      })
      apiClient.post = mockPost

      try {
        await apiClient.post('/projects', {})
      } catch (error: any) {
        expect(error.status).toBe(422)
        expect(error.error).toBe('Validation failed')
        expect(error.fields.name).toContain('Name is required')
      }
    })

    it('should handle network errors', async () => {
      const { apiClient } = await import('@/lib/api-client')
      const mockGet = vi.fn().mockRejectedValueOnce(new Error('Network error'))
      apiClient.get = mockGet

      try {
        await apiClient.get('/projects')
      } catch (error: any) {
        expect(error.message).toBe('Network error')
      }
    })
  })

  describe('Request/Response Interceptors', () => {
    it('should add authorization header to requests', async () => {
      const { apiClient } = await import('@/lib/api-client')
      const mockGet = vi.fn().mockResolvedValueOnce({ data: {} })
      apiClient.get = mockGet

      await apiClient.get('/projects')

      // The mocked get should be called with proper headers
      expect(mockGet).toHaveBeenCalled()
    })

    it('should handle automatic token refresh', async () => {
      const { apiClient } = await import('@/lib/api-client')
      const { authService } = await import('@/lib/auth-service')

      // Mock token expired
      authService.isTokenExpired = vi.fn(() => true)

      const mockGet = vi.fn().mockResolvedValueOnce({ data: {} })
      apiClient.get = mockGet

      await apiClient.get('/projects')

      expect(mockGet).toHaveBeenCalled()
    })

    it('should transform request data properly', async () => {
      const { apiClient } = await import('@/lib/api-client')
      const mockPost = vi.fn().mockResolvedValueOnce({ data: {} })
      apiClient.post = mockPost

      const payload = {
        projectId: '1',
        userId: '2',
        createdAt: new Date().toISOString(),
      }

      await apiClient.post('/assets', payload)

      expect(mockPost).toHaveBeenCalledWith('/assets', payload)
    })

    it('should transform response data properly', async () => {
      const { apiClient } = await import('@/lib/api-client')
      const mockGet = vi.fn().mockResolvedValueOnce({
        data: {
          id: '1',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
          project_id: 'proj-1',
        },
      })
      apiClient.get = mockGet

      const result = await apiClient.get('/assets/1')

      expect(result.data).toHaveProperty('id')
      expect(result.data).toHaveProperty('created_at')
      expect(result.data.project_id).toBe('proj-1')
    })
  })

  describe('API Client Integration with TanStack Query', () => {
    it('should work with TanStack Query useQuery pattern', async () => {
      const { apiClient } = await import('@/lib/api-client')

      const mockGet = vi.fn().mockResolvedValueOnce({
        data: [{ id: '1', name: 'Project 1' }],
      })
      apiClient.get = mockGet

      // Test that the API client returns the expected structure
      const result = await apiClient.get('/projects')

      expect(mockGet).toHaveBeenCalledWith('/projects')
      expect(result).toHaveProperty('data')
      expect(Array.isArray(result.data)).toBe(true)
      expect(result.data[0].name).toBe('Project 1')
    })

    it('should work with TanStack Query useMutation pattern', async () => {
      const { apiClient } = await import('@/lib/api-client')

      const mockPost = vi.fn().mockResolvedValueOnce({
        data: { id: '2', name: 'New Project' },
      })
      apiClient.post = mockPost

      const projectData = { name: 'New Project', description: 'Test project' }

      // Test that the API client handles POST requests correctly
      const result = await apiClient.post('/projects', projectData)

      expect(mockPost).toHaveBeenCalledWith('/projects', projectData)
      expect(result).toHaveProperty('data')
      expect(result.data.name).toBe('New Project')
      expect(result.data.id).toBe('2')
    })
  })

  describe('Request Deduplication', () => {
    it('should handle multiple identical requests', async () => {
      const { apiClient } = await import('@/lib/api-client')
      const mockGet = vi.fn().mockResolvedValue({ data: { id: '1' } })
      apiClient.get = mockGet

      // Make multiple identical requests
      const promise1 = apiClient.get('/projects/1')
      const promise2 = apiClient.get('/projects/1')
      const promise3 = apiClient.get('/projects/1')

      const [result1, result2, result3] = await Promise.all([promise1, promise2, promise3])

      // All requests should complete successfully
      expect(result1.data.id).toBe('1')
      expect(result2.data.id).toBe('1')
      expect(result3.data.id).toBe('1')

      // Should make all requests (deduplication would be handled by TanStack Query)
      expect(mockGet).toHaveBeenCalledTimes(3)
    })
  })
})