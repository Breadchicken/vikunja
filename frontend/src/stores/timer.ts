import {ref, computed} from 'vue'
import {defineStore, acceptHMRUpdate} from 'pinia'

import {getActiveTimer, startTimer as apiStartTimer, stopTimer as apiStopTimer} from '@/services/timerApi'
import TaskService from '@/services/task'
import TaskModel from '@/models/task'
import {useProjectStore} from '@/stores/projects'
import type {ITimeEntry} from '@/modelTypes/ITimeEntry'
import type {ITask} from '@/modelTypes/ITask'

export const useTimerStore = defineStore('timer', () => {
	const activeTimer = ref<ITimeEntry | null>(null)
	const taskTitle = ref('')
	const projectName = ref('')
	const taskId = ref<ITask['id'] | null>(null)
	const elapsedSeconds = ref(0)
	const isLoading = ref(false)

	let timerInterval: ReturnType<typeof setInterval> | null = null

	const isRunning = computed(() => {
		if (activeTimer.value === null) return false
		// Backend returns "0001-01-01T00:00:00Z" as Go's zero time for null end dates
		if (activeTimer.value.end === null) return true
		const endTime = new Date(activeTimer.value.end).getTime()
		// Year 0001 = Go zero value, meaning timer is still running
		return new Date(activeTimer.value.end).getFullYear() <= 1
	})

	const formattedElapsed = computed(() => {
		const h = Math.floor(elapsedSeconds.value / 3600)
		const m = Math.floor((elapsedSeconds.value % 3600) / 60)
		const s = elapsedSeconds.value % 60
		return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
	})

	function startTicking() {
		stopTicking()
		timerInterval = setInterval(() => {
			if (activeTimer.value?.start) {
				elapsedSeconds.value = Math.floor((Date.now() - new Date(activeTimer.value.start).getTime()) / 1000)
			}
		}, 1000)
		// Compute initial value immediately
		if (activeTimer.value?.start) {
			elapsedSeconds.value = Math.floor((Date.now() - new Date(activeTimer.value.start).getTime()) / 1000)
		}
	}

	function stopTicking() {
		if (timerInterval) {
			clearInterval(timerInterval)
			timerInterval = null
		}
		elapsedSeconds.value = 0
	}

	async function resolveTaskInfo(id: ITask['id']) {
		try {
			const taskService = new TaskService()
			const task = await taskService.get(new TaskModel({id}))
			taskTitle.value = task.title
			taskId.value = task.id

			const projectStore = useProjectStore()
			const project = projectStore.projects[task.projectId]
			projectName.value = project?.title || ''
		} catch {
			taskTitle.value = ''
			projectName.value = ''
		}
	}

	async function fetchActiveTimer() {
		isLoading.value = true
		try {
			const timer = await getActiveTimer()
			if (timer && (timer.end === null || new Date(timer.end).getFullYear() <= 1)) {
				activeTimer.value = timer
				taskId.value = timer.taskId
				startTicking()
				await resolveTaskInfo(timer.taskId)
			} else {
				clearState()
			}
		} catch {
			clearState()
		} finally {
			isLoading.value = false
		}
	}

	async function startTimer(timerTaskId: ITask['id'], billable: boolean, description: string): Promise<ITimeEntry> {
		const entry = await apiStartTimer(timerTaskId, billable, description)
		activeTimer.value = entry
		taskId.value = timerTaskId
		startTicking()
		// Resolve task info in background
		resolveTaskInfo(timerTaskId)
		return entry
	}

	async function stopTimer(timerTaskId: ITask['id'], billable: boolean, description: string): Promise<ITimeEntry> {
		const entry = await apiStopTimer(timerTaskId, billable, description)
		clearState()
		return entry
	}

	function clearState() {
		activeTimer.value = null
		taskId.value = null
		taskTitle.value = ''
		projectName.value = ''
		stopTicking()
	}

	return {
		activeTimer,
		taskTitle,
		projectName,
		taskId,
		elapsedSeconds,
		isLoading,

		isRunning,
		formattedElapsed,

		fetchActiveTimer,
		startTimer,
		stopTimer,
		clearState,
	}
})

if (import.meta.hot) {
	import.meta.hot.accept(acceptHMRUpdate(useTimerStore, import.meta.hot))
}
