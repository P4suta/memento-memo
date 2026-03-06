<script lang="ts">
	import { api } from '$lib/api/client';
	import type { Stats } from '$lib/api/types';
	import { onMount } from 'svelte';

	let stats = $state<Stats | null>(null);
	let loading = $state(true);

	onMount(async () => {
		try {
			stats = await api.getStats();
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Stats - Memento Memo</title>
</svelte:head>

<div class="mx-auto max-w-2xl p-4">
	<h2 class="mb-6 text-xl font-bold">Statistics</h2>

	{#if loading}
		<p class="text-center text-[var(--color-text-muted)]">Loading...</p>
	{:else if stats}
		<div class="grid grid-cols-2 gap-4">
			<div class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-4">
				<p class="text-2xl font-bold text-[var(--color-primary)]">{stats.total_memos}</p>
				<p class="text-sm text-[var(--color-text-secondary)]">Total Memos</p>
			</div>
			<div class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-4">
				<p class="text-2xl font-bold text-[var(--color-accent)]">{stats.total_sessions}</p>
				<p class="text-sm text-[var(--color-text-secondary)]">Sessions</p>
			</div>
			<div class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-4">
				<p class="text-2xl font-bold text-yellow-400">{stats.active_days}</p>
				<p class="text-sm text-[var(--color-text-secondary)]">Active Days</p>
			</div>
			<div class="rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-4">
				<p class="text-2xl font-bold text-purple-400">{stats.total_chars.toLocaleString()}</p>
				<p class="text-sm text-[var(--color-text-secondary)]">Total Characters</p>
			</div>
		</div>
	{/if}
</div>
