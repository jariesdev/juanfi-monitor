<script lang="ts">
	import DataTable from '$lib/components/DataTable.svelte';
	import DateTime from '$lib/components/DateTime.svelte';
	import NumberFormat from '$lib/components/NumberFormat.svelte';
	import type { RowItem, TableHeader } from '$lib/types/datatable';

	let dataTable: DataTable;

	const tableHeaders: TableHeader[] = [
		{ label: 'Vendo', field: 'vendo.name' },
		{ label: 'Amount', field: 'amount' },
		{ label: 'Date', field: 'created_at' }
	];

	export function loadData() {
		dataTable.loadData();
	}
</script>

<DataTable
	bind:this={dataTable}
	url="/x-api/withdrawals"
	headers={tableHeaders}
	filters={{}}
	title="Withdrawals"
>
	{#snippet cell(item: RowItem, header: TableHeader, getCellValue: Function)}
		{#if header.field === 'amount'}
			<NumberFormat value={item.amount} />
		{:else if header.field === 'created_at'}
			<DateTime date={item.created_at} class="uk-text-nowrap" />
		{:else}
			{getCellValue(item, header)}
		{/if}
	{/snippet}
</DataTable>
