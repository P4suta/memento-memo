<script lang="ts">
	import { api } from '$lib/api/client';
	import type { HeatmapEntry } from '$lib/api/types';
	import { onMount } from 'svelte';

	let heatmap = $state<HeatmapEntry[]>([]);
	let year = $state(new Date().getFullYear());
	let loading = $state(true);

	onMount(() => loadHeatmap());

	async function loadHeatmap() {
		loading = true;
		try {
			const res = await api.getHeatmap(year);
			heatmap = res.items;
		} finally {
			loading = false;
		}
	}

	function getColor(count: number): string {
		if (count === 0) return 'var(--color-bg-secondary)';
		if (count <= 3) return '#0e4429';
		if (count <= 6) return '#006d32';
		if (count <= 10) return '#26a641';
		return '#39d353';
	}

	// Generate all dates for the year
	function getDates() {
		const dates: { date: string; count: number }[] = [];
		const start = new Date(year, 0, 1);
		const end = new Date(year, 11, 31);
		const map = new Map(heatmap.map((e) => [e.date_label, e.memo_count]));

		for (let d = new Date(start); d <= end; d.setDate(d.getDate() + 1)) {
			const key = d.toISOString().slice(0, 10);
			dates.push({ date: key, count: map.get(key) || 0 });
		}
		return dates;
	}
</script>

<svelte:head>
	<title>Calendar - Memento Memo</title>
</svelte:head>

<div class="mx-auto max-w-4xl p-4">
	<div class="mb-4 flex items-center justify-between">
		<h2 class="text-xl font-bold">Calendar</h2>
		<div class="flex gap-2">
			<button onclick={() => { year--; loadHeatmap(); }} class="rounded px-2 py-1 text-sm hover:bg-[var(--color-bg-secondary)]">&lt;</button>
			<span class="text-sm">{year}</span>
			<button onclick={() => { year++; loadHeatmap(); }} class="rounded px-2 py-1 text-sm hover:bg-[var(--color-bg-secondary)]">&gt;</button>
		</div>
	</div>

	{#if loading}
		<p class="text-center text-[var(--color-text-muted)]">Loading...</p>
	{:else}
		<div class="flex flex-wrap gap-[3px]">
			{#each getDates() as { date, count }}
				<a
					href="/calendar?date={date}"
					class="h-3 w-3 rounded-sm"
					style="background-color: {getColor(count)}"
					title="{date}: {count} memos"
				></a>
			{/each}
		</div>
	{/if}
</div>
