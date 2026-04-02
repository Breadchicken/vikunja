import {Factory} from '../support/factory'

export class TaskTimeEntryFactory extends Factory {
	static table = 'task_time_entries'

	static factory() {
		const now = new Date()
		const start = new Date(now.getTime() - 3600 * 1000) // 1 hour ago

		return {
			id: '{increment}',
			task_id: 1,
			user_id: 1,
			start: start.toISOString(),
			end: now.toISOString(),
			duration: 3600,
			billable: true,
			description: '',
			created: now.toISOString(),
			updated: now.toISOString(),
		}
	}
}
