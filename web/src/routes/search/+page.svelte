<script lang="ts">
	import SearchBar from '$lib/components/search/SearchBar.svelte';
	import MemoCard from '$lib/components/memo/MemoCard.svelte';
	import { searchStore } from '$lib/stores/search.svelte';
</script>

<svelte:head>
	<title>Search - Memento Memo</title>
</svelte:head>

<div class="mx-auto max-w-2xl p-4">
	<h2 class="mb-4 text-xl font-bold">Search</h2>
	<SearchBar />

	<div class="mt-4">
		{#if searchStore.loading && searchStore.results.length === 0}
			<p class="text-center text-[var(--color-text-muted)]">Searching...</p>
		{/if}

		{#each searchStore.results as memo (memo.id)}
			<MemoCard {memo} />
		{/each}

		{#if !searchStore.loading && searchStore.query && searchStore.results.length === 0}
			<p class="text-center text-[var(--color-text-muted)]">No results found</p>
		{/if}

		{#if searchStore.hasMore}
			<button
				onclick={() => searchStore.loadMore()}
				class="mt-4 w-full rounded-lg border border-[var(--color-border)] p-2 text-sm text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-secondary)]"
			>
				Load more
			</button>
		{/if}
	</div>
</div>
