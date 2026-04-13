<template>
	<BaseButton
		v-if="timerStore.isRunning"
		:to="taskRoute"
		class="header-timer trigger-button"
		:aria-label="$t('task.timeTracking.headerTimer')"
	>
		<Icon
			icon="stopwatch"
			class="timer-icon"
		/>
		<span class="timer-time">{{ timerStore.formattedElapsed }}</span>
		<span
			v-if="timerStore.taskTitle"
			class="timer-task-title"
		>
			{{ timerStore.taskTitle }}
		</span>
	</BaseButton>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import BaseButton from '@/components/base/BaseButton.vue'
import {useTimerStore} from '@/stores/timer'

const timerStore = useTimerStore()

const taskRoute = computed(() => {
	if (!timerStore.taskId) return undefined
	return {name: 'task.detail', params: {id: timerStore.taskId}}
})
</script>

<style lang="scss" scoped>
.header-timer {
	display: inline-flex;
	align-items: center;
	gap: 0.4rem;
	padding-inline: 0.5rem;
	color: var(--primary);
	font-size: 0.85rem;
	text-decoration: none;
	white-space: nowrap;

	&:hover {
		color: var(--primary-dark, var(--primary));
		opacity: 0.8;
	}
}

.timer-icon {
	font-size: var(--navbar-icon-size);
	animation: pulse 2s ease-in-out infinite;
}

.timer-time {
	font-family: monospace;
	font-variant-numeric: tabular-nums;
	font-weight: 700;
}

.timer-task-title {
	max-inline-size: 150px;
	overflow: hidden;
	text-overflow: ellipsis;
	font-weight: 400;

	@media screen and (max-width: $tablet) {
		display: none;
	}
}

@keyframes pulse {
	0%, 100% {
		opacity: 1;
	}

	50% {
		opacity: 0.5;
	}
}
</style>
