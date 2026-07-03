<script lang="ts">
	import DataTable from '$lib/components/DataTable.svelte';
	import DateTime from '$lib/components/DateTime.svelte';
	import type { RowItem, TableHeader } from '$lib/types/datatable';
	import { incomingNotification } from '$lib/store/notifications';

	let dataTable: DataTable;

	const tableHeaders: TableHeader[] = [
		{ label: 'Time', field: 'created_at' },
		{ label: 'Message', field: 'message' }
	];

	// Live update: reload page 1 whenever a new notification frame arrives
	// over the WebSocket (already filtered for the current user upstream).
	$effect(() => {
		if ($incomingNotification) {
			dataTable?.refresh();
		}
	});
</script>

<DataTable
	bind:this={dataTable}
	url={`/x-api/notifications`}
	headers={tableHeaders}
	filters={{}}
	title="Notifications"
	showRefresh
>
	{#snippet cell(item: RowItem, header: TableHeader, getCellValue: Function)}
		<span>
			{#if header.field === 'created_at'}
				<DateTime date={item.created_at}></DateTime>
			{:else}
				<span>{getCellValue(item, header)}</span>
			{/if}
		</span>
	{/snippet}
</DataTable>
