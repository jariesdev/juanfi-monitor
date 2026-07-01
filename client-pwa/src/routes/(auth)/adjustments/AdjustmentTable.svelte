<script lang="ts">
	import { onMount } from 'svelte';
	import refreshIcon from '$lib/icons/refresh.svg';

	interface Adjustment {
		id: number;
		month: string; // YYYY-MM
		amount: number; // signed
		description: string;
	}

	let allRows: Adjustment[] = $state([]);
	let isLoading = $state(true);
	let isRefreshing = $state(false);

	// modal / form state
	let showModal = $state(false);
	let editingId: number | null = $state(null);
	let fMonth = $state(currentMonth());
	let fAmount: number | '' = $state('');
	let fDescription = $state('');
	let isSaving = $state(false);
	let formError = $state('');
	let listError = $state('');

	const netTotal = $derived(allRows.reduce((sum, r) => sum + r.amount, 0));

	function currentMonth() {
		const d = new Date();
		return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
	}

	function fmtSigned(n: number) {
		const abs = Math.abs(n);
		const s =
			'₱' +
			new Intl.NumberFormat(undefined, {
				minimumFractionDigits: 2,
				maximumFractionDigits: 2
			}).format(abs);
		return n < 0 ? '-' + s : '+' + s;
	}

	function fmtMonth(s: string) {
		const [y, m] = s.split('-');
		return new Date(Number(y), Number(m) - 1, 1).toLocaleDateString(undefined, {
			month: 'long',
			year: 'numeric'
		});
	}

	async function loadData() {
		isLoading = true;
		listError = '';
		const res = await fetch('/x-api/adjustments');
		if (res.ok) {
			const { data } = await res.json();
			allRows = data ?? [];
		} else {
			listError = 'Failed to load adjustments.';
		}
		isLoading = false;
		isRefreshing = false;
	}

	function refresh() {
		isRefreshing = true;
		loadData();
	}

	function openAdd() {
		editingId = null;
		fMonth = currentMonth();
		fAmount = '';
		fDescription = '';
		formError = '';
		showModal = true;
	}

	function openEdit(a: Adjustment) {
		editingId = a.id;
		fMonth = a.month;
		fAmount = a.amount;
		fDescription = a.description;
		formError = '';
		showModal = true;
	}

	function closeModal() {
		showModal = false;
	}

	async function save() {
		formError = '';
		if (fAmount === '' || Number(fAmount) === 0) {
			formError = 'Amount must be a non-zero value (negative to reduce sales).';
			return;
		}
		if (!fMonth) {
			formError = 'Month is required.';
			return;
		}
		isSaving = true;
		try {
			const url = editingId ? `/x-api/adjustments/${editingId}` : '/x-api/adjustments';
			const method = editingId ? 'PUT' : 'POST';
			const res = await fetch(url, {
				method,
				headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
				body: JSON.stringify({
					month: fMonth,
					amount: Number(fAmount),
					description: fDescription
				})
			});
			if (!res.ok) {
				const body = await res.json().catch(() => ({}));
				formError = body.detail ?? 'Failed to save adjustment.';
				return;
			}
			showModal = false;
			loadData();
		} catch {
			formError = 'An unexpected error occurred.';
		} finally {
			isSaving = false;
		}
	}

	async function remove(a: Adjustment) {
		if (!confirm(`Delete the ${fmtSigned(a.amount)} adjustment for ${fmtMonth(a.month)}?`)) return;
		const res = await fetch(`/x-api/adjustments/${a.id}`, { method: 'DELETE' });
		if (res.ok) {
			loadData();
		} else {
			listError = 'Failed to delete adjustment.';
		}
	}

	onMount(loadData);
</script>

{#if listError}
	<div class="uk-alert uk-alert-danger uk-margin-small-bottom" role="alert">{listError}</div>
{/if}

<div class="card">
	<div class="card-header">
		<div class="header-left">
			<span class="card-title">Sales Adjustments</span>
			<button
				class="refresh-btn"
				onclick={refresh}
				disabled={isRefreshing}
				aria-label="Refresh"
				title="Refresh"
			>
				<img src={refreshIcon} class:spinning={isRefreshing} alt="" />
			</button>
		</div>
		<div class="header-right">
			<button class="add-btn" onclick={openAdd} title="Add adjustment">
				<span uk-icon="icon: plus; ratio: 0.85"></span> Add
			</button>
		</div>
	</div>

	<p class="hint-bar">
		Use a positive amount when the device under-counted coins, or a negative amount when it
		over-counted. Adjustments are applied to the chosen month's revenue in the profit report.
	</p>

	<div class="table-wrap">
		<table class="data-table">
			<thead>
				<tr>
					<th>Month</th>
					<th>Description</th>
					<th class="ta-right">Amount</th>
					<th class="ta-right"></th>
				</tr>
			</thead>
			<tbody>
				{#if isLoading}
					{#each { length: 4 } as _}
						<tr>
							{#each { length: 4 } as _}
								<td><div class="skeleton-cell"></div></td>
							{/each}
						</tr>
					{/each}
				{:else if allRows.length === 0}
					<tr><td colspan="4" class="empty-cell">No adjustments recorded yet.</td></tr>
				{:else}
					{#each allRows as row}
						<tr>
							<td class="month">{fmtMonth(row.month)}</td>
							<td class="desc">{row.description || '—'}</td>
							<td class="amount ta-right" class:pos={row.amount > 0} class:neg={row.amount < 0}>
								{fmtSigned(row.amount)}
							</td>
							<td class="ta-right">
								<div class="action-btns">
									<button
										class="icon-btn"
										uk-icon="pencil"
										title="Edit"
										aria-label="Edit"
										onclick={() => openEdit(row)}
									></button>
									<button
										class="icon-btn danger"
										uk-icon="trash"
										title="Delete"
										aria-label="Delete"
										onclick={() => remove(row)}
									></button>
								</div>
							</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>

	<div class="card-footer">
		<span class="total-label">Net Adjustment ({allRows.length})</span>
		<span class="total-amount" class:pos={netTotal > 0} class:neg={netTotal < 0}
			>{fmtSigned(netTotal)}</span
		>
	</div>
</div>

{#if showModal}
	<div
		class="modal-backdrop"
		role="button"
		tabindex="-1"
		onclick={closeModal}
		onkeydown={(e) => e.key === 'Escape' && closeModal()}
	>
		<div
			class="modal"
			role="dialog"
			tabindex="-1"
			onclick={(e) => e.stopPropagation()}
			onkeydown={() => {}}
		>
			<div class="modal-title">{editingId ? 'Edit Adjustment' : 'Add Adjustment'}</div>
			<form
				onsubmit={(e) => {
					e.preventDefault();
					save();
				}}
			>
				{#if formError}
					<div class="uk-alert uk-alert-danger">{formError}</div>
				{/if}

				<div class="field-row">
					<div class="field">
						<label class="uk-form-label" for="adj-month">Month</label>
						<input id="adj-month" class="uk-input" type="month" bind:value={fMonth} required />
					</div>
					<div class="field">
						<label class="uk-form-label" for="adj-amount">Amount (₱)</label>
						<input
							id="adj-amount"
							class="uk-input"
							type="number"
							step="0.01"
							bind:value={fAmount}
							placeholder="e.g. 25 or -25"
							required
						/>
					</div>
				</div>

				<div class="field">
					<label class="uk-form-label" for="adj-desc">Reason</label>
					<input
						id="adj-desc"
						class="uk-input"
						type="text"
						bind:value={fDescription}
						placeholder="e.g. device miscounted coins"
					/>
				</div>

				<div class="modal-actions">
					<button type="button" class="uk-button uk-button-default" onclick={closeModal}
						>Cancel</button
					>
					<button type="submit" class="uk-button uk-button-primary" disabled={isSaving}>
						{isSaving ? 'Saving…' : 'Save'}
					</button>
				</div>
			</form>
		</div>
	</div>
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
		justify-content: space-between;
		gap: 12px;
		padding: 14px 18px;
		border-bottom: 1px solid #f0f0f0;
		flex-wrap: wrap;
	}
	.header-left {
		display: flex;
		align-items: center;
		gap: 6px;
		flex: 1;
	}
	.card-title {
		font-size: 0.88rem;
		font-weight: 700;
		color: #1a1a1a;
		margin-right: auto;
	}
	.refresh-btn {
		display: flex;
		align-items: center;
		background: none;
		border: none;
		padding: 4px;
		border-radius: 6px;
		cursor: pointer;
		color: #999;
		transition: background 0.15s;
	}
	.refresh-btn:hover:not(:disabled) {
		background: #f5f5f5;
	}
	.refresh-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}
	.add-btn {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		padding: 6px 12px;
		border: none;
		border-radius: 7px;
		font-size: 0.8rem;
		font-weight: 600;
		background: var(--color-theme-1, #2563eb);
		color: #fff;
		cursor: pointer;
	}
	.add-btn:hover {
		filter: brightness(0.95);
	}

	.hint-bar {
		margin: 0;
		padding: 10px 18px;
		font-size: 0.78rem;
		color: #777;
		background: #fafafa;
		border-bottom: 1px solid #f0f0f0;
	}

	.table-wrap {
		overflow-x: auto;
	}
	.data-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.85rem;
	}
	.data-table th {
		padding: 9px 16px;
		text-align: left;
		font-size: 0.72rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: #aaa;
		background: #fafafa;
		white-space: nowrap;
	}
	.data-table td {
		padding: 10px 16px;
		border-bottom: 1px solid #f6f6f6;
		vertical-align: middle;
	}
	.data-table tbody tr:last-child td {
		border-bottom: none;
	}
	.data-table tbody tr:hover td {
		background: #fafafa;
	}
	.ta-right {
		text-align: right;
	}
	.month {
		color: #333;
		font-weight: 600;
		white-space: nowrap;
	}
	.desc {
		color: #333;
	}
	.amount {
		font-weight: 700;
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
	}
	.pos {
		color: #059669 !important;
	}
	.neg {
		color: #dc2626 !important;
	}

	.action-btns {
		display: inline-flex;
		gap: 4px;
	}
	.icon-btn {
		background: none;
		border: none;
		padding: 4px;
		border-radius: 6px;
		cursor: pointer;
		color: #999;
	}
	.icon-btn:hover {
		background: #f0f0f0;
		color: #333;
	}
	.icon-btn.danger:hover {
		background: #fee2e2;
		color: #dc2626;
	}

	.empty-cell {
		text-align: center;
		color: #bbb;
		font-size: 0.82rem;
		font-style: italic;
		padding: 32px 16px !important;
	}

	.card-footer {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 6px;
		padding: 10px 18px;
		border-top: 1px solid #f0f0f0;
		background: #fafafa;
	}
	.total-label {
		font-size: 0.75rem;
		font-weight: 600;
		color: #bbb;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}
	.total-amount {
		font-size: 0.9rem;
		font-weight: 700;
		color: #1a1a1a;
		font-variant-numeric: tabular-nums;
	}

	.spinning {
		animation: spin 0.7s linear infinite;
	}
	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
	.skeleton-cell {
		height: 14px;
		border-radius: 4px;
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

	/* Modal */
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
	.modal {
		background: #fff;
		border-radius: 12px;
		padding: 22px;
		width: 100%;
		max-width: 440px;
		box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
	}
	.modal-title {
		font-size: 1.05rem;
		font-weight: 700;
		color: #1a1a1a;
		margin-bottom: 16px;
	}
	.field {
		margin-bottom: 12px;
	}
	.field-row {
		display: flex;
		gap: 12px;
	}
	.field-row .field {
		flex: 1;
	}
	.uk-form-label {
		display: block;
		font-size: 0.75rem;
		font-weight: 600;
		color: #666;
		margin-bottom: 4px;
	}
	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 8px;
		margin-top: 18px;
	}
</style>
