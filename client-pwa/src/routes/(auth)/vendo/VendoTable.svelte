<script lang="ts">
	import DataTable from '$lib/components/DataTable.svelte';
	import VendoForm from '$lib/components/VendoForm.svelte';
	import DateTime from '$lib/components/DateTime.svelte';
	import NumberFormat from '$lib/components/NumberFormat.svelte';
	import { hasPermission } from '$lib/acl.svelte.js';
	import type { RowItem, TableHeader } from '$lib/types/datatable';

	let dataTable: DataTable;

	const tableHeaders: TableHeader[] = [
		{ label: 'Name', field: 'name', sortable: true },
		{ label: 'API URL', field: 'api_url' },
		{ label: 'Total Sales', field: 'recent_status.total_sales', sortable: true },
		{ label: 'Current Sales', field: 'recent_status.current_sales', sortable: true },
		{ label: 'Users', field: 'recent_status.active_users', sortable: true },
		{ label: 'Last Reported', field: 'recent_status.created_at', sortable: true },
		{ label: '', field: 'actions' }
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
			<div class="action-btns">
				<button
					type="button"
					class="uk-icon-button"
					uk-icon="icon: more-vertical"
					aria-label="Open actions"
				></button>
				<div uk-dropdown="mode: click; pos: bottom-right">
					<ul class="uk-nav uk-dropdown-nav action-menu">
						<li>
							<a href={`/vendo/${item.id}/status`}>
								<span class="action-icon" uk-icon="icon: info"></span>
								<span>View details</span>
							</a>
						</li>
						<li>
							<a href={`/vendo/${item.id}/active-users`}>
								<span class="action-icon" uk-icon="icon: users"></span>
								<span>Active users</span>
							</a>
						</li>
						{#if hasPermission('vendoconfig')}
							<li>
								<a href={`/vendo/${item.id}/config`}>
									<span class="action-icon" uk-icon="icon: cog"></span>
									<span>System configuration</span>
								</a>
							</li>
						{/if}
						{#if hasPermission('rates')}
							<li>
								<a href={`/vendo/${item.id}/rates`}>
									<span class="action-icon" uk-icon="icon: tag"></span>
									<span>Manage rates</span>
								</a>
							</li>
						{/if}
						{#if hasPermission('vouchers')}
							<li>
								<a href={`/vendo/${item.id}/vouchers`}>
									<svg
										class="action-icon"
										aria-hidden="true"
										width="20"
										height="20"
										viewBox="0 0 20 20"
										fill="none"
										stroke="currentColor"
										stroke-width="1.5"
										stroke-linecap="round"
										stroke-linejoin="round"
									>
										<path
											d="M4 6.5A1.5 1.5 0 0 1 5.5 5h9A1.5 1.5 0 0 1 16 6.5v2a1.5 1.5 0 0 0 0 3v2A1.5 1.5 0 0 1 14.5 15h-9A1.5 1.5 0 0 1 4 13.5v-2a1.5 1.5 0 0 0 0-3z"
										/>
										<path d="M8 7.25v5.5" />
										<path d="M11 8h2" />
										<path d="M11 12h2" />
									</svg>
									<span>Manage vouchers</span>
								</a>
							</li>
						{/if}
						{#if hasPermission('logs')}
							<li>
								<a href={`/logs?vendo_id=${item.id}`}>
									<span class="action-icon" uk-icon="icon: file-text"></span>
									<span>Logs</span>
								</a>
							</li>
						{/if}
					</ul>
				</div>
			</div>
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

<style>
	.action-btns {
		display: flex;
		gap: 4px;
		align-items: center;
		justify-content: flex-end;
	}

	.action-btns :global(.uk-dropdown) {
		min-width: 190px;
		padding: 8px 0;
	}

	.action-menu a {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 8px 14px;
	}

	.action-icon {
		width: 20px;
		height: 20px;
		flex: 0 0 20px;
	}
</style>
