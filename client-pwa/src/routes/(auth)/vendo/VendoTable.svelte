<script lang="ts">
	import DataTable from '$lib/components/DataTable.svelte';
	import VendoForm from '$lib/components/VendoForm.svelte';
	import DateTime from '$lib/components/DateTime.svelte';
	import NumberFormat from '$lib/components/NumberFormat.svelte';
	import type { RowItem, TableHeader } from '$lib/types/datatable';

	let dataTable: DataTable;

	const tableHeaders: TableHeader[] = [
		{ label: 'Name',         field: 'name',                       sortable: true  },
		{ label: 'API URL',      field: 'api_url'                                     },
		{ label: 'Total Sales',  field: 'recent_status.total_sales',  sortable: true  },
		{ label: 'Current Sales',field: 'recent_status.current_sales',sortable: true  },
		{ label: 'Users',        field: 'recent_status.active_users', sortable: true  },
		{ label: 'Last Reported',field: 'recent_status.created_at',   sortable: true  },
		{ label: '',             field: 'actions'                                     },
	];

	export function loadData() {
		dataTable.loadData();
	}
</script>

<DataTable
	bind:this={dataTable}
	url={`/x-api/vendo-machines`}
	headers={tableHeaders}
	filters={{}}
	title="Vendo Machines"
	clientSort={true}
>
	{#snippet cell(item: RowItem, header: TableHeader, getCellValue: Function)}
		{#if header.field === 'recent_status.total_sales'}
			<NumberFormat value={item.recent_status?.total_sales} />
		{:else if header.field === 'recent_status.current_sales'}
			<NumberFormat value={item.recent_status?.current_sales} />
		{:else if header.field === 'recent_status.active_users'}
			<NumberFormat value={item.recent_status?.active_users || 0} />
		{:else if header.field === 'recent_status.created_at'}
			<DateTime date={item.recent_status?.created_at} />
		{:else if header.field === 'actions'}
			<a
				href={`/vendo/${item.id}/status`}
				class="uk-icon-button uk-button-primary"
				uk-icon="info"
				aria-label="View details"
			></a>
		{:else}
			{getCellValue(item, header)}
		{/if}
	{/snippet}

	{#snippet titleActions()}
		<button
			uk-toggle="target: #add-modal"
			type="button"
			class="uk-icon-button"
			uk-icon="icon: plus-circle"
			style="border: none;"
			title="Add vendo"
			aria-label="Add vendo"
		></button>
	{/snippet}

	{#snippet afterTable()}
		<!-- This is the modal -->
		<div id="add-modal" uk-modal class="uk-modal">
			<div class="uk-modal-dialog uk-modal-body">
				<h2 class="uk-modal-title">New Vendo</h2>
				<div class="uk-margin-small-bottom">
					<VendoForm onsuccess={loadData} />
				</div>
			</div>
		</div>
	{/snippet}
</DataTable>
