<script lang="ts">
	import moment from 'moment/moment';
	import DateTime from '$lib/components/DateTime.svelte';
	import { baseApiUrl } from '$lib/env';
	import DataTable from '$lib/components/DataTable.svelte';
	import type { TableHeader } from '$lib/types/datatable';
	import {getVendos} from "$lib/remote/vendo.remote";

	let date: string = $state(moment().format('Y-MM-DD'));
	let vendoId: number|undefined = $state(undefined);

	const tableHeaders: TableHeader[] = [
		{label: 'Time', field: 'log_time'},
		{label: 'Vendo', field: 'vendo.name'},
		{label: 'Description', field: 'description'},
		{label: 'Created At', field: 'created_at'},
	]

	let tableFilters = $derived({
		vendo_id: vendoId,
		date: date,
	})

</script>

<DataTable url={`${baseApiUrl}/logs`} headers={tableHeaders} filters={tableFilters} title="System Logs">
	<div
		slot="before-table"
			class="uk-margin-small-top uk-grid uk-grid-small uk-child-width-1-2@s uk-child-width-1-3@m"
			style="row-gap: 15px"
		>
			<div>
				<input
					bind:value={date}
					type="date"
					name="date"
					id="date"
					class="uk-input uk-form-small uk-width-1-1"
				/>
			</div>
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
		</div>
	<span slot="cell" let:item let:header let:getCellValue>
		{#if header.field === 'log_time'}
			<DateTime date={item.log_time}></DateTime>
		{:else if header.field === 'created_at'}
			<DateTime date={item.created_at}></DateTime>
		{:else}
			<span>{getCellValue(item, header)}</span>
		{/if}
	</span>
</DataTable>
