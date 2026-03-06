<script lang="ts">
	import { api } from '$lib/api/client';
	import type { Tag } from '$lib/api/types';
	import { onMount } from 'svelte';

	let tags = $state<Tag[]>([]);
	let loading = $state(true);

	onMount(async () => {
		try {
			const res = await api.getTags();
			tags = res.items;
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Tags - Memento Memo</title>
</svelte:head>

<div class="mx-auto max-w-2xl p-4">
	<h2 class="mb-4 text-xl font-bold">Tags</h2>

	{#if loading}
		<p class="text-center text-[var(--color-text-muted)]">Loading...</p>
	{:else if tags.length === 0}
		<p class="text-center text-[var(--color-text-muted)]">No tags yet</p>
	{:else}
		<div class="flex flex-wrap gap-2">
			{#each tags as tag}
				<a
					href="/tags/{tag.name}"
					class="flex items-center gap-2 rounded-lg border border-[var(--color-border)] px-3 py-2 transition-colors hover:bg-[var(--color-bg-secondary)]"
				>
					<span class="text-[var(--color-primary)]">#{tag.name}</span>
					<span class="text-xs text-[var(--color-text-muted)]">{tag.memo_count}</span>
				</a>
			{/each}
		</div>
	{/if}
</div>
