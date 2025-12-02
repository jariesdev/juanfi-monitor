<svelte:head>
	<title>Logs</title>
</svelte:head>

<script lang="ts">
	import LogTable from './LogTable.svelte';
	import { baseApiUrl } from '$lib/env';

	let logTable: LogTable
	let isReloading: boolean = false;

	function refreshLogs(): void {
		isReloading = true;

		const request = new Request(`${baseApiUrl}/log/refresh`, { method: 'POST' });
		fetch(request)
			.then(() => {
				logTable.loadData();
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
		</div>

		<LogTable bind:this={logTable} />
	</div>
</div>
