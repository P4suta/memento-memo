<script lang="ts">
	import { memoStore } from '$lib/stores/memos.svelte';

	let content = $state('');
	let submitting = $state(false);
	let error = $state('');

	async function submit() {
		if (!content.trim() || submitting) return;
		submitting = true;
		error = '';
		try {
			await memoStore.create(content);
			content = '';
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to create memo';
		} finally {
			submitting = false;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
			submit();
		}
	}
</script>

<div class="border-b border-[var(--color-border)] p-4">
	<textarea
		bind:value={content}
		onkeydown={handleKeydown}
		placeholder="What's on your mind? (Ctrl+Enter to post)"
		class="w-full resize-none rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-3 text-[var(--color-text)] placeholder-[var(--color-text-muted)] focus:border-[var(--color-primary)] focus:outline-none"
		rows="3"
		disabled={submitting}
	></textarea>
	{#if error}
		<p class="mt-1 text-sm text-red-400">{error}</p>
	{/if}
	<div class="mt-2 flex items-center justify-between">
		<span class="text-xs text-[var(--color-text-muted)]">
			Markdown supported
		</span>
		<button
			onclick={submit}
			disabled={!content.trim() || submitting}
			class="rounded-lg bg-[var(--color-primary)] px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-[var(--color-primary-hover)] disabled:opacity-50 disabled:cursor-not-allowed"
		>
			{submitting ? 'Posting...' : 'Post'}
		</button>
	</div>
</div>
