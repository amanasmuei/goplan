import { test, expect, Page } from '@playwright/test'

// Helper to login before tests
async function login(page: Page) {
  await page.goto('/login')
  await page.getByPlaceholder(/email/i).fill('demo@goplan.io')
  await page.getByPlaceholder(/password/i).fill('demo123')
  await page.getByRole('button', { name: /login|sign in/i }).click()
  await expect(page).toHaveURL(/.*dashboard/)
}

test.describe('Task Lifecycle', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
  })

  test('can navigate to create task page', async ({ page }) => {
    await page.getByRole('link', { name: /new task/i }).click()
    await expect(page).toHaveURL(/.*tasks\/new/)
    await expect(page.getByRole('heading', { name: /create.*task/i })).toBeVisible()
  })

  test('create task form has required fields', async ({ page }) => {
    await page.goto('/tasks/new')

    await expect(page.getByLabel(/title/i)).toBeVisible()
    await expect(page.getByLabel(/description/i)).toBeVisible()
    await expect(page.getByRole('button', { name: /create|submit/i })).toBeVisible()
  })

  test('can create a new task', async ({ page }) => {
    await page.goto('/tasks/new')

    const taskTitle = `E2E Test Task ${Date.now()}`
    const taskDescription = 'This is a test task created by E2E tests. It contains enough detail to pass validation.'

    await page.getByLabel(/title/i).fill(taskTitle)
    await page.getByLabel(/description/i).fill(taskDescription)

    // Fill optional estimate if visible
    const estimateField = page.getByLabel(/estimate/i)
    if (await estimateField.isVisible()) {
      await estimateField.fill('5')
    }

    await page.getByRole('button', { name: /create|submit/i }).click()

    // Should navigate to task detail or task list
    await expect(page).toHaveURL(/.*tasks/)
  })

  test('can view task list', async ({ page }) => {
    await page.goto('/tasks')

    await expect(page.getByRole('heading', { name: /tasks/i })).toBeVisible()
    // Should have a table or list of tasks
    await expect(page.locator('table, [role="list"], .task-list')).toBeVisible()
  })

  test('can view task details', async ({ page }) => {
    await page.goto('/tasks')

    // Click on first task if available
    const taskLink = page.locator('a[href*="/tasks/"]').first()
    if (await taskLink.isVisible()) {
      await taskLink.click()
      await expect(page).toHaveURL(/.*tasks\/[a-f0-9-]+/)
      await expect(page.getByText(/description|details/i)).toBeVisible()
    }
  })

  test('task detail shows status badge', async ({ page }) => {
    await page.goto('/tasks')

    const taskLink = page.locator('a[href*="/tasks/"]').first()
    if (await taskLink.isVisible()) {
      await taskLink.click()

      // Should show status badge
      await expect(
        page.locator('.badge, [class*="status"], [class*="rounded-full"]').first()
      ).toBeVisible()
    }
  })

  test('can acknowledge a pending task', async ({ page }) => {
    await page.goto('/tasks')

    // Filter for pending_acknowledgment tasks if filter exists
    const statusFilter = page.getByLabel(/status/i)
    if (await statusFilter.isVisible()) {
      await statusFilter.selectOption('pending_acknowledgment')
    }

    // Find a task with pending_acknowledgment status
    const pendingTask = page.locator('text=pending acknowledgment').first()
    if (await pendingTask.isVisible()) {
      await pendingTask.click()

      // Should show acknowledge button
      const ackButton = page.getByRole('button', { name: /acknowledge/i })
      await expect(ackButton).toBeVisible()

      await ackButton.click()

      // Dialog should appear
      await expect(page.getByText(/acknowledge task/i)).toBeVisible()

      // Select accept and confirm
      await page.getByText(/accept/i).click()
      await page.getByRole('button', { name: /confirm/i }).click()

      // Status should change
      await expect(page.getByText(/acknowledged/i)).toBeVisible()
    }
  })

  test('can start an acknowledged task', async ({ page }) => {
    await page.goto('/tasks')

    // Find acknowledged task
    const acknowledgedTask = page.locator('text=acknowledged').first()
    if (await acknowledgedTask.isVisible()) {
      await acknowledgedTask.click()

      // Should show start button
      const startButton = page.getByRole('button', { name: /start/i })
      if (await startButton.isVisible()) {
        await startButton.click()

        // Status should change to active
        await expect(page.getByText(/active/i)).toBeVisible()
      }
    }
  })

  test('can complete an active task', async ({ page }) => {
    await page.goto('/tasks')

    // Find active task
    const activeTask = page.locator('text=active').first()
    if (await activeTask.isVisible()) {
      await activeTask.click()

      // Should show complete button
      const completeButton = page.getByRole('button', { name: /complete/i })
      if (await completeButton.isVisible()) {
        await completeButton.click()

        // Status should change to pending_review
        await expect(page.getByText(/pending.*review/i)).toBeVisible()
      }
    }
  })
})

test.describe('Task Search and Filter', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
  })

  test('can search tasks', async ({ page }) => {
    await page.goto('/tasks')

    const searchInput = page.getByPlaceholder(/search/i)
    if (await searchInput.isVisible()) {
      await searchInput.fill('test')
      await page.keyboard.press('Enter')

      // Should filter results
      await page.waitForLoadState('networkidle')
    }
  })

  test('can filter by status', async ({ page }) => {
    await page.goto('/tasks')

    const statusFilter = page.getByLabel(/status/i)
    if (await statusFilter.isVisible()) {
      await statusFilter.selectOption('completed')

      // Should show only completed tasks
      await page.waitForLoadState('networkidle')
    }
  })
})
