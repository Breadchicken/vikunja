<template>
	<div class="content details time-tracking-container">
		<h3>
			<Icon icon="stopwatch" />
			{{ $t('task.timeTracking.title') }}
		</h3>

		<!-- Timer controls -->
		<div class="timer-controls">
			<div class="timer-display">
				<span class="timer-time">{{ isThisTaskRunning ? timerStore.formattedElapsed : '00:00:00' }}</span>
			</div>
			<div class="timer-buttons">
				<BaseButton
					v-if="!isThisTaskRunning"
					class="button is-primary"
					:loading="isStarting"
					:disabled="timerStore.isRunning && !isThisTaskRunning"
					@click="start"
				>
					<Icon icon="play" />
					{{ $t('task.timeTracking.start') }}
				</BaseButton>
				<BaseButton
					v-else
					class="button is-danger"
					:loading="isStopping"
					@click="stop"
				>
					<Icon icon="stop" />
					{{ $t('task.timeTracking.stop') }}
				</BaseButton>
			</div>
		</div>

		<!-- Timer options (shown when timer is running or before start) -->
		<div class="timer-options">
			<div class="field is-grouped">
				<div class="control">
					<label class="checkbox">
						<input
							v-model="timerBillable"
							type="checkbox"
						>
						{{ $t('task.timeTracking.billable') }}
					</label>
				</div>
				<div class="control is-expanded">
					<input
						v-model="timerDescription"
						type="text"
						class="input is-small"
						:placeholder="$t('task.timeTracking.descriptionPlaceholder')"
						@keyup.enter="!isRunning ? start() : undefined"
					>
				</div>
			</div>
		</div>

		<!-- Manual entry form -->
		<div
			v-if="showManualEntry"
			class="manual-entry mt-3"
		>
			<h4>{{ $t('task.timeTracking.manualEntry') }}</h4>
			<div class="field is-grouped">
				<div class="control">
					<label class="label">{{ $t('task.timeTracking.hours') }}</label>
					<input
						v-model="manualHours"
						type="number"
						min="0"
						class="input is-small"
						:placeholder="$t('task.timeTracking.hours')"
					>
				</div>
				<div class="control">
					<label class="label">{{ $t('task.timeTracking.minutes') }}</label>
					<input
						v-model="manualMinutes"
						type="number"
						min="0"
						max="59"
						class="input is-small"
						:placeholder="$t('task.timeTracking.minutes')"
					>
				</div>
				<div class="control">
					<label class="label">{{ $t('task.timeTracking.seconds') }}</label>
					<input
						v-model="manualSeconds"
						type="number"
						min="0"
						max="59"
						class="input is-small"
						:placeholder="$t('task.timeTracking.seconds')"
					>
				</div>
			</div>
			<div class="field is-grouped">
				<div class="control">
					<label class="checkbox">
						<input
							v-model="manualBillable"
							type="checkbox"
						>
						{{ $t('task.timeTracking.billable') }}
					</label>
				</div>
			</div>
			<div class="field">
				<label class="label">{{ $t('task.timeTracking.description') }}</label>
				<input
					v-model="manualDescription"
					type="text"
					class="input is-small"
					:placeholder="$t('task.timeTracking.descriptionPlaceholder')"
				>
			</div>
			<BaseButton
				class="button is-small is-primary"
				:loading="isSavingManual"
				@click="saveManualEntry"
			>
				{{ $t('task.timeTracking.save') }}
			</BaseButton>
		</div>

		<BaseButton
			class="has-text-primary is-small mt-2"
			@click="showManualEntry = !showManualEntry"
		>
			{{ showManualEntry ? $t('task.timeTracking.hideManualEntry') : $t('task.timeTracking.addManualEntry') }}
		</BaseButton>

		<!-- Time entries list -->
		<div
			v-if="timeEntries.length > 0"
			class="time-entries-list mt-4"
		>
			<h4>{{ $t('task.timeTracking.entries') }}</h4>
			<table class="table is-striped is-fullwidth is-hoverable">
				<thead>
					<tr>
						<th>{{ $t('task.timeTracking.user') }}</th>
						<th>{{ $t('task.timeTracking.date') }}</th>
						<th>{{ $t('task.timeTracking.durationLabel') }}</th>
						<th>{{ $t('task.timeTracking.billable') }}</th>
						<th>{{ $t('task.timeTracking.description') }}</th>
						<th
							v-if="canWrite"
							class="has-text-right"
						>
							{{ $t('task.timeTracking.actions') }}
						</th>
					</tr>
				</thead>
				<tbody>
					<tr
						v-for="entry in timeEntries"
						:key="entry.id"
					>
						<td>{{ entry.user?.name || entry.user?.username }}</td>
						<td>{{ formatDate(entry.start) }}</td>
						<td>
							<!-- Inline edit duration -->
							<template v-if="editingEntryId === entry.id">
								<div class="field is-grouped is-grouped-multiline">
									<input
										v-model.number="editHours"
										type="number"
										min="0"
										class="input is-small edit-duration-input"
									>
									<span>h</span>
									<input
										v-model.number="editMinutes"
										type="number"
										min="0"
										max="59"
										class="input is-small edit-duration-input"
									>
									<span>m</span>
									<input
										v-model.number="editSeconds"
										type="number"
										min="0"
										max="59"
										class="input is-small edit-duration-input"
									>
									<span>s</span>
								</div>
							</template>
							<template v-else>
								{{ formatDuration(entry.duration) }}
							</template>
						</td>
						<td>
							<!-- Inline edit billable -->
							<template v-if="editingEntryId === entry.id">
								<label class="checkbox">
									<input
										v-model="editBillable"
										type="checkbox"
									>
								</label>
							</template>
							<template v-else>
								<Icon
									:icon="entry.billable ? 'check' : 'times'"
									:class="entry.billable ? 'has-text-success' : 'has-text-grey'"
								/>
							</template>
						</td>
						<td>
							<!-- Inline edit description -->
							<template v-if="editingEntryId === entry.id">
								<input
									v-model="editDescription"
									type="text"
									class="input is-small"
								>
							</template>
							<template v-else>
								{{ entry.description }}
							</template>
						</td>
						<td
							v-if="canWrite"
							class="has-text-right actions-cell"
						>
							<template v-if="editingEntryId === entry.id">
								<BaseButton
									class="is-primary is-small mr-1"
									:loading="isSavingEdit"
									@click="saveEdit(entry)"
								>
									<Icon icon="check" />
								</BaseButton>
								<BaseButton
									class="is-small"
									@click="cancelEdit"
								>
									<Icon icon="times" />
								</BaseButton>
							</template>
							<template v-else>
								<BaseButton
									class="is-small mr-1"
									@click="startEdit(entry)"
								>
									<Icon icon="pen" />
								</BaseButton>
								<BaseButton
									class="is-danger is-small"
									@click="deleteEntry(entry)"
								>
									<Icon icon="trash" />
								</BaseButton>
							</template>
						</td>
					</tr>
				</tbody>
				<tfoot>
					<tr>
						<td colspan="2">
							<strong>{{ $t('task.timeTracking.total') }}</strong>
						</td>
						<td>
							<strong>{{ formatDuration(totalDuration) }}</strong>
						</td>
						<td>
							<strong>{{ formatDuration(billableDuration) }}</strong>
							<small class="has-text-grey">({{ $t('task.timeTracking.billable') }})</small>
						</td>
						<td colspan="2" />
					</tr>
				</tfoot>
			</table>
		</div>
	</div>
</template>

<script setup lang="ts">
import {ref, computed, onMounted, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'
import TimeEntryService from '@/services/timeEntry'
import TimeEntryModel from '@/models/timeEntry'
import type {ITimeEntry} from '@/modelTypes/ITimeEntry'
import type {ITask} from '@/modelTypes/ITask'
import {error, success} from '@/message'
import {formatDateShort} from '@/helpers/time/formatDate'
import {useTimerStore} from '@/stores/timer'

const props = defineProps<{
	taskId: ITask['id']
	canWrite: boolean
}>()

const {t} = useI18n()
const timerStore = useTimerStore()

const timeEntryService = new TimeEntryService()
const timeEntries = ref<ITimeEntry[]>([])
const isStarting = ref(false)
const isStopping = ref(false)
const isSavingManual = ref(false)
const showManualEntry = ref(false)

// Timer options (billable + description available when starting)
const timerBillable = ref(true)
const timerDescription = ref('')

// Manual entry fields
const manualHours = ref(0)
const manualMinutes = ref(0)
const manualSeconds = ref(0)
const manualBillable = ref(true)
const manualDescription = ref('')

// Inline edit state
const editingEntryId = ref<number | null>(null)
const editHours = ref(0)
const editMinutes = ref(0)
const editSeconds = ref(0)
const editBillable = ref(true)
const editDescription = ref('')
const isSavingEdit = ref(false)

const isThisTaskRunning = computed(() =>
	timerStore.isRunning && timerStore.activeTimer?.taskId === props.taskId,
)

const totalDuration = computed(() => {
	return timeEntries.value.reduce((sum, e) => sum + (e.duration || 0), 0)
})

const billableDuration = computed(() => {
	return timeEntries.value.filter(e => e.billable).reduce((sum, e) => sum + (e.duration || 0), 0)
})

function formatDuration(seconds: number): string {
	if (!seconds || seconds <= 0) return '0h 0m 0s'
	const h = Math.floor(seconds / 3600)
	const m = Math.floor((seconds % 3600) / 60)
	const s = seconds % 60
	return `${h}h ${m}m ${s}s`
}

function formatDate(date: Date): string {
	if (!date) return ''
	return formatDateShort(date)
}

async function loadEntries() {
	if (!props.taskId) return
	try {
		const entries = await timeEntryService.getAll({taskId: props.taskId})
		timeEntries.value = entries || []
	} catch {
		// Silently fail on load
	}
}

function syncFromStore() {
	if (timerStore.isRunning && timerStore.activeTimer?.taskId === props.taskId) {
		timerBillable.value = timerStore.activeTimer.billable
		timerDescription.value = timerStore.activeTimer.description || ''
	}
}

async function start() {
	isStarting.value = true
	try {
		await timerStore.startTimer(props.taskId, timerBillable.value, timerDescription.value)
		success({message: t('task.timeTracking.started')})
	} catch (e) {
		error(e)
	} finally {
		isStarting.value = false
	}
}

async function stop() {
	isStopping.value = true
	try {
		await timerStore.stopTimer(props.taskId, timerBillable.value, timerDescription.value)
		timerDescription.value = ''
		timerBillable.value = true
		await loadEntries()
		success({message: t('task.timeTracking.stopped')})
	} catch (e) {
		error(e)
	} finally {
		isStopping.value = false
	}
}

async function saveManualEntry() {
	const totalSecs = (manualHours.value * 3600) + (manualMinutes.value * 60) + manualSeconds.value
	if (totalSecs <= 0) {
		error({message: t('task.timeTracking.durationRequired')})
		return
	}

	isSavingManual.value = true
	try {
		const now = new Date()
		const startDate = new Date(now.getTime() - totalSecs * 1000)
		const entry = new TimeEntryModel({
			taskId: props.taskId,
			start: startDate,
			end: now,
			duration: totalSecs,
			billable: manualBillable.value,
			description: manualDescription.value,
		})

		await timeEntryService.create(entry)
		await loadEntries()

		manualHours.value = 0
		manualMinutes.value = 0
		manualSeconds.value = 0
		manualBillable.value = true
		manualDescription.value = ''
		showManualEntry.value = false

		success({message: t('task.timeTracking.entrySaved')})
	} catch (e) {
		error(e)
	} finally {
		isSavingManual.value = false
	}
}

// Inline editing
function startEdit(entry: ITimeEntry) {
	editingEntryId.value = entry.id
	const dur = entry.duration || 0
	editHours.value = Math.floor(dur / 3600)
	editMinutes.value = Math.floor((dur % 3600) / 60)
	editSeconds.value = dur % 60
	editBillable.value = entry.billable
	editDescription.value = entry.description || ''
}

function cancelEdit() {
	editingEntryId.value = null
}

async function saveEdit(entry: ITimeEntry) {
	isSavingEdit.value = true
	try {
		const newDuration = (editHours.value * 3600) + (editMinutes.value * 60) + editSeconds.value
		const startDate = new Date(entry.start)
		const endDate = new Date(startDate.getTime() + newDuration * 1000)

		const updated = new TimeEntryModel({
			...entry,
			id: entry.id,
			taskId: props.taskId,
			start: startDate,
			end: endDate,
			duration: newDuration,
			billable: editBillable.value,
			description: editDescription.value,
		})

		await timeEntryService.update(updated)
		editingEntryId.value = null
		await loadEntries()
		success({message: t('task.timeTracking.entryUpdated')})
	} catch (e) {
		error(e)
	} finally {
		isSavingEdit.value = false
	}
}

async function deleteEntry(entry: ITimeEntry) {
	try {
		await timeEntryService.delete({id: entry.id, taskId: props.taskId})
		await loadEntries()
		success({message: t('task.timeTracking.entryDeleted')})
	} catch (e) {
		error(e)
	}
}

watch(() => props.taskId, () => {
	loadEntries()
	syncFromStore()
})

onMounted(() => {
	loadEntries()
	syncFromStore()
})
</script>

<style scoped lang="scss">
.time-tracking-container {
	margin-top: 1rem;
}

.timer-controls {
	display: flex;
	align-items: center;
	gap: 1rem;
	margin-bottom: 0.5rem;
}

.timer-display {
	font-size: 1.5rem;
	font-weight: bold;
	font-variant-numeric: tabular-nums;
}

.timer-time {
	font-family: monospace;
}

.timer-options {
	margin-bottom: 1rem;

	.field.is-grouped {
		gap: 0.75rem;
		align-items: center;
	}
}

.manual-entry {
	padding: 1rem;
	background: var(--grey-100);
	border-radius: 0.5rem;

	.field.is-grouped {
		gap: 0.5rem;
		align-items: flex-end;
	}
}

.time-entries-list {
	table {
		font-size: 0.9rem;
	}
}

.edit-duration-input {
	width: 4rem !important;
}

.actions-cell {
	white-space: nowrap;
}
</style>
