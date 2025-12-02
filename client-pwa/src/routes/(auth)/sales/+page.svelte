<svelte:head>
	<title>Sales</title>
</svelte:head>

<script lang="ts">
	import SaleTable from './SaleTable.svelte';
	import { baseApiUrl } from '$lib/env';

	let isReloading: boolean = false;
	let saleTable: SaleTable;

	function refreshLogs(): void {
		isReloading = true;

		const request = new Request(`${baseApiUrl}/log/refresh`, { method: 'POST' });
		fetch(request)
			.then(() => {
				saleTable.loadData();
			})
			.finally(() => {
				isReloading = false;
			});
	}
</script>

<div class="uk-section">
	<div class="uk-container">
		<div class="uk-margin-bottom">
			<button class="uk-button uk-button-primary" disabled={isReloading} on:click={refreshLogs}
				>Refresh</button
			>
			<a href="/withdrawals" class="uk-button uk-button-primary">Withdrawals</a>
		</div>

		<SaleTable bind:this={saleTable} />
	</div>
</div>
