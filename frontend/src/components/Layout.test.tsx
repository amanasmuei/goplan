import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

// Mock the auth store
vi.mock('../store/authStore', () => ({
  useAuthStore: () => ({
    user: { id: '1', name: 'Test User', email: 'test@example.com' },
    logout: vi.fn(),
    isAuthenticated: true,
  }),
}))

// Mock the api
vi.mock('../services/api', () => ({
  taskApi: {
    getStats: vi.fn().mockResolvedValue({
      total: 10,
      by_status: { active: 5, completed: 3, blocked: 2 },
    }),
  },
}))

import Layout from './Layout'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
    },
  },
})

const renderWithProviders = (children: React.ReactNode) => {
  return render(
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>{children}</BrowserRouter>
    </QueryClientProvider>
  )
}

describe('Layout', () => {
  it('renders navigation items', () => {
    renderWithProviders(<Layout />)

    expect(screen.getByText('Dashboard')).toBeInTheDocument()
    expect(screen.getByText('Tasks')).toBeInTheDocument()
  })

  it('renders the GoPlan logo', () => {
    renderWithProviders(<Layout />)

    expect(screen.getByText('GoPlan')).toBeInTheDocument()
  })
})
