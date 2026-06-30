<script lang="ts">
	import DataTable from '$lib/components/DataTable.svelte';
	import RateForm from '$lib/components/RateForm.svelte';
	import { toast } from '$lib/store';
	import type { iVendoRate, iVendo } from '$lib/types/models';
	import type { RowItem, TableHeader } from '$lib/types/datatable';

	let dataTable: DataTable;
	let editingRate: iVendoRate | null = $state(null);
	let showFormModal = $state(false);
	let deleteError = $state('');

	let showApplyModal = $state(false);
	let vendos: iVendo[] = $state([]);
	let selectedVendoIds: Set<number> = $state(new Set());
	let isLoadingVendos = $state(false);
	let isApplying = $state(false);

	const headers: TableHeader[] = [
		{ label: 'Name', field: 'name', sortable: true },
		{ label: 'Price', field: 'price', sortable: true },
		{ label: 'Minutes', field: 'minutes', sortable: true },
		{ label: 'Validity (mins)', field: 'validity_minutes', sortable: true },
		{ label: 'Data Limit (MB)', field: 'data_limit_mb', sortable: false },
		{ label: 'Profile', field: 'user_profile', sortable: false },
		{ label: '', field: 'actions', sortable: false }
	];

	function openAdd() {
		editingRate = null;
		showFormModal = true;
	}

	function openEdit(rate: iVendoRate) {
		editingRate = rate;
		showFormModal = true;
	}

	function onFormSuccess() {
		showFormModal = false;
		dataTable.refresh();
	}

	async function deleteRate(rate: iVendoRate) {
		deleteError = '';
		if (!confirm(`Delete default rate "${rate.name}"?`)) return;
		const res = await fetch(`/x-api/vendo-rates/${rate.id}`, { method: 'DELETE' });
		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			deleteError = body.detail ?? 'Failed to delete rate.';
			return;
		}
		dataTable.refresh();
	}

	const allSelected = $derived(vendos.length > 0 && selectedVendoIds.size === vendos.length);

	async function openApplyModal() {
		showApplyModal = true;
		isLoadingVendos = true;
		selectedVendoIds = new Set();
		try {
			const res = await fetch('/x-api/vendo-machines');
			const body = await res.json().catch(() => ({ data: [] }));
			vendos = body.data ?? [];
		} finally {
			isLoadingVendos = false;
		}
	}

	function toggleVendo(id: number) {
		const next = new Set(selectedVendoIds);
		if (next.has(id)) next.delete(id);
		else next.add(id);
		selectedVendoIds = next;
	}

	function toggleSelectAll() {
		selectedVendoIds = allSelected ? new Set() : new Set(vendos.map((v) => v.id));
	}

	async function applyToAll() {
		if (selectedVendoIds.size === 0) return;
		isApplying = true;
		try {
			const res = await fetch('/x-api/vendo-rates/apply-to-all', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ vendo_ids: [...selectedVendoIds] })
			});
			if (!res.ok) {
				const body = await res.json().catch(() => ({}));
				toast.set({ message: body.detail ?? 'Failed to apply default rates.', type: 'error' });
				return;
			}
			toast.set({
				message: `Default rates applied to ${selectedVendoIds.size} vendo(s).`,
				type: 'success'
			});
			showApplyModal = false;
		} finally {
			isApplying = false;
		}
	}
</script>

<svelte:head>
	<title>Default Rates</title>
</svelte:head>

{#if deleteError}
	<div class="uk-alert uk-alert-danger uk-margin-small-bottom" role="alert">
		{deleteError}
		<button
			type="button"
			class="uk-alert-close"
			onclick={() => (deleteError = '')}
			aria-label="Close"
		></button>
	</div>
{/if}

<DataTable
	bind:this={dataTable}
	url="/x-api/vendo-rates/default"
	{headers}
	filters={{}}
	title="Default Rates"
	clientSort={true}
>
	{#snippet titleActions()}
		<button
			type="button"
			class="uk-icon-button"
			uk-icon="icon: world"
			style="border: none;"
			title="Apply to all vendos"
			aria-label="Apply to all vendos"
			onclick={openApplyModal}
		></button>
		<button
			type="button"
			class="uk-icon-button"
			uk-icon="icon: plus-circle"
			style="border: none;"
			title="Add rate"
			aria-label="Add rate"
			onclick={openAdd}
		></button>
	{/snippet}

	{#snippet cell(item: RowItem, header: TableHeader, getCellValue: Function)}
		{#if header.field === 'price'}
			₱{Number((item as iVendoRate).price).toLocaleString('en-PH', {
				minimumFractionDigits: 2,
				maximumFractionDigits: 2
			})}
		{:else if header.field === 'data_limit_mb'}
			{(item as iVendoRate).data_limit_mb ?? '—'}
		{:else if header.field === 'user_profile'}
			{(item as iVendoRate).user_profile ?? '—'}
		{:else if header.field === 'actions'}
			<div class="action-btns">
				<button
					type="button"
					class="uk-icon-button"
					uk-icon="pencil"
					title="Edit"
					onclick={() => openEdit(item as iVendoRate)}
				></button>
				<button
					type="button"
					class="uk-icon-button uk-icon-button-danger"
					uk-icon="trash"
					title="Delete"
					onclick={() => deleteRate(item as iVendoRate)}
				></button>
			</div>
		{:else}
			{getCellValue(item, header)}
		{/if}
	{/snippet}

	{#snippet afterTable()}
		{#if showFormModal}
			<div id="default-rate-modal" class="uk-modal uk-open" style="display: block;">
				<div class="uk-modal-dialog uk-modal-body">
					<button
						class="uk-modal-close-default"
						type="button"
						uk-close
						onclick={() => (showFormModal = false)}
					></button>
					<h2 class="uk-modal-title">{editingRate ? 'Edit Rate' : 'New Default Rate'}</h2>
					<RateForm
						rate={editingRate}
						vendoId={null}
						onsuccess={onFormSuccess}
						oncancel={() => (showFormModal = false)}
					/>
				</div>
			</div>
			<div class="uk-modal-overlay" onclick={() => (showFormModal = false)}></div>
		{/if}

		{#if showApplyModal}
			<div id="apply-modal" class="uk-modal uk-open" style="display: block;">
				<div class="uk-modal-dialog uk-modal-body">
					<button
						class="uk-modal-close-default"
						type="button"
						uk-close
						onclick={() => (showApplyModal = false)}
					></button>
					<h2 class="uk-modal-title">Apply Default Rates</h2>
					<p class="apply-hint">
						Choose which vendos should receive the default rate plan. This replaces each selected
						vendo's current rates.
					</p>

					{#if isLoadingVendos}
						<p class="uk-text-muted uk-text-small">Loading vendos…</p>
					{:else if vendos.length === 0}
						<p class="uk-text-muted uk-text-small">No vendos found.</p>
					{:else}
						<label class="vendo-item select-all">
							<input
								type="checkbox"
								class="uk-checkbox"
								checked={allSelected}
								onchange={toggleSelectAll}
							/>
							<span>Select all</span>
						</label>
						<div class="vendo-list">
							{#each vendos as v (v.id)}
								<label class="vendo-item">
									<input
										type="checkbox"
										class="uk-checkbox"
										checked={selectedVendoIds.has(v.id)}
										onchange={() => toggleVendo(v.id)}
									/>
									<span>{v.name}</span>
								</label>
							{/each}
						</div>
					{/if}

					<div class="uk-text-center uk-margin-top">
						<button
							type="button"
							class="uk-modal-close uk-button"
							onclick={() => (showApplyModal = false)}
						>
							Cancel
						</button>
						<button
							type="button"
							class="uk-button uk-button-primary uk-margin-left"
							disabled={isApplying || selectedVendoIds.size === 0}
							onclick={applyToAll}
						>
							{isApplying ? 'Applying…' : `Apply to Selected (${selectedVendoIds.size})`}
						</button>
					</div>
				</div>
			</div>
			<div class="uk-modal-overlay" onclick={() => (showApplyModal = false)}></div>
		{/if}
	{/snippet}
</DataTable>

<style>
	.action-btns {
		display: flex;
		gap: 4px;
		align-items: center;
	}

	.uk-icon-button-danger {
		color: #dc2626;
	}

	.apply-hint {
		font-size: 0.82rem;
		color: #666;
	}

	.vendo-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
		max-height: 240px;
		overflow-y: auto;
		border: 1px solid #e8e8e8;
		border-radius: 4px;
		padding: 10px 12px;
	}

	.vendo-item {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 0.875rem;
		cursor: pointer;
	}

	.select-all {
		font-weight: 600;
		margin-bottom: 8px;
	}

	.uk-modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.4);
		z-index: 999;
	}

	#default-rate-modal,
	#apply-modal {
		position: fixed;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}

	#default-rate-modal .uk-modal-dialog,
	#apply-modal .uk-modal-dialog {
		max-width: 520px;
		width: 90%;
		max-height: 85vh;
		overflow-y: auto;
	}
</style>
