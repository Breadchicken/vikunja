<template>
	<div class="content details time-tracking-container">
		<h3>
			<Icon icon="history" />
			{{ $t('task.timeTracking.title') }}
		</h3>

		<!-- Timer controls -->
		<div class="timer-controls">
			<div class="timer-display">
				<span class="timer-time">{{ formattedElapsed }}</span>
			</div>
			<div class="timer-buttons">
				<BaseButton
					v-if="!isRunning"
					class="button is-primary"
					:loading="isStarting"
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

		<!-- Manual entry form -->
		<div
			v-if="showManualEntry"
			class="manual-entry"
		>
			<h4>{{ $t('task.timeTracking.manualEntry') }}</h4>
			<div class="field is-grouped">
				<div class="control">
					<label class="label">{{ $t('task.timeTracking.duration') }}</label>
					<input
						v-model="manualHours"
						type="number"
						min="0"
						class="input is-small"
						:placeholder="$t('task.timeTracking.hours')"
					>
				</div>
				<div class="control">
					<label class="label">&nbsp;</label>
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
					<label class="label">&nbsp;</label>
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
						<td>{{ formatDuration(entry.duration) }}</td>
						<td>
							<Icon
								:icon="entry.billable ? 'check' : 'times'"
								:class="entry.billable ? 'has-text-success' : 'has-text-grey'"
							/>
						</td>
						<td>{{ entry.description }}</td>
						<td
							v-if="canWrite"
							class="has-text-right"
						>
							<BaseButton
								class="is-danger is-small"
								@click="deleteEntry(entry)"
							>
								<Icon icon="trash" />
							</BaseButton>
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
import {ref, computed, onMounted, onUnmounted, watch} from 'vue'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'
import TimeEntryService from '@/services/timeEntry'
import TimeEntryModel from '@/models/timeEntry'
import {startTimer, stopTimer, getActiveTimer} from '@/services/timerApi'
import type {ITimeEntry} from '@/modelTypes/ITimeEntry'
import type {ITask} from '@/modelTypes/ITask'
import {error, success} from '@/message'
import {formatDateShort} from '@/helpers/time/formatDate'

const props = defineProps<{
	taskId: ITask['id']
	canWrite: boolean
}>()

const {t} = useI18n()

const timeEntryService = new TimeEntryService()
const timeEntries = ref<ITimeEntry[]>([])
const isRunning = ref(false)
const activeEntry = ref<ITimeEntry | null>(null)
const elapsedSeconds = ref(0)
const isStarting = ref(false)
const isStopping = ref(false)
const isSavingManual = ref(false)
const showManualEntry = ref(false)

// Manual entry fields
const manualHours = ref(0)
const manualMinutes = ref(0)
const manualBillable = ref(true)
const manualDescription = ref('')

let timerInterval: ReturnType<typeof setInterval> | null = null

const formattedElapsed = computed(() => {
	const h = Math.floor(elapsedSeconds.value / 3600)
	const m = Math.floor((elapsedSeconds.value % 3600) / 60)
	const s = elapsedSeconds.value % 60
	return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
})

const totalDuration = computed(() => {
	return timeEntries.value.reduce((sum, e) => sum + (e.duration || 0), 0)
})

const billableDuration = computed(() => {
	return timeEntries.value.filter(e => e.billable).reduce((sum, e) => sum + (e.duration || 0), 0)
})

function formatDuration(seconds: number): string {
	if (!seconds || seconds <= 0) return '0h 0m'
	const h = Math.floor(seconds / 3600)
	const m = Math.floor((seconds % 3600) / 60)
	return `${h}h ${m}m`
}

function formatDate(date: Date): string {
	if (!date) return ''
	return formatDateShort(date)
}

function startTicking() {
	if (timerInterval) clearInterval(timerInterval)
	timerInterval = setInterval(() => {
		if (activeEntry.value?.start) {
			elapsedSeconds.value = Math.floor((Date.now() - new Date(activeEntry.value.start).getTime()) / 1000)
		}
	}, 1000)
}

function stopTicking() {
	if (timerInterval) {
		clearInterval(timerInterval)
		timerInterval = null
	}
	elapsedSeconds.value = 0
}

async function loadEntries() {
	try {
		const entries = await timeEntryService.getAll({taskId: props.taskId})
		timeEntries.value = entries
	} catch (e) {
		error(e)
	}
}

async function checkActiveTimer() {
	try {
		const timer = await getActiveTimer()
		if (timer && timer.taskId === props.taskId) {
			activeEntry.value = timer
			isRunning.value = true
			startTicking()
		} else {
			isRunning.value = false
			activeEntry.value = null
			stopTicking()
		}
	} catch (e) {
		// Silently fail – user may not have an active timer
	}
}

async function start() {
	isStarting.value = true
	try {
		const entry = await startTimer(props.taskId)
		activeEntry.value = entry
		isRunning.value = true
		startTicking()
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
		await stopTimer(props.taskId)
		isRunning.value = false
		activeEntry.value = null
		stopTicking()
		await loadEntries()
		success({message: t('task.timeTracking.stopped')})
	} catch (e) {
		error(e)
	} finally {
		isStopping.value = false
	}
}

async function saveManualEntry() {
	const totalSeconds = (manualHours.value * 3600) + (manualMinutes.value * 60)
	if (totalSeconds <= 0) {
		error({message: t('task.timeTracking.durationRequired')})
		return
	}

	isSavingManual.value = true
	try {
		const now = new Date()
		const start = new Date(now.getTime() - totalSeconds * 1000)
		const entry = new TimeEntryModel({
			taskId: props.taskId,
			start,
			end: now,
			duration: totalSeconds,
			billable: manualBillable.value,
			description: manualDescription.value,
		})

		await timeEntryService.create(entry)
		await loadEntries()

		// Reset form
		manualHours.value = 0
		manualMinutes.value = 0
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
	checkActiveTimer()
})

onMounted(() => {
	loadEntries()
	checkActiveTimer()
})

onUnmounted(() => {
	stopTicking()
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
	margin-bottom: 1rem;
}

.timer-display {
	font-size: 1.5rem;
	font-weight: bold;
	font-variant-numeric: tabular-nums;
}

.timer-time {
	font-family: monospace;
}

.manual-entry {
	padding: 1rem;
	background: var(--grey-100);
	border-radius: 0.5rem;
	margin-top: 0.5rem;

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
</style>
