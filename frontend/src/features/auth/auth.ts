import { useQuery } from '@tanstack/react-query'
import { api, ApiError } from '../../lib/api'

export type User = {
  id: string
  email: string
  displayName: string
  isAdmin: boolean
  createdAt: string
}

export const meKey = ['me'] as const

// A 401 means "not logged in", which is a normal state, not an error.
async function fetchMe(): Promise<User | null> {
  try {
    return await api.get<User>('/auth/me')
  } catch (e) {
    if (e instanceof ApiError && e.status === 401) return null
    throw e
  }
}

export function useMe() {
  return useQuery({ queryKey: meKey, queryFn: fetchMe, staleTime: 60_000, retry: false })
}

export const login = (email: string, password: string) =>
  api.post<User>('/auth/login', { email, password })

export const logout = () => api.post<void>('/auth/logout')
