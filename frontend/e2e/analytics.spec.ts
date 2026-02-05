import { test, expect, Page } from '@playwright/test'

// Helper to login before tests
async function login(page: Page) {
  await page.goto('/login')
  await page.getByPlaceholder(/email/i).fill('demo@goplan.io')
  await page.getByPlaceholder(/password/i).fill('demo123')
  await page.getByRole('button', { name: /login|sign in/i }).click()
  await expect(page).toHaveURL(/.*dashboard/)
}

test.describe('Analytics Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
  })

  test('can navigate to analytics page', async ({ page }) => {
    await page.getByRole('link', { name: /analytics/i }).click()
    await expect(page).toHaveURL(/.*analytics/)
  })

  test('analytics page renders dashboard title', async ({ page }) => {
    await page.goto('/analytics')
    await expect(page.getByRole('heading', { name: /analytics.*dashboard/i })).toBeVisible()
  })

  test('displays stat cards', async ({ page }) => {
    await page.goto('/analytics')

    // Should show key metrics
    await expect(page.getByText(/total tasks/i)).toBeVisible()
    await expect(page.getByText(/completed/i)).toBeVisible()
    await expect(page.getByText(/prediction accuracy/i)).toBeVisible()
  })

  test('displays prediction accuracy chart', async ({ page }) => {
    await page.goto('/analytics')

    await expect(page.getByText(/prediction accuracy breakdown/i)).toBeVisible()
  })

  test('displays tasks by status chart', async ({ page }) => {
    await page.goto('/analytics')

    await expect(page.getByText(/tasks by status/i)).toBeVisible()
  })

  test('displays recent completions section', async ({ page }) => {
    await page.goto('/analytics')

    await expect(page.getByText(/recent completions/i)).toBeVisible()
  })

  test('displays key insights section', async ({ page }) => {
    await page.goto('/analytics')

    await expect(page.getByText(/key insights/i)).toBeVisible()
  })

  test('recent completions links to task details', async ({ page }) => {
    await page.goto('/analytics')

    // Wait for data to load
    await page.waitForLoadState('networkidle')

    // Check if there are task links in recent completions
    const taskLinks = page.locator('a[href*="/tasks/"]')
    const count = await taskLinks.count()

    if (count > 0) {
      // Click the first task link
      await taskLinks.first().click()

      // Should navigate to task detail
      await expect(page).toHaveURL(/.*tasks\/[a-f0-9-]+/)
    }
  })

  test('shows loading state before data loads', async ({ page }) => {
    // Slow down network to see loading state
    await page.route('**/api/**', async (route) => {
      await new Promise((resolve) => setTimeout(resolve, 1000))
      await route.continue()
    })

    await page.goto('/analytics')

    // Should show loading indicator
    await expect(page.getByText(/loading/i)).toBeVisible()
  })

  test('handles empty data gracefully', async ({ page }) => {
    // Mock empty response
    await page.route('**/api/v1/tasks*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ tasks: [], total: 0 }),
      })
    })

    await page.goto('/analytics')

    // Should show appropriate message for no data
    await expect(page.getByText(/no.*tasks|no data/i)).toBeVisible()
  })
})

test.describe('Analytics Responsiveness', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
  })

  test('stat cards stack on mobile', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/analytics')

    // Stat cards should be visible and properly stacked
    const statCards = page.locator('[class*="shadow"]')
    await expect(statCards.first()).toBeVisible()
  })

  test('charts are visible on tablet', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 })
    await page.goto('/analytics')

    await expect(page.getByText(/prediction accuracy breakdown/i)).toBeVisible()
    await expect(page.getByText(/tasks by status/i)).toBeVisible()
  })

  test('full layout on desktop', async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 })
    await page.goto('/analytics')

    // All sections should be visible
    await expect(page.getByRole('heading', { name: /analytics.*dashboard/i })).toBeVisible()
    await expect(page.getByText(/total tasks/i)).toBeVisible()
    await expect(page.getByText(/key insights/i)).toBeVisible()
  })
})
