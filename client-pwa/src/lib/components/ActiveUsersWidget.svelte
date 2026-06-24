<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { CountUp } from 'countup.js';
	import type { iVendo } from '$lib/types/models';
	import type { RowItem, SimpleTableHeader } from '$lib/types/datatable';
	import SimpleTable from '$lib/components/SimpleTable.svelte';

	const MAX_VISIBLE = 5;

	const modalHeaders: SimpleTableHeader[] = [
		{ label: 'Name', field: 'name' },
		{ label: 'Active Users', field: 'active_users', align: 'center' },
		{ label: 'Current Sales', field: 'current_sales', align: 'right' }
	];

	let vendos: iVendo[] = $state([]);
	let isLoading = $state(true);
	let showModal = $state(false);
	let controller: AbortController | undefined;
	let intervalId: ReturnType<typeof setInterval>;

	function rawUsers(v: iVendo): number {
		return v.recent_status?.active_users ?? v.active_users ?? 0;
	}

	function rawSales(v: iVendo): number {
		return v.recent_status?.current_sales ?? v.current_sales ?? 0;
	}

	function fmt(n: number): string {
		return new Intl.NumberFormat().format(Math.round(n));
	}

	// flattened rows for the "All Vendo Machines" table — sortable by plain field name
	const modalRows: RowItem[] = $derived(
		vendos.map((v) => ({
			id: v.id,
			name: v.name,
			active_users: rawUsers(v),
			current_sales: rawSales(v)
		}))
	);

	// Svelte action: animate an integer counter on mount; update() animates to new value
	function countupInt(node: HTMLElement, value: number) {
		const counter = new CountUp(node, value, { duration: 0.9, useEasing: true });
		counter.start();
		return {
			update(newValue: number) {
				counter.update(newValue);
			}
		};
	}

	// Svelte action: animate a peso sales counter with ₱ prefix and comma separator
	function countupSales(node: HTMLElement, value: number) {
		const counter = new CountUp(node, value, {
			duration: 0.9,
			useEasing: true,
			prefix: '₱',
			separator: ',',
			decimalPlaces: 0
		});
		counter.start();
		return {
			update(newValue: number) {
				counter.update(newValue);
			}
		};
	}

	function loadData() {
		controller?.abort();
		controller = new AbortController();

		fetch('/x-api/vendo-machines', { signal: controller.signal })
			.then((r) => (r.ok ? r.json() : Promise.reject()))
			.then(({ data }) => {
				// highest active users first, left to right
				vendos = (data ?? [])
					.filter((v: iVendo) => v.is_active)
					.sort((a: iVendo, b: iVendo) => rawUsers(b) - rawUsers(a));
			})
			.catch(() => {})
			.finally(() => {
				isLoading = false;
			});
	}

	onMount(() => {
		loadData();
		intervalId = setInterval(loadData, 30_000);
	});

	onDestroy(() => {
		controller?.abort();
		clearInterval(intervalId);
	});
</script>

<!-- Modal — all vendos table -->
{#if showModal}
	<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
	<div class="modal-backdrop" onclick={() => (showModal = false)}></div>
	<div class="modal" role="dialog" aria-label="All Vendo Machines">
		<div class="modal-header">
			<span class="modal-title">All Vendo Machines</span>
			<button class="modal-close" onclick={() => (showModal = false)} aria-label="Close">
				<span uk-icon="icon: close"></span>
			</button>
		</div>
		<div class="modal-body">
			<SimpleTable headers={modalHeaders} items={modalRows}>
				{#snippet cell(item: RowItem, header: SimpleTableHeader, getCellValue: Function)}
					{#if header.field === 'active_users'}
						<strong>{getCellValue(item, header)}</strong>
					{:else if header.field === 'current_sales'}
						<span class="text-muted">₱{fmt(getCellValue(item, header))}</span>
					{:else}
						{getCellValue(item, header)}
					{/if}
				{/snippet}
			</SimpleTable>
		</div>
	</div>
{/if}

<!-- Widget bar -->
<div class="widget">
	{#if isLoading}
		<div class="loading-row">
			{#each { length: 5 } as _}
				<div class="vendo-card skeleton"></div>
			{/each}
		</div>
	{:else}

		{#each vendos.slice(0, MAX_VISIBLE) as v (v.id)}
			<div class="vendo-card">
				<div class="vendo-name">{v.name}</div>
				<div class="vendo-users" use:countupInt={rawUsers(v)}></div>
				<div class="vendo-sales" use:countupSales={rawSales(v)}></div>
			</div>
		{/each}

		<!-- "Show all" button — only when vendos exceed MAX_VISIBLE -->
		{#if vendos.length > MAX_VISIBLE}
			<button class="show-all-btn" onclick={() => (showModal = true)} title="Show all vendo machines">
				<span uk-icon="icon: table"></span>
				<span class="show-all-count">+{vendos.length - MAX_VISIBLE}</span>
			</button>
		{/if}
	{/if}
</div>

<style>
	/* Widget container */
	.widget {
		display: flex;
		align-items: stretch;
		gap: 8px;
		width: 100%;
		padding: 10px 0 8px;
	}

	/* Vendo card */
	.vendo-card {
		flex: 1;
		min-width: 0;
		background: #fff;
		border: 1px solid #ebebeb;
		border-radius: 8px;
		padding: 10px 14px;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.vendo-name {
		font-size: 0.72rem;
		color: #aaa;
		font-weight: 500;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}

	.vendo-users {
		font-size: 1.6rem;
		font-weight: 700;
		color: #111;
		line-height: 1.1;
	}

	.vendo-sales {
		font-size: 0.78rem;
		color: #bbb;
		font-weight: 400;
	}

	/* "Show all" button — left-most */
	.show-all-btn {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 2px;
		background: #fff;
		border: 1px solid #ebebeb;
		border-radius: 8px;
		padding: 10px 12px;
		cursor: pointer;
		color: #888;
		min-width: 52px;
		transition: background 0.15s, color 0.15s;
		flex-shrink: 0;
	}

	.show-all-btn:hover {
		background: #f5f5f5;
		color: #333;
	}

	.show-all-count {
		font-size: 0.7rem;
		font-weight: 600;
		color: var(--color-theme-1);
	}

	/* Skeleton loader */
	.loading-row {
		display: flex;
		gap: 8px;
		width: 100%;
	}

	.skeleton {
		flex: 1;
		height: 72px;
		background: linear-gradient(90deg, #f0f0f0 25%, #e8e8e8 50%, #f0f0f0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.4s infinite;
		border: none;
	}

	@keyframes shimmer {
		0% { background-position: 200% 0; }
		100% { background-position: -200% 0; }
	}

	/* Modal */
	.modal-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.4);
		z-index: 500;
	}

	.modal {
		position: fixed;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		background: #fff;
		border-radius: 10px;
		z-index: 501;
		width: min(600px, 92vw);
		max-height: 80vh;
		display: flex;
		flex-direction: column;
		box-shadow: 0 8px 40px rgba(0, 0, 0, 0.18);
	}

	.modal-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 16px 20px 14px;
		border-bottom: 1px solid #f0f0f0;
		flex-shrink: 0;
	}

	.modal-title {
		font-size: 0.95rem;
		font-weight: 700;
		color: #222;
	}

	.modal-close {
		background: none;
		border: none;
		cursor: pointer;
		color: #999;
		padding: 4px;
		border-radius: 5px;
		display: flex;
		align-items: center;
	}

	.modal-close:hover {
		background: #f5f5f5;
		color: #333;
	}

	.modal-body {
		overflow-y: auto;
		padding: 0 8px 8px;
	}

	.text-muted {
		color: #aaa;
	}
</style>
