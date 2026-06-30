<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import SimpleTable from '$lib/components/SimpleTable.svelte';
	import RateForm from '$lib/components/RateForm.svelte';
	import ActionButton from '$lib/components/ActionButton.svelte';
	import { toast } from '$lib/store';
	import type { iVendoRate } from '$lib/types/models';
	import type { RowItem, SimpleTableHeader } from '$lib/types/datatable';

	interface Props {
		vendoId: number;
		isAdmin?: boolean;
	}

	const { vendoId, isAdmin = false }: Props = $props();

	let rates: iVendoRate[] = $state([]);
	let isLoading = $state(true);
	let isImporting = $state(false);
	let isSyncing = $state(false);
	let isSettingDefault = $state(false);
	let controller: AbortController | undefined;

	let showModal = $state(false);
	let editingRate: iVendoRate | null = $state(null);

	let showDefaultModal = $state(false);

	const headers: SimpleTableHeader[] = [
		{ label: 'Name', field: 'name' },
		{ label: 'Price', field: 'price', align: 'right' },
		{ label: 'Minutes', field: 'minutes', align: 'right' },
		{ label: 'Validity (mins)', field: 'validity_minutes', align: 'right' },
		{ label: 'Data Limit (MB)', field: 'data_limit_mb', align: 'right' },
		{ label: 'Profile', field: 'user_profile' },
		{ label: '', field: 'actions', sortable: false }
	];

	function loadRates() {
		controller?.abort();
		controller = new AbortController();
		isLoading = true;

		fetch(`/x-api/vendo-machines/${vendoId}/rates`, { signal: controller.signal })
			.then((r) => (r.ok ? r.json() : Promise.reject(r.statusText)))
			.then(({ data }) => {
				rates = data ?? [];
			})
			.catch(() => {})
			.finally(() => {
				isLoading = false;
			});
	}

	function openAdd() {
		editingRate = null;
		showModal = true;
	}

	function openEdit(rate: iVendoRate) {
		editingRate = rate;
		showModal = true;
	}

	function onFormSuccess() {
		showModal = false;
		loadRates();
	}

	async function deleteRate(rate: iVendoRate) {
		if (!confirm(`Delete rate "${rate.name}"?`)) return;
		const res = await fetch(`/x-api/vendo-rates/${rate.id}`, { method: 'DELETE' });
		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			toast.set({ message: body.detail ?? 'Failed to delete rate.', type: 'error' });
			return;
		}
		loadRates();
	}

	async function importFromMachine() {
		if (
			!confirm(
				"Importing will replace this vendo's current rate plan with the rates configured on the machine. Continue?"
			)
		)
			return;
		isImporting = true;
		try {
			const res = await fetch(`/x-api/vendo-machines/${vendoId}/rates/import`, { method: 'POST' });
			if (!res.ok) {
				const body = await res.json().catch(() => ({}));
				toast.set({ message: body.detail ?? 'Failed to import rates.', type: 'error' });
				return;
			}
			toast.set({ message: 'Rates imported from machine.', type: 'success' });
			loadRates();
		} finally {
			isImporting = false;
		}
	}

	async function syncToMachine() {
		if (
			!confirm(
				"Upload this vendo's saved rate plan to the device, overwriting the rates currently on the machine. Continue?"
			)
		)
			return;
		isSyncing = true;
		try {
			const res = await fetch(`/x-api/vendo-machines/${vendoId}/rates/sync`, { method: 'POST' });
			if (!res.ok) {
				const body = await res.json().catch(() => ({}));
				toast.set({ message: body.detail ?? 'Failed to sync rates to machine.', type: 'error' });
				return;
			}
			toast.set({ message: 'Rates synced to machine.', type: 'success' });
		} finally {
			isSyncing = false;
		}
	}

	async function setAsDefault(mode: 'copy' | 'replace') {
		isSettingDefault = true;
		try {
			const res = await fetch(`/x-api/vendo-machines/${vendoId}/rates/set-as-default`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ mode })
			});
			if (!res.ok) {
				const body = await res.json().catch(() => ({}));
				toast.set({ message: body.detail ?? 'Failed to set default rates.', type: 'error' });
				return;
			}
			toast.set({
				message:
					mode === 'copy'
						? 'Rates added to the default plan.'
						: 'Default rate plan replaced.',
				type: 'success'
			});
			showDefaultModal = false;
		} finally {
			isSettingDefault = false;
		}
	}

	onMount(loadRates);
	onDestroy(() => controller?.abort());
</script>

<div class="card">
	<div class="card-header">
		<span class="header-icon" uk-icon="icon: tag; ratio: 1"></span>
		<div class="header-left">
			<span class="card-title">Rate Plans</span>
			<button class="refresh-btn" onclick={loadRates} disabled={isLoading} title="Refresh">
				<span uk-icon="icon: refresh; ratio: 0.8"></span>
			</button>
		</div>
	</div>

	<div class="header-actions">
		<ActionButton
			icon="download"
			label={isImporting ? 'Importing…' : 'Import from Machine'}
			title="Import from Machine"
			onclick={importFromMachine}
			disabled={isImporting}
		/>
		<ActionButton
			icon="cloud-upload"
			label={isSyncing ? 'Syncing…' : 'Sync'}
			title="Sync"
			onclick={syncToMachine}
			disabled={isSyncing}
		/>
		{#if isAdmin}
			<ActionButton icon="list" label="Default Rates" href="/settings/default-rates" />
			<ActionButton
				icon="star"
				label="Set as Default"
				onclick={() => (showDefaultModal = true)}
				disabled={isSettingDefault}
			/>
		{/if}
		<ActionButton icon="plus" label="Add Rate" primary onclick={openAdd} />
	</div>

	{#if isLoading}
		<div class="skeleton-rows">
			{#each { length: 4 } as _}
				<div class="skeleton-row"></div>
			{/each}
		</div>
	{:else}
		<div class="table-wrap">
			<SimpleTable {headers} items={rates} emptyText="No rate plans yet.">
				{#snippet cell(item: RowItem, header: SimpleTableHeader, getCellValue: Function)}
					{#if header.field === 'price'}
						₱{Number(item.price).toLocaleString('en-PH', {
							minimumFractionDigits: 2,
							maximumFractionDigits: 2
						})}
					{:else if header.field === 'data_limit_mb'}
						{item.data_limit_mb ?? '—'}
					{:else if header.field === 'user_profile'}
						{item.user_profile ?? '—'}
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
			</SimpleTable>
		</div>
	{/if}
</div>

{#if showModal}
	<div id="rate-modal" class="uk-modal uk-open" style="display: block;">
		<div class="uk-modal-dialog uk-modal-body">
			<button
				class="uk-modal-close-default"
				type="button"
				uk-close
				onclick={() => (showModal = false)}
			></button>
			<h2 class="uk-modal-title">{editingRate ? 'Edit Rate' : 'Add Rate'}</h2>
			<RateForm
				rate={editingRate}
				{vendoId}
				onsuccess={onFormSuccess}
				oncancel={() => (showModal = false)}
			/>
		</div>
	</div>
	<div class="uk-modal-overlay" onclick={() => (showModal = false)}></div>
{/if}

{#if showDefaultModal}
	<div id="set-default-modal" class="uk-modal uk-open" style="display: block;">
		<div class="uk-modal-dialog uk-modal-body">
			<button
				class="uk-modal-close-default"
				type="button"
				uk-close
				onclick={() => (showDefaultModal = false)}
			></button>
			<h2 class="uk-modal-title">Set as Default Rates</h2>
			<p class="default-hint">
				Use this vendo's rate plan for the shared default template. <strong>Replace</strong> clears
				the current default plan first; <strong>Copy</strong> adds these rates on top of the existing
				default plan.
			</p>
			<div class="uk-text-center uk-margin-top">
				<button
					type="button"
					class="uk-button"
					disabled={isSettingDefault}
					onclick={() => setAsDefault('copy')}
				>
					Copy
				</button>
				<button
					type="button"
					class="uk-button uk-button-primary uk-margin-left"
					disabled={isSettingDefault}
					onclick={() => setAsDefault('replace')}
				>
					{isSettingDefault ? 'Saving…' : 'Replace'}
				</button>
			</div>
		</div>
	</div>
	<div class="uk-modal-overlay" onclick={() => (showDefaultModal = false)}></div>
{/if}

<style>
	.card {
		background: #fff;
		border: 1px solid #e8e8e8;
		border-radius: 10px;
		overflow: hidden;
	}

	.card-header {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 14px 18px;
		border-bottom: 1px solid #f0f0f0;
	}

	.header-left {
		display: flex;
		align-items: center;
		gap: 8px;
		flex: 1;
	}

	.header-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 30px;
		height: 30px;
		background: #fff4f1;
		border-radius: 7px;
		color: var(--color-theme-1);
		flex-shrink: 0;
	}

	.card-title {
		font-size: 0.88rem;
		font-weight: 700;
		color: #1a1a1a;
		margin-right: auto;
	}

	.refresh-btn {
		background: none;
		border: 1px solid #e8e8e8;
		border-radius: 6px;
		padding: 5px 8px;
		cursor: pointer;
		color: #888;
		display: flex;
		align-items: center;
		transition:
			background 0.15s,
			color 0.15s;
	}

	.refresh-btn:hover:not(:disabled) {
		background: #f5f5f5;
		color: #333;
	}

	.refresh-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.header-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		padding: 12px 18px;
		border-bottom: 1px solid #f0f0f0;
	}

	.table-wrap {
		overflow-x: auto;
	}

	.action-btns {
		display: flex;
		gap: 4px;
		align-items: center;
	}

	.uk-icon-button-danger {
		color: #dc2626;
	}

	.default-hint {
		font-size: 0.82rem;
		color: #666;
	}

	.skeleton-rows {
		padding: 14px 18px;
	}

	.skeleton-row {
		height: 32px;
		margin-bottom: 8px;
		border-radius: 6px;
		background: linear-gradient(90deg, #f0f0f0 25%, #e8e8e8 50%, #f0f0f0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.4s infinite;
	}

	@keyframes shimmer {
		0% {
			background-position: 200% 0;
		}
		100% {
			background-position: -200% 0;
		}
	}

	.uk-modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.4);
		z-index: 999;
	}

	#rate-modal,
	#set-default-modal {
		position: fixed;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}

	#rate-modal .uk-modal-dialog,
	#set-default-modal .uk-modal-dialog {
		max-width: 520px;
		width: 90%;
		max-height: 85vh;
		overflow-y: auto;
	}
</style>
