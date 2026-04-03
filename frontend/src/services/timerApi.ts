import {AuthenticatedHTTPFactory} from '@/helpers/fetcher'
import TimeEntryModel from '@/models/timeEntry'
import type {ITimeEntry} from '@/modelTypes/ITimeEntry'

export async function startTimer(taskId: number, billable = true, description = ''): Promise<ITimeEntry> {
	const http = AuthenticatedHTTPFactory()
	const response = await http.post(`/tasks/${taskId}/timers/start`, {
		billable,
		description,
	})
	return new TimeEntryModel(response.data)
}

export async function stopTimer(taskId: number, billable = true, description = ''): Promise<ITimeEntry> {
	const http = AuthenticatedHTTPFactory()
	const response = await http.post(`/tasks/${taskId}/timers/stop`, {
		billable,
		description,
	})
	return new TimeEntryModel(response.data)
}

export async function getActiveTimer(): Promise<ITimeEntry | null> {
	const http = AuthenticatedHTTPFactory()
	const response = await http.get('/users/timers/active')
	if (response.data === null) {
		return null
	}
	return new TimeEntryModel(response.data)
}
