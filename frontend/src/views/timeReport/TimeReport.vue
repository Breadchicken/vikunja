<template>
	<div class="content time-report">
		<h1>{{ $t('timeReport.title') }}</h1>

		<!-- Filters -->
		<div class="filters card">
			<div class="card-content">
				<div class="columns">
					<div class="column">
						<label class="label">{{ $t('timeReport.project') }}</label>
						<div class="control">
							<select
								v-model="filters.projectId"
								class="input"
							>
								<option :value="0">
									{{ $t('timeReport.allProjects') }}
								</option>
								<option
									v-for="project in projects"
									:key="project.id"
									:value="project.id"
								>
									{{ project.title }}
								</option>
							</select>
						</div>
					</div>
					<div class="column">
						<label class="label">{{ $t('timeReport.from') }}</label>
						<div class="control">
							<input
								v-model="filters.from"
								type="date"
								class="input"
							>
						</div>
					</div>
					<div class="column">
						<label class="label">{{ $t('timeReport.to') }}</label>
						<div class="control">
							<input
								v-model="filters.to"
								type="date"
								class="input"
							>
						</div>
					</div>
					<div class="column">
						<label class="label">{{ $t('timeReport.billableFilter') }}</label>
						<div class="control">
							<select
								v-model="filters.billable"
								class="input"
							>
								<option value="">
									{{ $t('timeReport.all') }}
								</option>
								<option value="true">
									{{ $t('timeReport.billableOnly') }}
								</option>
								<option value="false">
									{{ $t('timeReport.nonBillableOnly') }}
								</option>
							</select>
						</div>
					</div>
					<div class="column is-narrow is-flex is-align-items-flex-end">
						<BaseButton
							class="button is-primary"
							:loading="isLoading"
							@click="loadReport"
						>
							<Icon icon="search" />
							{{ $t('timeReport.generate') }}
						</BaseButton>
					</div>
				</div>
			</div>
		</div>

		<!-- Summary cards -->
		<div
			v-if="report"
			class="columns summary-cards mt-4"
		>
			<div class="column">
				<div class="card">
					<div class="card-content has-text-centered">
						<p class="heading">
							{{ $t('timeReport.totalTime') }}
						</p>
						<p class="title">
							{{ formatDuration(report.total_duration) }}
						</p>
					</div>
				</div>
			</div>
			<div class="column">
				<div class="card">
					<div class="card-content has-text-centered">
						<p class="heading">
							{{ $t('timeReport.billableTime') }}
						</p>
						<p class="title has-text-success">
							{{ formatDuration(report.billable_duration) }}
						</p>
					</div>
				</div>
			</div>
			<div class="column">
				<div class="card">
					<div class="card-content has-text-centered">
						<p class="heading">
							{{ $t('timeReport.nonBillableTime') }}
						</p>
						<p class="title has-text-grey">
							{{ formatDuration(report.total_duration - report.billable_duration) }}
						</p>
					</div>
				</div>
			</div>
			<div class="column">
				<div class="card">
					<div class="card-content has-text-centered">
						<p class="heading">
							{{ $t('timeReport.totalEntries') }}
						</p>
						<p class="title">
							{{ report.total_entries }}
						</p>
					</div>
				</div>
			</div>
		</div>

		<!-- Export button -->
		<div
			v-if="report && report.entries.length > 0"
			class="mt-4 mb-4"
		>
			<BaseButton
				class="button is-outlined"
				@click="exportCSV"
			>
				<Icon icon="download" />
				{{ $t('timeReport.exportCSV') }}
			</BaseButton>
		</div>

		<!-- Entries table -->
		<div
			v-if="report && report.entries.length > 0"
			class="mt-4"
		>
			<table class="table is-striped is-fullwidth is-hoverable">
				<thead>
					<tr>
						<th>{{ $t('timeReport.taskTitle') }}</th>
						<th>{{ $t('timeReport.projectName') }}</th>
						<th>{{ $t('timeReport.user') }}</th>
						<th>{{ $t('timeReport.date') }}</th>
						<th>{{ $t('timeReport.duration') }}</th>
						<th>{{ $t('timeReport.billable') }}</th>
						<th>{{ $t('timeReport.description') }}</th>
					</tr>
				</thead>
				<tbody>
					<tr
						v-for="entry in report.entries"
						:key="entry.time_entry_id"
					>
						<td>
							<RouterLink :to="{ name: 'task.detail', params: { id: entry.task_id } }">
								{{ entry.task_title }}
							</RouterLink>
						</td>
						<td>{{ entry.project_name }}</td>
						<td>{{ entry.username }}</td>
						<td>{{ formatDate(entry.start) }}</td>
						<td>{{ formatDuration(entry.duration) }}</td>
						<td>
							<Icon
								:icon="entry.billable ? 'check' : 'times'"
								:class="entry.billable ? 'has-text-success' : 'has-text-grey'"
							/>
						</td>
						<td>{{ entry.description }}</td>
					</tr>
				</tbody>
			</table>
		</div>

		<!-- Empty state -->
		<div
			v-if="report && report.entries.length === 0"
			class="has-text-centered mt-6"
		>
			<p class="has-text-grey">
				{{ $t('timeReport.noEntries') }}
			</p>
		</div>
	</div>
</template>

<script setup lang="ts">
import {ref, reactive, onMounted} from 'vue'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'
import {AuthenticatedHTTPFactory} from '@/helpers/fetcher'
import {formatDateShort} from '@/helpers/time/formatDate'
import {error} from '@/message'
import {useProjectStore} from '@/stores/projects'

const {t} = useI18n()
const projectStore = useProjectStore()

interface TimeReportEntryData {
	time_entry_id: number
	task_id: number
	task_title: string
	project_id: number
	project_name: string
	user_id: number
	username: string
	start: string
	end: string
	duration: number
	billable: boolean
	description: string
}

interface TimeReportData {
	entries: TimeReportEntryData[]
	total_duration: number
	billable_duration: number
	total_entries: number
}

const filters = reactive({
	projectId: 0,
	from: '',
	to: '',
	billable: '',
})

const report = ref<TimeReportData | null>(null)
const isLoading = ref(false)
const projects = ref<{id: number, title: string}[]>([])

function formatDuration(seconds: number): string {
	if (!seconds || seconds <= 0) return '0h 0m'
	const h = Math.floor(seconds / 3600)
	const m = Math.floor((seconds % 3600) / 60)
	return `${h}h ${m}m`
}

function formatDate(dateStr: string): string {
	if (!dateStr) return ''
	return formatDateShort(new Date(dateStr))
}

function buildQueryParams(): URLSearchParams {
	const params = new URLSearchParams()
	if (filters.projectId > 0) {
		params.set('project_id', String(filters.projectId))
	}
	if (filters.from) {
		params.set('from', new Date(filters.from).toISOString())
	}
	if (filters.to) {
		// Set to end of day
		const toDate = new Date(filters.to)
		toDate.setHours(23, 59, 59, 999)
		params.set('to', toDate.toISOString())
	}
	if (filters.billable) {
		params.set('billable', filters.billable)
	}
	return params
}

async function loadReport() {
	isLoading.value = true
	try {
		const params = buildQueryParams()
		const http = AuthenticatedHTTPFactory()
		const response = await http.get(`/time-report?${params.toString()}`)
		const data = response.data
		data.entries = data.entries || []
		report.value = data
	} catch (e) {
		error(e)
	} finally {
		isLoading.value = false
	}
}

async function exportCSV() {
	try {
		const params = buildQueryParams()
		const http = AuthenticatedHTTPFactory()
		const response = await http.get(`/time-report/csv?${params.toString()}`, {
			responseType: 'blob',
		})
		const blob = new Blob([response.data], {type: 'text/csv'})
		const url = window.URL.createObjectURL(blob)
		const a = document.createElement('a')
		a.href = url
		a.download = 'time-report.csv'
		a.click()
		window.URL.revokeObjectURL(url)
	} catch (e) {
		error(e)
	}
}

onMounted(async () => {
	// Load all projects for filter dropdown
	const allProjects = Object.values(projectStore.projects)
	projects.value = allProjects.map(p => ({id: p.id, title: p.title})).sort((a, b) => a.title.localeCompare(b.title))

	// Auto-load report on mount
	await loadReport()
})
</script>

<style scoped lang="scss">
.time-report {
	max-width: 1200px;
	margin: 0 auto;
	padding: 1rem;
}

.summary-cards .card {
	.heading {
		font-size: 0.85rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.title {
		font-size: 1.8rem;
		margin-top: 0.5rem;
	}
}

.filters .card-content {
	padding: 1rem;
}
</style>
