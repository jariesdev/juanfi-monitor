<script lang="ts">
	import DataTable from '$lib/components/DataTable.svelte';
	import DateTime from '$lib/components/DateTime.svelte';
	import type { RowItem, TableHeader } from '$lib/types/datatable';
	import { incomingNotification } from '$lib/store/notifications';
	import { notificationTypeFilter } from '$lib/store';

	let dataTable: DataTable;

	const tableHeaders: TableHeader[] = [
		{ label: 'Time', field: 'created_at' },
		{ label: 'Type', field: 'type' },
		{ label: 'Message', field: 'message' }
	];

	// The list is scoped to the current user server-side (GET /notifications
	// always filters by the authenticated caller, never a client-supplied ID).

	// Bound directly to the persisted store, so the selected filter survives
	// navigating away and back (or a full page reload).
	let tableFilters = $derived({ type: $notificationTypeFilter || undefined });

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
	filters={tableFilters}
	title="Notifications"
	showRefresh
>
	{#snippet beforeTable()}
		<div class="uk-margin-small-top uk-grid uk-grid-small uk-child-width-1-2@s uk-child-width-1-3@m">
			<div>
				<select
					bind:value={$notificationTypeFilter}
					name="type"
					id="type"
					class="uk-select uk-form-small uk-child-width-1-1"
				>
					<option value="">All</option>
					<option value="info">Info</option>
					<option value="alert">Alert</option>
				</select>
			</div>
		</div>
	{/snippet}

	{#snippet cell(item: RowItem, header: TableHeader, getCellValue: Function)}
		<span>
			{#if header.field === 'created_at'}
				<DateTime date={item.created_at}></DateTime>
			{:else if header.field === 'type'}
				<span class="type-badge {item.type ?? 'info'}">{item.type ?? 'info'}</span>
			{:else}
				<span>{getCellValue(item, header)}</span>
			{/if}
		</span>
	{/snippet}
</DataTable>

<style>
	.type-badge {
		display: inline-block;
		padding: 2px 8px;
		border-radius: 12px;
		font-size: 0.78rem;
		font-weight: 600;
		text-transform: capitalize;
	}

	.type-badge.info {
		background: #eef2ff;
		color: #3730a3;
	}

	.type-badge.alert {
		background: #fee2e2;
		color: #991b1b;
	}
</style>
