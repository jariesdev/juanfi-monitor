<script lang="ts">
	import DataTable from '$lib/components/DataTable.svelte';
	import VendoForm from '$lib/components/VendoForm.svelte';
	import DateTime from '$lib/components/DateTime.svelte';
	import NumberFormat from '$lib/components/NumberFormat.svelte';
	import VendoStatusBadge from '$lib/components/VendoStatusBadge.svelte';
	import { hasPermission } from '$lib/acl.svelte.js';
	import { refreshingVendos, vendoOnline } from '$lib/store/vendoActivity';
	import type { RowItem, TableHeader } from '$lib/types/datatable';

	let dataTable: DataTable;

	const tableHeaders: TableHeader[] = [
		{ label: 'Name', field: 'name', sortable: true },
		{ label: 'API URL', field: 'api_url' },
		{ label: 'Status', field: 'is_online', sortable: true },
		{ label: 'Total Sales', field: 'recent_status.total_sales', sortable: true },
		{ label: 'Current Sales', field: 'recent_status.current_sales', sortable: true },
		{ label: 'Users', field: 'recent_status.active_users', sortable: true },
		{ label: 'Commission', field: 'commission', sortable: true },
		{ label: 'Last Reported', field: 'recent_status.created_at', sortable: true },
		{ label: '', field: 'actions' }
	];

	export function loadData() {
		dataTable.loadData();
	}

	// Edit modal state
	let showEdit = $state(false);
	let editId: number | null = $state(null);
	let editName = $state('');
	let editApiUrl = $state('');
	let editCommission: number | '' = $state(0);
	let editSaving = $state(false);
	let editError = $state('');

	function openEdit(item: RowItem) {
		editId = item.id;
		editName = item.name ?? '';
		editApiUrl = item.api_url ?? '';
		editCommission = item.commission ?? 0;
		editError = '';
		showEdit = true;
	}

	async function saveEdit() {
		editError = '';
		if (editCommission === '' || Number(editCommission) < 0 || Number(editCommission) > 100) {
			editError = 'Commission must be between 0 and 100.';
			return;
		}
		editSaving = true;
		try {
			const res = await fetch(`/x-api/vendo-machines/${editId}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
				body: JSON.stringify({
					name: editName,
					api_url: editApiUrl,
					commission: Number(editCommission)
				})
			});
			if (!res.ok) {
				const body = await res.json().catch(() => ({}));
				editError = body.detail ?? 'Failed to save vendo.';
				return;
			}
			showEdit = false;
			loadData();
		} catch {
			editError = 'An unexpected error occurred.';
		} finally {
			editSaving = false;
		}
	}
</script>

<DataTable
	bind:this={dataTable}
	url={`/x-api/vendo-machines`}
	headers={tableHeaders}
	filters={{}}
	title="Vendo Machines"
	clientSort={true}
	rowClass={(item) => (item.is_active ? undefined : 'inactive')}
>
	{#snippet cell(item: RowItem, header: TableHeader, getCellValue: Function)}
		{#if header.field === 'is_online'}
			<VendoStatusBadge
				online={$vendoOnline.get(item.id) ?? !!item.is_online}
				active={!!item.is_active}
				progress={$refreshingVendos.get(item.id) ?? null}
			/>
		{:else if header.field === 'recent_status.total_sales'}
			₱<NumberFormat value={item.recent_status?.total_sales} />
		{:else if header.field === 'recent_status.current_sales'}
			₱<NumberFormat value={item.recent_status?.current_sales} />
		{:else if header.field === 'recent_status.active_users'}
			<NumberFormat value={item.recent_status?.active_users || 0} />
		{:else if header.field === 'commission'}
			{item.commission ?? 0}%
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
							<button type="button" class="action-link" onclick={() => openEdit(item)}>
								<span class="action-icon" uk-icon="icon: pencil"></span>
								<span>Edit vendo</span>
							</button>
						</li>
						<li>
							<a href={`/vendo/${item.id}/active-users`}>
								<span class="action-icon" uk-icon="icon: users"></span>
								<span>Active users</span>
							</a>
						</li>
						{#if hasPermission('withdrawals')}
							<li>
								<a href={`/vendo/${item.id}/withdraw`}>
									<span class="action-icon" uk-icon="icon: credit-card"></span>
									<span>Withdraw Sale</span>
								</a>
							</li>
						{/if}
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

{#if showEdit}
	<div
		class="modal-backdrop"
		role="button"
		tabindex="-1"
		onclick={() => (showEdit = false)}
		onkeydown={(e) => e.key === 'Escape' && (showEdit = false)}
	>
		<div
			class="edit-modal"
			role="dialog"
			tabindex="-1"
			onclick={(e) => e.stopPropagation()}
			onkeydown={() => {}}
		>
			<div class="edit-title">Edit Vendo</div>
			<form
				onsubmit={(e) => {
					e.preventDefault();
					saveEdit();
				}}
			>
				{#if editError}
					<div class="uk-alert uk-alert-danger">{editError}</div>
				{/if}
				<div class="uk-margin-small-bottom">
					<label class="uk-form-label" for="edit-name">Name</label>
					<input id="edit-name" class="uk-input" type="text" bind:value={editName} required />
				</div>
				<div class="uk-margin-small-bottom">
					<label class="uk-form-label" for="edit-apiurl">API URL</label>
					<input id="edit-apiurl" class="uk-input" type="text" bind:value={editApiUrl} />
				</div>
				<div class="uk-margin-bottom">
					<label class="uk-form-label" for="edit-commission">Commission (%)</label>
					<input
						id="edit-commission"
						class="uk-input"
						type="number"
						min="0"
						max="100"
						step="0.01"
						bind:value={editCommission}
					/>
				</div>
				<div class="edit-actions">
					<button
						type="button"
						class="uk-button uk-button-default"
						onclick={() => (showEdit = false)}>Cancel</button
					>
					<button type="submit" class="uk-button uk-button-primary" disabled={editSaving}>
						{editSaving ? 'Saving…' : 'Save'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

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

	.action-menu a,
	.action-menu .action-link {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 8px 14px;
	}

	.action-menu .action-link {
		width: 100%;
		background: none;
		border: none;
		cursor: pointer;
		font: inherit;
		color: inherit;
		text-align: left;
	}
	.action-menu .action-link:hover {
		background: #f5f5f5;
	}

	.modal-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.4);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1010;
		padding: 16px;
	}
	.edit-modal {
		background: #fff;
		border-radius: 12px;
		padding: 22px;
		width: 100%;
		max-width: 420px;
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
	}
	.edit-title {
		font-size: 1.05rem;
		font-weight: 700;
		color: #1a1a1a;
		margin-bottom: 16px;
	}
	.edit-actions {
		display: flex;
		justify-content: flex-end;
		gap: 8px;
		margin-top: 8px;
	}

	.action-icon {
		width: 20px;
		height: 20px;
		flex: 0 0 20px;
	}

	/* Inactive vendos are de-emphasized across the whole row. The row lives inside
	   DataTable's scope, so target it globally. */
	:global(.data-table tr.inactive td) {
		color: #9ca3af;
	}
</style>
