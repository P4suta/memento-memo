<script lang="ts">
	import type { Memo } from '$lib/api/types';
	import { memoStore } from '$lib/stores/memos.svelte';
	import MemoContent from './MemoContent.svelte';

	let { memo }: { memo: Memo } = $props();

	let showActions = $state(false);
	let editing = $state(false);
	let editContent = $state('');

	function startEdit() {
		editContent = memo.content;
		editing = true;
		showActions = false;
	}

	async function saveEdit() {
		if (!editContent.trim()) return;
		await memoStore.update(memo.id, editContent);
		editing = false;
	}

	function cancelEdit() {
		editing = false;
	}

	function formatTime(dateStr: string) {
		const d = new Date(dateStr);
		return d.toLocaleString('ja-JP', {
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}
</script>

<article class="border-b border-[var(--color-border)] p-4 transition-colors hover:bg-[var(--color-bg-secondary)]">
	<div class="flex items-start justify-between">
		<time class="text-xs text-[var(--color-text-muted)]">{formatTime(memo.created_at)}</time>
		<div class="relative">
			<button
				onclick={() => (showActions = !showActions)}
				class="text-[var(--color-text-muted)] hover:text-[var(--color-text)] p-1"
			>
				...
			</button>
			{#if showActions}
				<div class="absolute right-0 top-6 z-10 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-secondary)] py-1 shadow-lg">
					<button onclick={startEdit} class="block w-full px-4 py-1.5 text-left text-sm hover:bg-[var(--color-bg-tertiary)]">Edit</button>
					<button onclick={() => memoStore.togglePin(memo.id)} class="block w-full px-4 py-1.5 text-left text-sm hover:bg-[var(--color-bg-tertiary)]">
						{memo.pinned ? 'Unpin' : 'Pin'}
					</button>
					<button onclick={() => memoStore.remove(memo.id)} class="block w-full px-4 py-1.5 text-left text-sm text-red-400 hover:bg-[var(--color-bg-tertiary)]">Delete</button>
				</div>
			{/if}
		</div>
	</div>

	{#if editing}
		<div class="mt-2">
			<textarea
				bind:value={editContent}
				class="w-full resize-none rounded-lg border border-[var(--color-border)] bg-[var(--color-bg)] p-2 text-sm text-[var(--color-text)]"
				rows="4"
			></textarea>
			<div class="mt-2 flex gap-2">
				<button onclick={saveEdit} class="rounded bg-[var(--color-primary)] px-3 py-1 text-xs text-white">Save</button>
				<button onclick={cancelEdit} class="rounded bg-[var(--color-bg-tertiary)] px-3 py-1 text-xs">Cancel</button>
			</div>
		</div>
	{:else}
		<div class="mt-2">
			<MemoContent html={memo.content_html} />
		</div>
	{/if}

	{#if memo.tags && memo.tags.length > 0}
		<div class="mt-2 flex flex-wrap gap-1">
			{#each memo.tags as tag}
				<a href="/tags/{tag.name}" class="rounded-full bg-[var(--color-bg-tertiary)] px-2 py-0.5 text-xs text-[var(--color-primary)]">
					#{tag.name}
				</a>
			{/each}
		</div>
	{/if}

	{#if memo.pinned}
		<span class="mt-1 inline-block text-xs text-[var(--color-accent)]">Pinned</span>
	{/if}
</article>
