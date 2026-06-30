<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import ActionButton from '$lib/components/ActionButton.svelte';
	import SimpleTable from '$lib/components/SimpleTable.svelte';
	import VoucherForm from '$lib/components/VoucherForm.svelte';
	import { toast } from '$lib/store';
	import type { RowItem, SimpleTableHeader } from '$lib/types/datatable';
	import type { iVendoVoucher } from '$lib/types/models';

	interface Props {
		vendoId: number;
	}

	const { vendoId }: Props = $props();

	let vouchers = $state<iVendoVoucher[]>([]);
	let isLoading = $state(true);
	let showModal = $state(false);
	let controller: AbortController | undefined;
	let latestCodes = $state<string[]>([]);

	const headers: SimpleTableHeader[] = [
		{ label: 'Code', field: 'code' },
		{ label: 'Prefix', field: 'prefix' },
		{ label: 'Amount', field: 'amount', align: 'right' },
		{ label: 'Duration', field: 'duration_minutes' },
		{ label: 'Added to Sales', field: 'added_to_sales' },
		{ label: 'Created At', field: 'created_at' }
	];

	function formatMinutes(mins: number): string {
		const total = Math.max(0, Math.floor(Number(mins) || 0));
		const days = Math.floor(total / 1440);
		const hours = Math.floor((total % 1440) / 60);
		const minutes = total % 60;
		let out = '';
		if (days) out += `${days}d`;
		if (hours) out += `${hours}h`;
		if (minutes || !out) out += `${minutes}m`;
		return out;
	}

	function formatDate(value: string): string {
		const date = new Date(value);
		if (Number.isNaN(date.getTime())) return '—';
		return date.toLocaleString('en-PH', {
			year: 'numeric',
			month: 'short',
			day: '2-digit',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function loadVouchers() {
		controller?.abort();
		controller = new AbortController();
		isLoading = true;

		fetch(`/x-api/vendo-machines/${vendoId}/vouchers`, { signal: controller.signal })
			.then(async (res) => {
				if (!res.ok) {
					const body = await res.json().catch(() => ({}));
					throw new Error(body.detail ?? 'Failed to load vouchers.');
				}
				return res.json();
			})
			.then(({ data }) => {
				vouchers = data ?? [];
			})
			.catch((err) => {
				if (err?.name !== 'AbortError') {
					toast.set({ message: err.message ?? 'Failed to load vouchers.', type: 'error' });
				}
			})
			.finally(() => {
				isLoading = false;
			});
	}

	function onFormSuccess(created: iVendoVoucher[]) {
		showModal = false;
		latestCodes = created.map((v) => v.code);
		loadVouchers();
		if (latestCodes.length > 0) {
			toast.set({
				message: `Generated: ${latestCodes.join(', ')}`,
				type: 'success'
			});
		}
	}

	onMount(loadVouchers);
	onDestroy(() => controller?.abort());
</script>

<div class="card">
	<div class="card-header">
		<span class="header-icon" uk-icon="icon: credit-card; ratio: 1"></span>
		<div class="header-left">
			<span class="card-title">Generated Vouchers</span>
			<button class="refresh-btn" onclick={loadVouchers} disabled={isLoading} title="Refresh">
				<span uk-icon="icon: refresh; ratio: 0.8"></span>
			</button>
		</div>
	</div>

	<div class="header-actions">
		<ActionButton icon="plus" label="Generate" primary onclick={() => (showModal = true)} />
	</div>

	{#if isLoading}
		<div class="skeleton-rows">
			{#each { length: 4 } as _}
				<div class="skeleton-row"></div>
			{/each}
		</div>
	{:else}
		<div class="table-wrap">
			<SimpleTable {headers} items={vouchers} emptyText="No generated vouchers yet.">
				{#snippet cell(item: RowItem, header: SimpleTableHeader, getCellValue: Function)}
					{#if header.field === 'code'}
						<span class:highlight={latestCodes.includes(String(item.code))}>{item.code}</span>
					{:else if header.field === 'amount'}
						₱{Number(item.amount).toLocaleString('en-PH', {
							minimumFractionDigits: 2,
							maximumFractionDigits: 2
						})}
					{:else if header.field === 'duration_minutes'}
						{item.duration_minutes} ({formatMinutes(item.duration_minutes as number)})
					{:else if header.field === 'added_to_sales'}
						{item.added_to_sales ? 'Yes' : 'No'}
					{:else if header.field === 'created_at'}
						{formatDate(String(item.created_at))}
					{:else}
						{getCellValue(item, header)}
					{/if}
				{/snippet}
			</SimpleTable>
		</div>
	{/if}
</div>

{#if showModal}
	<div id="voucher-modal" class="uk-modal uk-open" style="display: block;">
		<div class="uk-modal-dialog uk-modal-body">
			<button
				class="uk-modal-close-default"
				type="button"
				uk-close
				aria-label="Close"
				onclick={() => (showModal = false)}
			></button>
			<h2 class="uk-modal-title">Generate Vouchers</h2>
			<VoucherForm {vendoId} onsuccess={onFormSuccess} oncancel={() => (showModal = false)} />
		</div>
	</div>
	<button
		type="button"
		class="uk-modal-overlay"
		aria-label="Close"
		onclick={() => (showModal = false)}
	></button>
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

	.highlight {
		background: #fff8da;
		padding: 2px 6px;
		border-radius: 5px;
		font-weight: 600;
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
		border: 0;
		padding: 0;
		z-index: 999;
	}

	#voucher-modal {
		position: fixed;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}

	#voucher-modal .uk-modal-dialog {
		max-width: 520px;
		width: 90%;
		max-height: 85vh;
		overflow-y: auto;
	}
</style>
