<script lang="ts">
	import { searchStore } from '$lib/stores/search.svelte';

	function handleInput(e: Event) {
		const target = e.target as HTMLInputElement;
		searchStore.setQuery(target.value);
	}
</script>

<div class="relative">
	<input
		type="text"
		value={searchStore.query}
		oninput={handleInput}
		placeholder="Search memos..."
		class="w-full rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-4 py-2 pl-10 text-sm text-[var(--color-text)] placeholder-[var(--color-text-muted)] focus:border-[var(--color-primary)] focus:outline-none"
	/>
	<svg xmlns="http://www.w3.org/2000/svg" class="absolute left-3 top-2.5 h-4 w-4 text-[var(--color-text-muted)]" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
		<path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
	</svg>
</div>
