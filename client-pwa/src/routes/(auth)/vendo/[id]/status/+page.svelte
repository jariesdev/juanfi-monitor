<svelte:head>
	<title>Vendo Status</title>
</svelte:head>

<script lang="ts">
	import SystemStatus from './SystemStatus.svelte';

	export let data;
</script>

<div class="uk-section">
	<div class="uk-container">
		<SystemStatus vendoId={data.id} />
	</div>
</div>
