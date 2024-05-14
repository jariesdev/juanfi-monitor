<script lang="ts">
	import { onMount } from 'svelte';
	import moment from 'moment/moment';
	import type { iVendo } from '$lib/types/models.js';
	import DateTime from '$lib/components/DateTime.svelte';
	import { baseApiUrl } from '$lib/env';
	import DataTable from '$lib/components/DataTable.svelte';
	import type { TableHeader } from '$lib/types/datatable';

	let date: string = moment().format('Y-MM-DD');
	let vendoId: number;
	let vendos: iVendo[] = [];

	const tableHeaders: TableHeader[] = [
		{label: 'Time', field: 'log_time'},
		{label: 'Vendo', field: 'vendo.name'},
		{label: 'Description', field: 'description'},
		{label: 'Created At', field: 'created_at'},
	]
	$: ({tableFilters} = {tableFilters: {
		vendo_id: vendoId,
		date: date,
		}})

	function loadOptions(): void {
		let url = `${baseApiUrl}/vendo-machines`;
		const request = new Request(url, { method: 'GET' });
		fetch(request)
			.then((response) => {
				if (response.status === 200) {
					return response.json();
				} else {
					throw new Error('Something went wrong on API server!');
				}
			})
			.then((response) => {
				vendos = response.data;
			})
			.catch(() => {
				vendos = [];
			});
	}

	onMount(() => {
		loadOptions();
	});
</script>

<DataTable url={`${baseApiUrl}/logs`} headers={tableHeaders} bind:filters={tableFilters} title="System Logs">
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
					{#each vendos as vendo}
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
