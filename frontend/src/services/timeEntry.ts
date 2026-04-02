import AbstractService from './abstractService'
import TimeEntryModel from '@/models/timeEntry'
import type {ITimeEntry} from '@/modelTypes/ITimeEntry'

export default class TimeEntryService extends AbstractService<ITimeEntry> {
	constructor() {
		super({
			create: '/tasks/{taskId}/time-entries',
			getAll: '/tasks/{taskId}/time-entries',
			get: '/tasks/{taskId}/time-entries/{id}',
			update: '/tasks/{taskId}/time-entries/{id}',
			delete: '/tasks/{taskId}/time-entries/{id}',
		})
	}

	modelFactory(data) {
		return new TimeEntryModel(data)
	}
}
