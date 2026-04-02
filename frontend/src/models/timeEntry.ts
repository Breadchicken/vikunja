import AbstractModel from './abstractModel'
import UserModel from './user'

import type {ITimeEntry} from '@/modelTypes/ITimeEntry'
import type {IUser} from '@/modelTypes/IUser'

export default class TimeEntryModel extends AbstractModel<ITimeEntry> implements ITimeEntry {
	id = 0
	taskId = 0
	user: IUser = UserModel
	start: Date = null
	end: Date | null = null
	duration = 0
	billable = true
	description = ''

	created: Date = null
	updated: Date = null

	constructor(data: Partial<ITimeEntry> = {}) {
		super()
		this.assignData(data)

		this.user = new UserModel(this.user)
		this.start = this.start ? new Date(this.start) : null
		this.end = this.end ? new Date(this.end) : null
		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}
