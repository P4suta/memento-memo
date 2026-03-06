<script lang="ts">
	let { html }: { html: string } = $props();
</script>

<div class="prose prose-sm max-w-none">
	{@html html}
</div>
