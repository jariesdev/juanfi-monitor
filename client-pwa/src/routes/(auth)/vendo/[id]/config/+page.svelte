<script lang="ts">
	import SystemConfig from './SystemConfig.svelte';

	const { data } = $props();
</script>

<svelte:head>
	<title>System Configuration</title>
</svelte:head>

<div class="uk-section">
	<div class="uk-container">
		<SystemConfig vendoId={data.id} />
	</div>
</div>
