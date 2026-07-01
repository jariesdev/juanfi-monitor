<script lang="ts">
	import { onMount } from 'svelte';
	import refreshIcon from '$lib/icons/refresh.svg';

	interface Expense {
		id: number;
		category: string;
		description: string;
		amount: number;
		is_recurring: boolean;
		expense_date: string;
		end_date: string | null;
	}

	const CATEGORIES = [
		{ value: 'subscription', label: 'Subscription' },
		{ value: 'electricity', label: 'Electricity' },
		{ value: 'materials', label: 'Materials / Equipment' },
		{ value: 'consumables', label: 'Consumables' },
		{ value: 'labor', label: 'Labor' },
		{ value: 'other', label: 'Other' }
	];
	const CATEGORY_LABELS: Record<string, string> = Object.fromEntries(
		CATEGORIES.map((c) => [c.value, c.label])
	);

	let allRows: Expense[] = $state([]);
	let isLoading = $state(true);
	let isRefreshing = $state(false);
	let categoryFilter: string = $state('');

	// modal / form state
	let showModal = $state(false);
	let editingId: number | null = $state(null);
	let fCategory = $state('materials');
	let fDescription = $state('');
	let fAmount: number | '' = $state('');
	let fDate = $state(todayISO());
	let fRecurring = $state(false);
	let fEndDate = $state('');
	let isSaving = $state(false);
	let formError = $state('');
	let listError = $state('');

	const rows = $derived(
		categoryFilter ? allRows.filter((r) => r.category === categoryFilter) : allRows
	);

	const totalAmount = $derived(rows.reduce((sum, r) => sum + r.amount, 0));

	function todayISO() {
		const d = new Date();
		return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
	}

	function fmt(n: number) {
		return (
			'₱' +
			new Intl.NumberFormat(undefined, {
				minimumFractionDigits: 2,
				maximumFractionDigits: 2
			}).format(n)
		);
	}

	function fmtDate(s: string) {
		return new Date(s).toLocaleDateString(undefined, {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function fmtMonthShort(s: string) {
		return new Date(s).toLocaleDateString(undefined, { month: 'short', year: 'numeric' });
	}

	async function loadData() {
		isLoading = true;
		listError = '';
		const res = await fetch('/x-api/expenses');
		if (res.ok) {
			const { data } = await res.json();
			allRows = data ?? [];
		} else {
			listError = 'Failed to load expenses.';
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
		fCategory = 'materials';
		fDescription = '';
		fAmount = '';
		fDate = todayISO();
		fRecurring = false;
		fEndDate = '';
		formError = '';
		showModal = true;
	}

	function openEdit(e: Expense) {
		editingId = e.id;
		fCategory = e.category;
		fDescription = e.description;
		fAmount = e.amount;
		fDate = e.expense_date.slice(0, 10);
		fRecurring = e.is_recurring;
		fEndDate = e.end_date ? e.end_date.slice(0, 10) : '';
		formError = '';
		showModal = true;
	}

	function closeModal() {
		showModal = false;
	}

	async function save() {
		formError = '';
		if (fAmount === '' || Number(fAmount) <= 0) {
			formError = 'Amount must be greater than zero.';
			return;
		}
		if (!fDate) {
			formError = 'Expense date is required.';
			return;
		}
		if (fRecurring && fEndDate && fEndDate < fDate) {
			formError = 'End date must be on or after the expense date.';
			return;
		}
		isSaving = true;
		try {
			const url = editingId ? `/x-api/expenses/${editingId}` : '/x-api/expenses';
			const method = editingId ? 'PUT' : 'POST';
			const res = await fetch(url, {
				method,
				headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
				body: JSON.stringify({
					category: fCategory,
					description: fDescription,
					amount: Number(fAmount),
					is_recurring: fRecurring,
					expense_date: fDate,
					end_date: fRecurring ? fEndDate : ''
				})
			});
			if (!res.ok) {
				const body = await res.json().catch(() => ({}));
				formError = body.detail ?? 'Failed to save expense.';
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

	async function remove(e: Expense) {
		if (
			!confirm(
				`Delete this ${CATEGORY_LABELS[e.category] ?? e.category} expense of ${fmt(e.amount)}?`
			)
		)
			return;
		const res = await fetch(`/x-api/expenses/${e.id}`, { method: 'DELETE' });
		if (res.ok) {
			loadData();
		} else {
			listError = 'Failed to delete expense.';
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
			<span class="card-title">Expenses</span>
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
			<select class="vendo-filter" bind:value={categoryFilter} aria-label="Filter by category">
				<option value="">All Categories</option>
				{#each CATEGORIES as c}
					<option value={c.value}>{c.label}</option>
				{/each}
			</select>
			<button class="add-btn" onclick={openAdd} title="Add expense">
				<span uk-icon="icon: plus; ratio: 0.85"></span> Add
			</button>
		</div>
	</div>

	<div class="table-wrap">
		<table class="data-table">
			<thead>
				<tr>
					<th>Date</th>
					<th>Category</th>
					<th>Description</th>
					<th>Type</th>
					<th class="ta-right">Amount</th>
					<th class="ta-right"></th>
				</tr>
			</thead>
			<tbody>
				{#if isLoading}
					{#each { length: 5 } as _}
						<tr>
							{#each { length: 6 } as _}
								<td><div class="skeleton-cell"></div></td>
							{/each}
						</tr>
					{/each}
				{:else if rows.length === 0}
					<tr><td colspan="6" class="empty-cell">No expenses recorded yet.</td></tr>
				{:else}
					{#each rows as row}
						<tr>
							<td class="date">{fmtDate(row.expense_date)}</td>
							<td><span class="cat-badge">{CATEGORY_LABELS[row.category] ?? row.category}</span></td
							>
							<td class="desc">{row.description || '—'}</td>
							<td>
								{#if row.is_recurring}
									<span class="type-badge recurring">Monthly</span>
									{#if row.end_date}
										<span class="until">until {fmtMonthShort(row.end_date)}</span>
									{/if}
								{:else}
									<span class="type-badge onetime">One-time</span>
								{/if}
							</td>
							<td class="amount ta-right">{fmt(row.amount)}</td>
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
		<span class="total-label">Total ({rows.length})</span>
		<span class="total-amount">{fmt(totalAmount)}</span>
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
			<div class="modal-title">{editingId ? 'Edit Expense' : 'Add Expense'}</div>
			<form
				onsubmit={(e) => {
					e.preventDefault();
					save();
				}}
			>
				{#if formError}
					<div class="uk-alert uk-alert-danger">{formError}</div>
				{/if}

				<div class="field">
					<label class="uk-form-label" for="ex-cat">Category</label>
					<select id="ex-cat" class="uk-select" bind:value={fCategory}>
						{#each CATEGORIES as c}
							<option value={c.value}>{c.label}</option>
						{/each}
					</select>
				</div>

				<div class="field">
					<label class="uk-form-label" for="ex-desc">Description</label>
					<input
						id="ex-desc"
						class="uk-input"
						type="text"
						bind:value={fDescription}
						placeholder="e.g. Starlink device, June electric bill"
					/>
				</div>

				<div class="field-row">
					<div class="field">
						<label class="uk-form-label" for="ex-amount">Amount (₱)</label>
						<input
							id="ex-amount"
							class="uk-input"
							type="number"
							min="0"
							step="0.01"
							bind:value={fAmount}
							required
						/>
					</div>
					<div class="field">
						<label class="uk-form-label" for="ex-date">Expense Date</label>
						<input id="ex-date" class="uk-input" type="date" bind:value={fDate} required />
					</div>
				</div>

				<label class="recurring-toggle">
					<input type="checkbox" class="uk-checkbox" bind:checked={fRecurring} />
					<span
						>Repeats monthly <span class="hint"
							>(recurring cost like a subscription or electricity)</span
						></span
					>
				</label>

				{#if fRecurring}
					<div class="field">
						<label class="uk-form-label" for="ex-end"
							>End Date <span class="hint">(optional)</span></label
						>
						<input id="ex-end" class="uk-input" type="date" min={fDate} bind:value={fEndDate} />
						<p class="field-hint">Leave blank if the cost is still ongoing.</p>
					</div>
				{/if}

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
	.header-right {
		display: flex;
		align-items: center;
		gap: 8px;
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

	.vendo-filter {
		padding: 6px 10px;
		border: 1px solid #e8e8e8;
		border-radius: 7px;
		font-size: 0.8rem;
		font-family: inherit;
		color: #333;
		background: #fafafa;
		outline: none;
		cursor: pointer;
	}
	.vendo-filter:focus {
		border-color: #ccc;
		background: #fff;
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

	.table-wrap {
		overflow-x: auto;
	}
	.data-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.85rem;
	}
	.data-table thead tr {
		border-bottom: 1px solid #f0f0f0;
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

	.date {
		color: #888;
		font-size: 0.8rem;
		white-space: nowrap;
	}
	.desc {
		color: #333;
	}
	.amount {
		color: #1a1a1a;
		font-weight: 600;
		font-variant-numeric: tabular-nums;
	}

	.cat-badge {
		display: inline-block;
		padding: 2px 8px;
		border-radius: 20px;
		font-size: 0.72rem;
		font-weight: 600;
		background: #eef2ff;
		color: #4f46e5;
	}
	.type-badge {
		display: inline-block;
		padding: 2px 8px;
		border-radius: 20px;
		font-size: 0.7rem;
		font-weight: 600;
	}
	.type-badge.recurring {
		background: #fef3c7;
		color: #b45309;
	}
	.type-badge.onetime {
		background: #f1f5f9;
		color: #64748b;
	}
	.until {
		margin-left: 6px;
		font-size: 0.72rem;
		color: #999;
		white-space: nowrap;
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
	.recurring-toggle {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 0.82rem;
		color: #333;
		margin: 6px 0 4px;
		cursor: pointer;
	}
	.recurring-toggle .hint {
		color: #999;
		font-weight: 400;
		font-size: 0.76rem;
	}
	.field-hint {
		margin: 4px 0 0;
		font-size: 0.72rem;
		color: #999;
	}
	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 8px;
		margin-top: 18px;
	}
</style>
