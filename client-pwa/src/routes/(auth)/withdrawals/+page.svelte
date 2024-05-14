<script lang="ts">
	import { apiUrl } from '$lib/store';
	import WithdrawalTable from './WithdrawalTable.svelte';
	import { baseApiUrl } from '$lib/env';

	let isReloading: boolean = false;
	let reloadData: Function;

	function refreshLogs(): void {
		isReloading = true;

		const request = new Request(`${baseApiUrl}/withdrawals`, { method: 'GET' });
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
