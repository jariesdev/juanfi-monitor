<script lang="ts">
	import LogTable from './LogTable.svelte';

	const { data } = $props();
</script>

<svelte:head>
	<title>Logs</title>
</svelte:head>

<div class="uk-section">
	<div class="uk-container">
		<LogTable initialVendoId={data.initialVendoId} />
	</div>
</div>
