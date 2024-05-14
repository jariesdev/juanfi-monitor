<script lang="ts">
	import { useLocalTimeFormat } from '$lib/store/app';

	export let date: string = '';
	export let humanized: boolean = false;

	useLocalTimeFormat.subscribe((value) => (humanized = value));

	function humanizeTime(): string {
		return date ? new Date(Date.parse(date)).toLocaleString() : '';
	}
</script>

<div
	on:click={() => useLocalTimeFormat.set(!humanized)}
	on:keydown
	on:keyup
	class={$$restProps.class || ''}
>
	{#if humanized}
		<span title={date}>
			{humanizeTime()}
		</span>
	{:else}
		<span title={humanizeTime()}>
			{date}
		</span>
	{/if}
</div>
