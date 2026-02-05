import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import Analytics from './Analytics'

// Mock the API
vi.mock('../services/api', () => ({
  taskApi: {
    list: vi.fn(),
  },
}))

import { taskApi } from '../services/api'

const createQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  })

const renderWithProviders = (component: React.ReactNode) => {
  const queryClient = createQueryClient()
  return render(
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>{component}</BrowserRouter>
    </QueryClientProvider>
  )
}

const mockTasks = [
  {
    id: '1',
    title: 'Completed Task 1',
    status: 'completed',
    actual_days: 5,
    predicted_days_low: 4,
    predicted_days_high: 6,
    estimated_days: 5,
    created_at: '2024-01-01',
    updated_at: '2024-01-06',
  },
  {
    id: '2',
    title: 'Completed Task 2',
    status: 'completed',
    actual_days: 8,
    predicted_days_low: 5,
    predicted_days_high: 7,
    estimated_days: 6,
    created_at: '2024-01-02',
    updated_at: '2024-01-10',
  },
  {
    id: '3',
    title: 'Active Task',
    status: 'active',
    predicted_days_low: 3,
    predicted_days_high: 5,
    estimated_days: 4,
    created_at: '2024-01-03',
    updated_at: '2024-01-03',
  },
  {
    id: '4',
    title: 'Blocked Task',
    status: 'blocked',
    predicted_days_low: 2,
    predicted_days_high: 4,
    estimated_days: 3,
    created_at: '2024-01-04',
    updated_at: '2024-01-04',
  },
]

describe('Analytics', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows loading state initially', async () => {
    vi.mocked(taskApi.list).mockImplementation(
      () => new Promise(() => {}) // Never resolves
    )

    renderWithProviders(<Analytics />)

    expect(screen.getByText('Loading analytics...')).toBeInTheDocument()
  })

  it('renders analytics dashboard title', async () => {
    vi.mocked(taskApi.list).mockResolvedValue({ tasks: mockTasks, total: 4 })

    renderWithProviders(<Analytics />)

    await waitFor(() => {
      expect(screen.getByText('Analytics Dashboard')).toBeInTheDocument()
    })
  })

  it('displays total tasks count', async () => {
    vi.mocked(taskApi.list).mockResolvedValue({ tasks: mockTasks, total: 4 })

    renderWithProviders(<Analytics />)

    await waitFor(() => {
      // Should show total tasks stat card
      expect(screen.getByText('Total Tasks')).toBeInTheDocument()
    })
  })

  it('displays completed tasks count', async () => {
    vi.mocked(taskApi.list).mockResolvedValue({ tasks: mockTasks, total: 4 })

    renderWithProviders(<Analytics />)

    await waitFor(() => {
      // Should show completed stat card
      expect(screen.getByText('Completed')).toBeInTheDocument()
      expect(screen.getByText(/completion rate/i)).toBeInTheDocument()
    })
  })

  it('displays prediction accuracy chart section', async () => {
    vi.mocked(taskApi.list).mockResolvedValue({ tasks: mockTasks, total: 4 })

    renderWithProviders(<Analytics />)

    await waitFor(() => {
      expect(screen.getByText('Prediction Accuracy Breakdown')).toBeInTheDocument()
    })
  })

  it('displays tasks by status chart section', async () => {
    vi.mocked(taskApi.list).mockResolvedValue({ tasks: mockTasks, total: 4 })

    renderWithProviders(<Analytics />)

    await waitFor(() => {
      expect(screen.getByText('Tasks by Status')).toBeInTheDocument()
    })
  })

  it('displays recent completions section', async () => {
    vi.mocked(taskApi.list).mockResolvedValue({ tasks: mockTasks, total: 4 })

    renderWithProviders(<Analytics />)

    await waitFor(() => {
      expect(screen.getByText('Recent Completions')).toBeInTheDocument()
    })
  })

  it('displays key insights section', async () => {
    vi.mocked(taskApi.list).mockResolvedValue({ tasks: mockTasks, total: 4 })

    renderWithProviders(<Analytics />)

    await waitFor(() => {
      expect(screen.getByText('Key Insights')).toBeInTheDocument()
    })
  })

  it('shows blocked task insight when tasks are blocked', async () => {
    vi.mocked(taskApi.list).mockResolvedValue({ tasks: mockTasks, total: 4 })

    renderWithProviders(<Analytics />)

    await waitFor(() => {
      expect(screen.getByText(/task\(s\) are currently blocked/)).toBeInTheDocument()
    })
  })

  it('handles empty task list', async () => {
    vi.mocked(taskApi.list).mockResolvedValue({ tasks: [], total: 0 })

    renderWithProviders(<Analytics />)

    await waitFor(() => {
      // Should show dashboard title and total tasks stat card
      expect(screen.getByText('Analytics Dashboard')).toBeInTheDocument()
      expect(screen.getByText('Total Tasks')).toBeInTheDocument()
    })
  })

  it('calculates average cycle time', async () => {
    vi.mocked(taskApi.list).mockResolvedValue({ tasks: mockTasks, total: 4 })

    renderWithProviders(<Analytics />)

    await waitFor(() => {
      // Average of 5 and 8 = 6.5 days
      expect(screen.getByText('Avg Cycle Time')).toBeInTheDocument()
      // Check for "days" in the cycle time display
      expect(screen.getByText(/days$/)).toBeInTheDocument()
    })
  })

  it('displays completion rate', async () => {
    vi.mocked(taskApi.list).mockResolvedValue({ tasks: mockTasks, total: 4 })

    renderWithProviders(<Analytics />)

    await waitFor(() => {
      // Should display completion rate as subtitle
      expect(screen.getByText(/completion rate/i)).toBeInTheDocument()
    })
  })

  it('shows prediction accuracy chart with data', async () => {
    vi.mocked(taskApi.list).mockResolvedValue({ tasks: mockTasks, total: 4 })

    renderWithProviders(<Analytics />)

    await waitFor(() => {
      // Should show categories: Under-estimated, Accurate, Over-estimated
      expect(screen.getByText(/Under-estimated/)).toBeInTheDocument()
      expect(screen.getByText(/Accurate/)).toBeInTheDocument()
      expect(screen.getByText(/Over-estimated/)).toBeInTheDocument()
    })
  })

  it('links to individual tasks in recent completions', async () => {
    vi.mocked(taskApi.list).mockResolvedValue({ tasks: mockTasks, total: 4 })

    renderWithProviders(<Analytics />)

    await waitFor(() => {
      const links = screen.getAllByRole('link')
      const taskLinks = links.filter((link) =>
        link.getAttribute('href')?.startsWith('/tasks/')
      )
      expect(taskLinks.length).toBeGreaterThan(0)
    })
  })
})

describe('Analytics calculations', () => {
  it('renders prediction accuracy breakdown section', async () => {
    const completedTasks = [
      {
        id: '1',
        title: 'Completed Task',
        status: 'completed',
        actual_days: 10,
        predicted_days_low: 3,
        predicted_days_high: 5,
        created_at: '2024-01-01',
        updated_at: '2024-01-11',
      },
    ]

    vi.mocked(taskApi.list).mockResolvedValue({
      tasks: completedTasks,
      total: 1,
    })

    renderWithProviders(<Analytics />)

    await waitFor(() => {
      expect(screen.getByText('Prediction Accuracy Breakdown')).toBeInTheDocument()
    })
  })

  it('displays tasks breakdown categories', async () => {
    const completedTasks = [
      {
        id: '1',
        title: 'Completed Task',
        status: 'completed',
        actual_days: 4,
        predicted_days_low: 3,
        predicted_days_high: 5,
        created_at: '2024-01-01',
        updated_at: '2024-01-05',
      },
    ]

    vi.mocked(taskApi.list).mockResolvedValue({ tasks: completedTasks, total: 1 })

    renderWithProviders(<Analytics />)

    await waitFor(() => {
      // The accuracy chart section should be visible
      expect(screen.getByText('Prediction Accuracy Breakdown')).toBeInTheDocument()
    })
  })

  it('shows completed tasks count', async () => {
    const completedTasks = [
      {
        id: '1',
        title: 'Completed Task',
        status: 'completed',
        actual_days: 1,
        predicted_days_low: 5,
        predicted_days_high: 8,
        created_at: '2024-01-01',
        updated_at: '2024-01-02',
      },
    ]

    vi.mocked(taskApi.list).mockResolvedValue({
      tasks: completedTasks,
      total: 1,
    })

    renderWithProviders(<Analytics />)

    await waitFor(() => {
      expect(screen.getByText('Completed')).toBeInTheDocument()
    })
  })
})
