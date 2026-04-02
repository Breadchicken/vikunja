import {test, expect} from '../../support/fixtures'
import {TaskFactory} from '../../factories/task'
import {ProjectFactory} from '../../factories/project'
import {TaskTimeEntryFactory} from '../../factories/task_time_entry'
import {createDefaultViews} from '../project/prepareProjects'

test.describe('Time Tracking', () => {
	test.beforeEach(async ({authenticatedPage: page}) => {
		const projects = await ProjectFactory.create(1)
		await createDefaultViews(projects[0].id)
		await TaskFactory.create(1, {
			project_id: projects[0].id,
		})
	})

	test('Shows time tracking section on task detail', async ({authenticatedPage: page}) => {
		await page.goto('/tasks/1')
		await expect(page.locator('.time-tracking-container')).toBeVisible()
		await expect(page.locator('.time-tracking-container h3')).toContainText('Time Tracking')
	})

	test('Can start and stop a timer', async ({authenticatedPage: page}) => {
		await page.goto('/tasks/1')
		await expect(page.locator('.time-tracking-container')).toBeVisible()

		// Start timer
		const startButton = page.locator('.timer-buttons .button.is-primary')
		await expect(startButton).toBeVisible()
		await expect(startButton).toContainText('Start Timer')

		const startPromise = page.waitForResponse(response =>
			response.url().includes('/timers/start') && response.status() === 201,
		)
		await startButton.click()
		await startPromise

		// Timer should be running - stop button visible
		const stopButton = page.locator('.timer-buttons .button.is-danger')
		await expect(stopButton).toBeVisible()
		await expect(stopButton).toContainText('Stop Timer')

		// Timer display should show time
		await expect(page.locator('.timer-time')).toBeVisible()

		// Stop timer
		const stopPromise = page.waitForResponse(response =>
			response.url().includes('/timers/stop') && response.status() === 200,
		)
		await stopButton.click()
		await stopPromise

		// Should show success notification
		await expect(page.locator('.global-notification')).toContainText('Success', {timeout: 4000})

		// Start button should be back
		await expect(page.locator('.timer-buttons .button.is-primary')).toBeVisible()

		// Time entry should appear in the list
		await expect(page.locator('.time-entries-list')).toBeVisible()
		await expect(page.locator('.time-entries-list tbody tr')).toHaveCount(1)
	})

	test('Can add a manual time entry', async ({authenticatedPage: page}) => {
		await page.goto('/tasks/1')
		await expect(page.locator('.time-tracking-container')).toBeVisible()

		// Open manual entry form
		await page.locator('.time-tracking-container').getByText('Add time manually').click()

		// Fill in the form
		await page.locator('.manual-entry input[type="number"]').first().fill('1') // hours
		await page.locator('.manual-entry input[type="number"]').nth(1).fill('30') // minutes
		await page.locator('.manual-entry input[type="text"]').fill('Worked on feature')

		// Save
		const createPromise = page.waitForResponse(response =>
			response.url().includes('/time-entries') && response.request().method() === 'PUT' && response.status() === 201,
		)
		await page.locator('.manual-entry .button.is-primary').click()
		await createPromise

		// Should show success notification
		await expect(page.locator('.global-notification')).toContainText('Success', {timeout: 4000})

		// Time entry should appear
		await expect(page.locator('.time-entries-list tbody tr')).toHaveCount(1)
	})

	test('Can delete a time entry', async ({authenticatedPage: page}) => {
		// Seed a time entry
		await TaskTimeEntryFactory.create(1, {
			task_id: 1,
			user_id: 1,
		})

		await page.goto('/tasks/1')
		await expect(page.locator('.time-entries-list')).toBeVisible()
		await expect(page.locator('.time-entries-list tbody tr')).toHaveCount(1)

		// Delete entry
		const deletePromise = page.waitForResponse(response =>
			response.url().includes('/time-entries/') && response.request().method() === 'DELETE',
		)
		await page.locator('.time-entries-list tbody tr .is-danger').first().click()
		await deletePromise

		await expect(page.locator('.global-notification')).toContainText('Success', {timeout: 4000})
	})

	test('Shows existing time entries with billable info', async ({authenticatedPage: page}) => {
		// Seed time entries
		await TaskTimeEntryFactory.create(1, {
			task_id: 1,
			user_id: 1,
			duration: 7200,
			billable: true,
		})

		await page.goto('/tasks/1')
		await expect(page.locator('.time-entries-list')).toBeVisible()
		await expect(page.locator('.time-entries-list tbody tr')).toHaveCount(1)

		// Check total is displayed
		await expect(page.locator('.time-entries-list tfoot')).toContainText('Total')
	})

	test('Timer persists across page navigation', async ({authenticatedPage: page}) => {
		await page.goto('/tasks/1')
		await expect(page.locator('.time-tracking-container')).toBeVisible()

		// Start timer
		const startPromise = page.waitForResponse(response =>
			response.url().includes('/timers/start') && response.status() === 201,
		)
		await page.locator('.timer-buttons .button.is-primary').click()
		await startPromise

		// Stop button should be visible
		await expect(page.locator('.timer-buttons .button.is-danger')).toBeVisible()

		// Navigate away and back
		await page.goto('/')
		await page.goto('/tasks/1')

		// Timer should still be running (stop button visible)
		await expect(page.locator('.timer-buttons .button.is-danger')).toBeVisible({timeout: 5000})
	})
})

test.describe('Time Report', () => {
	test.beforeEach(async ({authenticatedPage: page}) => {
		const projects = await ProjectFactory.create(1)
		await createDefaultViews(projects[0].id)
		await TaskFactory.create(1, {
			project_id: projects[0].id,
		})
	})

	test('Can navigate to time report from sidebar', async ({authenticatedPage: page}) => {
		await page.goto('/')
		await page.locator('.menu-container').getByText('Time Report').click()
		await expect(page).toHaveURL(/\/time-report/)
		await expect(page.locator('.time-report h1')).toContainText('Time Report')
	})

	test('Shows empty state when no entries exist', async ({authenticatedPage: page}) => {
		await page.goto('/time-report')
		await expect(page.locator('.time-report')).toContainText('No time entries found')
	})

	test('Shows time entries in report', async ({authenticatedPage: page}) => {
		await TaskTimeEntryFactory.create(1, {
			task_id: 1,
			user_id: 1,
			duration: 3600,
			billable: true,
		})

		await page.goto('/time-report')

		// Summary cards should show data
		await expect(page.locator('.summary-cards')).toBeVisible()

		// Table should have entries
		await expect(page.locator('.time-report table tbody tr')).toHaveCount(1)
	})

	test('Can filter by billable status', async ({authenticatedPage: page}) => {
		await TaskTimeEntryFactory.create(1, {
			task_id: 1,
			user_id: 1,
			duration: 3600,
			billable: true,
		})

		await page.goto('/time-report')

		// Select "Non-billable only" filter
		await page.locator('select').last().selectOption('false')
		await page.locator('.button.is-primary').filter({hasText: 'Generate Report'}).click()

		// Should show no entries (our entry is billable)
		await expect(page.locator('.time-report')).toContainText('No time entries found')
	})

	test('Page loads without errors', async ({authenticatedPage: page}) => {
		const errors: string[] = []
		page.on('pageerror', error => errors.push(error.message))

		await page.goto('/time-report')
		await page.waitForLoadState('networkidle')

		// No JS errors should occur
		expect(errors).toHaveLength(0)
	})
})

test.describe('Task Detail - No JS Errors', () => {
	test.beforeEach(async ({authenticatedPage: page}) => {
		const projects = await ProjectFactory.create(1)
		await createDefaultViews(projects[0].id)
		await TaskFactory.create(1, {
			project_id: projects[0].id,
		})
	})

	test('Task detail page loads without JS errors', async ({authenticatedPage: page}) => {
		const errors: string[] = []
		page.on('pageerror', error => errors.push(error.message))

		await page.goto('/tasks/1')
		await page.waitForLoadState('networkidle')

		expect(errors).toHaveLength(0)
	})
})
