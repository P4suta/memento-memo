<script lang="ts">
	import MemoComposer from '$lib/components/memo/MemoComposer.svelte';
	import MemoTimeline from '$lib/components/memo/MemoTimeline.svelte';
</script>

<svelte:head>
	<title>Memento Memo</title>
</svelte:head>

<div class="mx-auto max-w-2xl">
	<MemoComposer />
	<MemoTimeline />
</div>
