<script lang="ts">
	import { api } from '$lib/api/client';
	import type { Memo } from '$lib/api/types';
	import MemoContent from '$lib/components/memo/MemoContent.svelte';
	import { onMount } from 'svelte';

	let memos = $state<Memo[]>([]);
	let loading = $state(true);

	onMount(async () => {
		loading = true;
		try {
			// Use the memos endpoint with deleted filter (handled server-side)
			const res = await fetch('/api/v1/memos?deleted=true&limit=50');
			if (res.ok) {
				const data = await res.json();
				memos = data.items || [];
			}
		} finally {
			loading = false;
		}
	});

	async function restore(id: number) {
		await api.restoreMemo(id);
		memos = memos.filter((m) => m.id !== id);
	}

	async function permanentDelete(id: number) {
		if (!confirm('Permanently delete this memo? This cannot be undone.')) return;
		await api.permanentDeleteMemo(id);
		memos = memos.filter((m) => m.id !== id);
	}
</script>

<svelte:head>
	<title>Trash - Memento Memo</title>
</svelte:head>

<div class="mx-auto max-w-2xl p-4">
	<h2 class="mb-4 text-xl font-bold">Trash</h2>

	{#if loading}
		<p class="text-center text-[var(--color-text-muted)]">Loading...</p>
	{:else if memos.length === 0}
		<p class="text-center text-[var(--color-text-muted)]">Trash is empty</p>
	{:else}
		{#each memos as memo (memo.id)}
			<div class="border-b border-[var(--color-border)] p-4">
				<MemoContent html={memo.content_html} />
				<div class="mt-2 flex gap-2">
					<button onclick={() => restore(memo.id)} class="rounded bg-[var(--color-accent)] px-3 py-1 text-xs text-white">Restore</button>
					<button onclick={() => permanentDelete(memo.id)} class="rounded bg-red-600 px-3 py-1 text-xs text-white">Delete forever</button>
				</div>
			</div>
		{/each}
	{/if}
</div>
