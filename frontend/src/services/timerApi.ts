import {AuthenticatedHTTPFactory} from '@/helpers/fetcher'
import TimeEntryModel from '@/models/timeEntry'
import type {ITimeEntry} from '@/modelTypes/ITimeEntry'

const http = AuthenticatedHTTPFactory()

export async function startTimer(taskId: number): Promise<ITimeEntry> {
	const response = await http.post(`/tasks/${taskId}/timers/start`)
	return new TimeEntryModel(response.data)
}

export async function stopTimer(taskId: number): Promise<ITimeEntry> {
	const response = await http.post(`/tasks/${taskId}/timers/stop`)
	return new TimeEntryModel(response.data)
}

export async function getActiveTimer(): Promise<ITimeEntry | null> {
	const response = await http.get('/users/timers/active')
	if (response.data === null) {
		return null
	}
	return new TimeEntryModel(response.data)
}
