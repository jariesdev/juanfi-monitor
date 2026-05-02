<svelte:head>
	<title>Withdrawals</title>
</svelte:head>

<script lang="ts">
	import WithdrawalTable from './WithdrawalTable.svelte';

	let isReloading: boolean = false;
	let reloadData: Function;

	function refreshLogs(): void {
		isReloading = true;

		const request = new Request(`/x-api/withdrawals`, { method: 'GET' });
		fetch(request)
			.then(() => {
				reloadData();
			})
			.finally(() => {
				isReloading = false;
			});
	}
</script>

<div class="uk-section">
	<div class="uk-container">
		<WithdrawalTable bind:loadData={reloadData} />
	</div>
</div>
