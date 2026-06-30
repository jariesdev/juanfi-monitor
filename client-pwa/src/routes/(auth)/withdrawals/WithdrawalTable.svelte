<script lang="ts">
	import { onMount } from 'svelte';
	import refreshIcon from '$lib/icons/refresh.svg';

	interface Vendo { id: number; name: string; }
	interface Withdrawal { id: number; vendo: Vendo | null; amount: number; created_at: string; }

	type SortField = 'vendo' | 'amount' | 'created_at';
	type SortDir   = 'asc' | 'desc';

	let allRows:    Withdrawal[] = $state([]);
	let vendos:     Vendo[]      = $state([]);
	let isLoading   = $state(true);
	let isRefreshing = $state(false);
	let vendoId: number | '' = $state('');
	let sortField: SortField = $state('created_at');
	let sortDir:   SortDir   = $state('desc');

	// filtered + sorted view
	const rows = $derived.by(() => {
		let list = vendoId
			? allRows.filter(r => r.vendo?.id === vendoId)
			: allRows;

		return [...list].sort((a, b) => {
			let av: string | number, bv: string | number;
			if (sortField === 'vendo') {
				av = a.vendo?.name ?? '';
				bv = b.vendo?.name ?? '';
			} else {
				av = a[sortField];
				bv = b[sortField];
			}
			if (typeof av === 'number' && typeof bv === 'number') {
				return sortDir === 'asc' ? av - bv : bv - av;
			}
			return sortDir === 'asc'
				? String(av).localeCompare(String(bv))
				: String(bv).localeCompare(String(av));
		});
	});

	function sort(field: SortField) {
		if (sortField === field) {
			sortDir = sortDir === 'asc' ? 'desc' : 'asc';
		} else {
			sortField = field;
			sortDir   = 'asc';
		}
	}

	function sortIcon(field: SortField): string {
		if (sortField !== field) return 'chevron-up';
		return sortDir === 'asc' ? 'chevron-up' : 'chevron-down';
	}

	function fmt(n: number) {
		return '₱' + new Intl.NumberFormat().format(n);
	}

	function fmtDate(s: string) {
		return new Date(s).toLocaleString(undefined, {
			month: 'short', day: 'numeric', year: 'numeric',
			hour: '2-digit', minute: '2-digit'
		});
	}

	async function loadVendos() {
		const res = await fetch('/x-api/vendo-machines');
		if (res.ok) {
			const { data } = await res.json();
			vendos = (data ?? []).sort((a: Vendo, b: Vendo) => a.name.localeCompare(b.name));
		}
	}

	async function loadData() {
		isLoading = true;
		const res = await fetch('/x-api/withdrawals');
		if (res.ok) {
			const { data } = await res.json();
			allRows = data ?? [];
		}
		isLoading = false;
		isRefreshing = false;
	}

	function refresh() {
		isRefreshing = true;
		loadData();
	}

	onMount(() => {
		loadVendos();
		loadData();
	});
</script>

<div class="card">
	<!-- Header -->
	<div class="card-header">
		<div class="header-left">
			<span class="card-title">Withdrawals</span>
			<button class="refresh-btn" onclick={refresh} disabled={isRefreshing} aria-label="Refresh" title="Refresh">
				<img src={refreshIcon} class:spinning={isRefreshing} alt="" />
			</button>
		</div>
		<div class="header-right">
			<select class="vendo-filter" bind:value={vendoId} aria-label="Filter by vendo">
				<option value="">All Vendos</option>
				{#each vendos as v}
					<option value={v.id}>{v.name}</option>
				{/each}
			</select>
		</div>
	</div>

	<!-- Table -->
	<div class="table-wrap">
		<table class="data-table">
			<thead>
				<tr>
					<th class="sortable" onclick={() => sort('vendo')}>
						Vendo
						<span class="sort-icon" class:active={sortField === 'vendo'}
							uk-icon="icon: {sortIcon('vendo')}; ratio: 0.75"></span>
					</th>
					<th class="sortable" onclick={() => sort('amount')}>
						Amount
						<span class="sort-icon" class:active={sortField === 'amount'}
							uk-icon="icon: {sortIcon('amount')}; ratio: 0.75"></span>
					</th>
					<th class="sortable" onclick={() => sort('created_at')}>
						Date
						<span class="sort-icon" class:active={sortField === 'created_at'}
							uk-icon="icon: {sortIcon('created_at')}; ratio: 0.75"></span>
					</th>
				</tr>
			</thead>
			<tbody>
				{#if isLoading}
					{#each { length: 5 } as _}
						<tr>
							{#each { length: 3 } as _}
								<td><div class="skeleton-cell"></div></td>
							{/each}
						</tr>
					{/each}
				{:else if rows.length === 0}
					<tr>
						<td colspan="3" class="empty-cell">No withdrawal records found.</td>
					</tr>
				{:else}
					{#each rows as row}
						<tr>
							<td class="vendo-name">{row.vendo?.name ?? '—'}</td>
							<td class="amount">{fmt(row.amount)}</td>
							<td class="date">{fmtDate(row.created_at)}</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>

	<!-- Footer -->
	<div class="card-footer">
		<span class="total-label">Total</span>
		<span class="total-count">{rows.length}</span>
	</div>
</div>

<style>
	.card {
		background: #fff;
		border: 1px solid #e8e8e8;
		border-radius: 10px;
		overflow: hidden;
	}

	/* Header */
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

	.refresh-btn:hover:not(:disabled) { background: #f5f5f5; }
	.refresh-btn:disabled { opacity: 0.4; cursor: not-allowed; }

	/* Vendo filter */
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
		transition: border-color 0.15s;
	}

	.vendo-filter:focus { border-color: #ccc; background: #fff; }

	/* Table */
	.table-wrap { overflow-x: auto; }

	.data-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.85rem;
	}

	.data-table thead tr { border-bottom: 1px solid #f0f0f0; }

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
		user-select: none;
	}

	.data-table th.sortable { cursor: pointer; }
	.data-table th.sortable:hover { color: #555; }

	.sort-icon {
		display: inline-flex;
		vertical-align: middle;
		opacity: 0.25;
		margin-left: 2px;
	}
	.sort-icon.active { opacity: 1; color: var(--color-theme-1); }

	.data-table td {
		padding: 10px 16px;
		border-bottom: 1px solid #f6f6f6;
		vertical-align: middle;
	}

	.data-table tbody tr:last-child td { border-bottom: none; }
	.data-table tbody tr:hover td { background: #fafafa; }

	.vendo-name { color: #333; font-weight: 500; }
	.amount     { color: #1a1a1a; font-weight: 600; font-variant-numeric: tabular-nums; }
	.date       { color: #888; font-size: 0.8rem; white-space: nowrap; }

	.empty-cell {
		text-align: center;
		color: #bbb;
		font-size: 0.82rem;
		font-style: italic;
		padding: 32px 16px !important;
	}

	/* Footer */
	.card-footer {
		display: flex;
		align-items: center;
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

	.total-count {
		font-size: 0.82rem;
		font-weight: 700;
		color: #666;
	}

	/* Animations */
	.spinning { animation: spin 0.7s linear infinite; }
	@keyframes spin { to { transform: rotate(360deg); } }

	.skeleton-cell {
		height: 14px;
		border-radius: 4px;
		background: linear-gradient(90deg, #f0f0f0 25%, #e8e8e8 50%, #f0f0f0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.4s infinite;
	}
	@keyframes shimmer {
		0%   { background-position: 200% 0; }
		100% { background-position: -200% 0; }
	}
</style>
