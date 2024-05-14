<script lang="ts">
	import LogTable from './SaleTable.svelte';
	import { apiUrl } from '$lib/store';

	let isReloading: boolean = false;
	let reloadData: Function;
	let baseApiUrl: string = '';

	function refreshLogs(): void {
		isReloading = true;

		const request = new Request(`${baseApiUrl}/log/refresh`, { method: 'POST' });
		fetch(request)
			.then(() => {
				reloadData();
			})
			.finally(() => {
				isReloading = false;
			});
	}

	apiUrl.subscribe(function (value) {
		baseApiUrl = value;
	});
</script>

<div class="uk-section">
	<div class="uk-container">
		<div class="uk-margin-bottom">
			<button class="uk-button uk-button-primary" disabled={isReloading} on:click={refreshLogs}
				>Refresh</button
			>
			<a href="/withdrawals" class="uk-button uk-button-primary">Withdrawals</a>
		</div>

		<LogTable bind:loadData={reloadData} />
	</div>
</div>
