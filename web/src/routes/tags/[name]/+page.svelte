<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import MemoCard from '$lib/components/memo/MemoCard.svelte';
	import type { Memo } from '$lib/api/types';
	import { onMount } from 'svelte';

	let memos = $state<Memo[]>([]);
	let cursor = $state<string | null>(null);
	let hasMore = $state(true);
	let loading = $state(true);

	$effect(() => {
		const name = $page.params.name;
		if (name) loadMemos(name);
	});

	async function loadMemos(name: string) {
		loading = true;
		try {
			const res = await api.getTagMemos(name, { limit: 20 });
			memos = res.items;
			cursor = res.next_cursor || null;
			hasMore = res.has_more;
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>#{$page.params.name} - Memento Memo</title>
</svelte:head>

<div class="mx-auto max-w-2xl p-4">
	<h2 class="mb-4 text-xl font-bold">
		<span class="text-[var(--color-primary)]">#{$page.params.name}</span>
	</h2>

	{#if loading}
		<p class="text-center text-[var(--color-text-muted)]">Loading...</p>
	{:else}
		{#each memos as memo (memo.id)}
			<MemoCard {memo} />
		{/each}
		{#if memos.length === 0}
			<p class="text-center text-[var(--color-text-muted)]">No memos with this tag</p>
		{/if}
	{/if}
</div>
