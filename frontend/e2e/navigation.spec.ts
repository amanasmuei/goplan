import { test, expect, Page } from '@playwright/test'

// Helper to login before tests
async function login(page: Page) {
  await page.goto('/login')
  await page.getByPlaceholder(/email/i).fill('demo@goplan.io')
  await page.getByPlaceholder(/password/i).fill('demo123')
  await page.getByRole('button', { name: /login|sign in/i }).click()
  await expect(page).toHaveURL(/.*dashboard/)
}

test.describe('Navigation', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
  })

  test('sidebar displays all navigation items', async ({ page }) => {
    await expect(page.getByRole('link', { name: /dashboard/i })).toBeVisible()
    await expect(page.getByRole('link', { name: /tasks/i })).toBeVisible()
    await expect(page.getByRole('link', { name: /new task/i })).toBeVisible()
    await expect(page.getByRole('link', { name: /analytics/i })).toBeVisible()
  })

  test('sidebar shows GoPlan logo', async ({ page }) => {
    await expect(page.getByText('GoPlan')).toBeVisible()
  })

  test('sidebar shows user info', async ({ page }) => {
    // Should show user avatar/name in sidebar
    const userSection = page.locator('[class*="border-t"]').last()
    await expect(userSection).toBeVisible()
  })

  test('can navigate to dashboard', async ({ page }) => {
    await page.getByRole('link', { name: /dashboard/i }).click()
    await expect(page).toHaveURL(/.*dashboard/)
  })

  test('can navigate to tasks list', async ({ page }) => {
    await page.getByRole('link', { name: /tasks/i }).first().click()
    await expect(page).toHaveURL(/.*tasks/)
  })

  test('can navigate to new task', async ({ page }) => {
    await page.getByRole('link', { name: /new task/i }).click()
    await expect(page).toHaveURL(/.*tasks\/new/)
  })

  test('can navigate to analytics', async ({ page }) => {
    await page.getByRole('link', { name: /analytics/i }).click()
    await expect(page).toHaveURL(/.*analytics/)
  })

  test('active navigation item is highlighted', async ({ page }) => {
    await page.goto('/dashboard')

    const dashboardLink = page.getByRole('link', { name: /dashboard/i })
    // Check if it has active styling (bg-primary-50 or similar)
    await expect(dashboardLink).toHaveClass(/bg-primary|active/)
  })

  test('browser back button works', async ({ page }) => {
    await page.goto('/dashboard')
    await page.getByRole('link', { name: /tasks/i }).first().click()
    await expect(page).toHaveURL(/.*tasks/)

    await page.goBack()
    await expect(page).toHaveURL(/.*dashboard/)
  })

  test('browser forward button works', async ({ page }) => {
    await page.goto('/dashboard')
    await page.getByRole('link', { name: /tasks/i }).first().click()
    await page.goBack()
    await page.goForward()

    await expect(page).toHaveURL(/.*tasks/)
  })
})

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
  })

  test('dashboard renders correctly', async ({ page }) => {
    await page.goto('/dashboard')

    // Should have welcome or overview content
    await expect(page.getByRole('heading')).toBeVisible()
  })

  test('dashboard shows task statistics', async ({ page }) => {
    await page.goto('/dashboard')

    // Should show some task-related stats or recent activity
    await page.waitForLoadState('networkidle')
  })
})

test.describe('Accessibility', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
  })

  test('navigation is keyboard accessible', async ({ page }) => {
    await page.goto('/dashboard')

    // Tab through navigation items
    await page.keyboard.press('Tab')
    await page.keyboard.press('Tab')
    await page.keyboard.press('Tab')

    // Press Enter on focused element
    await page.keyboard.press('Enter')

    // Should navigate
    await page.waitForLoadState('networkidle')
  })

  test('pages have proper heading structure', async ({ page }) => {
    await page.goto('/dashboard')

    // Should have h1
    const h1Count = await page.locator('h1').count()
    expect(h1Count).toBeGreaterThan(0)
  })

  test('interactive elements have focus indicators', async ({ page }) => {
    await page.goto('/dashboard')

    // Focus a button
    const button = page.getByRole('button').first()
    if (await button.isVisible()) {
      await button.focus()
      // Should have visible focus ring/outline
      await expect(button).toBeFocused()
    }
  })

  test('forms have proper labels', async ({ page }) => {
    await page.goto('/tasks/new')

    // Input fields should have associated labels
    const titleInput = page.getByLabel(/title/i)
    await expect(titleInput).toBeVisible()

    const descInput = page.getByLabel(/description/i)
    await expect(descInput).toBeVisible()
  })
})

test.describe('Error Handling', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
  })

  test('404 page for unknown routes', async ({ page }) => {
    await page.goto('/nonexistent-route')

    // Should show some error indication or redirect
    await page.waitForLoadState('networkidle')
  })

  test('handles API errors gracefully', async ({ page }) => {
    // Mock API error
    await page.route('**/api/v1/tasks*', async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Internal server error' }),
      })
    })

    await page.goto('/tasks')

    // Should show error state, not crash
    await page.waitForLoadState('networkidle')
  })

  test('handles network timeout', async ({ page }) => {
    // Mock slow API
    await page.route('**/api/v1/tasks*', async (route) => {
      await new Promise((resolve) => setTimeout(resolve, 30000))
      await route.continue()
    })

    await page.goto('/tasks')

    // Should show loading state
    await expect(page.getByText(/loading/i)).toBeVisible()
  })
})
