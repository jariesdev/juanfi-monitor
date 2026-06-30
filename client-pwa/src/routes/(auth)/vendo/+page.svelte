<svelte:head>
	<title>Vendo</title>
	<meta name="descript" content="List of vendo machines" />
</svelte:head>

<script lang="ts">
	import VendoTable from './VendoTable.svelte';

	const { data } = $props();
	const canManageRates = $derived(data.permissions?.includes('rates') ?? false);
</script>

<div class="uk-section">
	<div class="uk-container">
		<VendoTable {canManageRates} />
	</div>
</div>
