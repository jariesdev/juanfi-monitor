<script lang="ts">
	import DataTable from '$lib/components/DataTable.svelte';
	import DateTime from '$lib/components/DateTime.svelte';

	import {getVendos} from "$lib/remote/vendo.remote";
	import type {RowItem, TableHeader} from "$lib/types/datatable";

	let dataTable: DataTable
	let headers = [
		{ label: 'Time', field: 'sale_time' },
		{ label: 'MAC Address', field: 'mac_address' },
		{ label: 'Vendo', field: 'vendo.name' },
		{ label: 'Amount', field: 'amount' },
		{ label: 'Voucher', field: 'voucher' }
	];
	let vendoId: number|undefined = $state(undefined);
	let saleTime: string = $state('');
	let tableFilters = $derived({
		vendo_id: vendoId,
		date: saleTime
	})

	export function loadData() {
    dataTable.loadData()
  }
</script>

<DataTable bind:this={dataTable} url={`/x-api/sales`} {headers} filters={tableFilters} title="Sales">

	{#snippet beforeTable()}
		<div>
			<div
				class="uk-margin-small-top uk-grid uk-grid-small uk-child-width-1-2@s uk-child-width-1-3@m"
				style="row-gap: 15px"
			>
				<div>
					<select
						bind:value={vendoId}
						name="vendo_id"
						id="vendo_id"
						class="uk-select uk-form-small uk-child-width-1-1"
					>
						<option value={undefined}>All</option>
						{#each await getVendos() as vendo}
							<option value={vendo.id}>{vendo.name}</option>
						{/each}
					</select>
				</div>

				<div>
					<input
						type="date"
						bind:value={saleTime}
						class="uk-input uk-form-small"
						max={new Date().toISOString().split('T')[0]}
					/>
				</div>
			</div>
		</div>
	{/snippet}

	{#snippet cell(item: RowItem, header: TableHeader, getCellValue: Function)}
	<span>
		{#if header.field === 'sale_time'}
			<DateTime humanized={true} date={item.sale_time} />
		{:else}
			{getCellValue(item, header)}
		{/if}
	</span>
	{/snippet}
</DataTable>
