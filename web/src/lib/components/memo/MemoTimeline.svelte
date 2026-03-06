<script lang="ts">
	import { memoStore } from '$lib/stores/memos.svelte';
	import MemoCard from './MemoCard.svelte';

	let sentinel: HTMLDivElement;

	$effect(() => {
		if (!sentinel) return;
		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0].isIntersecting) {
					memoStore.loadMore();
				}
			},
			{ rootMargin: '200px' }
		);
		observer.observe(sentinel);
		return () => observer.disconnect();
	});
</script>

<div>
	{#each memoStore.items as memo (memo.id)}
		<MemoCard {memo} />
	{/each}

	{#if memoStore.loading}
		<div class="p-4 text-center text-[var(--color-text-muted)]">Loading...</div>
	{/if}

	{#if !memoStore.loading && memoStore.items.length === 0}
		<div class="p-8 text-center text-[var(--color-text-muted)]">
			<p class="text-lg">No memos yet</p>
			<p class="mt-1 text-sm">Write your first memo above!</p>
		</div>
	{/if}

	<div bind:this={sentinel}></div>
</div>
