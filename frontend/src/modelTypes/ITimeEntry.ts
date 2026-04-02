import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'
import type {ITask} from './ITask'

export interface ITimeEntry extends IAbstract {
	id: number
	taskId: ITask['id']
	user: IUser
	start: Date
	end: Date | null
	duration: number
	billable: boolean
	description: string

	created: Date
	updated: Date
}
