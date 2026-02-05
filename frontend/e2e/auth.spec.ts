import { test, expect } from '@playwright/test'

test.describe('Authentication', () => {
  test('redirects to login when not authenticated', async ({ page }) => {
    await page.goto('/dashboard')
    await expect(page).toHaveURL(/.*login/)
  })

  test('login page renders correctly', async ({ page }) => {
    await page.goto('/login')
    await expect(page.getByRole('heading', { name: /login|sign in/i })).toBeVisible()
    await expect(page.getByPlaceholder(/email/i)).toBeVisible()
    await expect(page.getByPlaceholder(/password/i)).toBeVisible()
    await expect(page.getByRole('button', { name: /login|sign in/i })).toBeVisible()
  })

  test('shows error for invalid credentials', async ({ page }) => {
    await page.goto('/login')
    await page.getByPlaceholder(/email/i).fill('invalid@example.com')
    await page.getByPlaceholder(/password/i).fill('wrongpassword')
    await page.getByRole('button', { name: /login|sign in/i }).click()

    // Should show error message
    await expect(page.getByText(/invalid|error|failed/i)).toBeVisible()
  })

  test('successful login redirects to dashboard', async ({ page }) => {
    await page.goto('/login')

    // Use demo credentials
    await page.getByPlaceholder(/email/i).fill('demo@goplan.io')
    await page.getByPlaceholder(/password/i).fill('demo123')
    await page.getByRole('button', { name: /login|sign in/i }).click()

    // Should redirect to dashboard
    await expect(page).toHaveURL(/.*dashboard/)
  })

  test('logout clears session', async ({ page }) => {
    // First login
    await page.goto('/login')
    await page.getByPlaceholder(/email/i).fill('demo@goplan.io')
    await page.getByPlaceholder(/password/i).fill('demo123')
    await page.getByRole('button', { name: /login|sign in/i }).click()

    await expect(page).toHaveURL(/.*dashboard/)

    // Click logout
    await page.getByRole('button', { name: /logout/i }).click()

    // Should redirect to login
    await expect(page).toHaveURL(/.*login/)

    // Trying to access dashboard should redirect back to login
    await page.goto('/dashboard')
    await expect(page).toHaveURL(/.*login/)
  })
})
