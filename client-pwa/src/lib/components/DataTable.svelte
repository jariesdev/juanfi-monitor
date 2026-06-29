<script lang="ts">
	import {browser} from '$app/environment';
	import {onDestroy, onMount, type Snippet} from 'svelte';
	import debounce from 'lodash/debounce';
	import get from 'lodash/get';
	import type {Filter, RowItem, TableHeader} from '$lib/types/datatable';
	import refreshIcon from '$lib/icons/refresh.svg'
	import {countupInt} from '$lib/utils/countup';

	// props
	interface Props {
		url: string
		headers: TableHeader[]
		filters: Filter
		title: string
		perPage?: number
		clientSort?: boolean   // sort loaded items in memory instead of via query params

		// snippets
		titleActions?: Snippet
		beforeTable?: Snippet
		afterTable?: Snippet
		empty?: Snippet
		row?: Snippet<[RowItem]>
		cell?: Snippet<[RowItem, TableHeader, Function]>
	}

	const {url, headers = [], filters = {}, title = 'Table Records', perPage = 15, clientSort = false, titleActions, beforeTable, afterTable, row, cell, empty}: Props = $props()

	// states
	let isRefreshing: boolean = $state(false);
	let currentPage: number = $state(1);
	let maxPage: number = $state(1);
	let totalItems: number = $state(1);
	let isLoading: boolean = $state(true);
	let tableItems: RowItem[] = $state([]);
	let searchInput: string = $state('');
	let sortField: string = $state('');
	let sortDir: 'asc' | 'desc' = $state('asc');
	let controller: AbortController | undefined = undefined;
	let infiniteScrollEl: HTMLDivElement;

	function toggleSort(header: TableHeader) {
		if (!header.sortable) return;
		const key = header.sortKey ?? header.field;
		if (sortField === key) {
			sortDir = sortDir === 'asc' ? 'desc' : 'asc';
		} else {
			sortField = key;
			sortDir = 'asc';
		}
		if (!clientSort) {
			currentPage = 1;
			tableItems = [];
		}
	}

	// client-side sorted view (used only when clientSort=true)
	const displayItems: RowItem[] = $derived.by(() => {
		if (!clientSort || !sortField) return tableItems;
		const dir = sortDir === 'asc' ? 1 : -1;
		return [...tableItems].sort((a, b) => {
			const av = get(a, sortField, '');
			const bv = get(b, sortField, '');
			if (typeof av === 'number' && typeof bv === 'number') return (av - bv) * dir;
			return String(av).localeCompare(String(bv)) * dir;
		});
	});

	const queryParams: Filter = $derived.by(() => {
		let params: Filter = {};
		// remove empty params
		Object.keys(filters).forEach((k) => {
			if (filters[k] != undefined && filters[k] != null && filters[k] != '') {
				params[k] = filters[k];
			}
		});

		const base: Filter = {
			...params,
			q: searchInput,
			page: currentPage,
			size: perPage
		};

		// include sort params for server-side sort only
		if (!clientSort && sortField) {
			base['sort_by']  = sortField;
			base['sort_dir'] = sortDir;
		}

		return base;
	})

	const isFirstLoad: boolean = $derived(currentPage == 1 && tableItems.length === 0)

	// load table data
	export const loadData: Function = debounce(
		async (): Promise<void> => {
			let localUrl = url;
			isLoading = true;

			if (Object.keys(queryParams).length > 0) {
				localUrl =
					localUrl +
					'?' +
					Object.keys(queryParams)
						.map((k: string) => `${k}=${queryParams[k]}`)
						.join('&');
			}
			controller = new AbortController();
			const signal = controller.signal;
			const request = new Request(localUrl, {method: 'GET', signal: signal});

			fetch(request)
				.then((response) => {
					if (response.status === 200) {
						return response.json();
					} else {
						throw new Error('Something went wrong on API server!');
					}
				})
				.then((response) => {
					// fall back to a flat `{ data: [...] }` shape for non-paginated endpoints
					const items = response.items || response.data || [];
					totalItems = response.total ?? items.length;
					maxPage = response.pages || 1;
					tableItems = [...tableItems, ...items];
				})
				.catch((error) => {
					console.error(error);
					tableItems = [];
				})
				.finally(() => {
					isLoading = false;
					isRefreshing = false;
				});
		},
		250,
		{maxWait: 1000}
	);

	function handleRefresh() {
		currentPage = 1;
		tableItems = [];
		isRefreshing = true;
		loadData();
	}

	function getCellValue(item: RowItem, header: TableHeader) {
		return get(item, header.field, '');
	}

	// reset when search
	const resetTableQuery = () => {
		currentPage = 1;
		tableItems = [];
	}
	$effect(() => {
		filters // allow to watch for changes
		resetTableQuery()
	})
	$effect(() => {
		searchInput // allow to watch for changes
		resetTableQuery()
	})
	$effect(() => {
		queryParams // allow to watch for changes
		loadData()
	})

	// watcher infinite scroll
	const initializeInfiniteScroll = () => {
		if (browser) {
			let options = {
				rootMargin: '-10px',
				threshold: 0
			};
			let callback = (entries: any[]) => {
				entries.forEach((e) => {
					if (currentPage < maxPage && !isLoading && e.isIntersecting && !isFirstLoad) {
						currentPage += 1;
						isLoading = true;
						console.log('infinit scroll')
						loadData();
					}
				});
			};

			let observer = new IntersectionObserver(callback, options);
			observer.observe(infiniteScrollEl);
		}
	}

	onMount(() => {
		// load data from url
		loadData();
		initializeInfiniteScroll()
	});

	onDestroy(() => {
		// on component destroy, cancel ongoing HTTP request
		// controller && controller.abort('component destroyed');
	});
</script>

{#snippet rowFallback(item: RowItem)}

	{#snippet cellFallback(item: RowItem, header: TableHeader, getCellValue: Function)}
		{getCellValue(item, header)}
	{/snippet}

	<tr>
		{#each headers as header}
			<td>
				{@render (cell || cellFallback)(item, header, getCellValue)}
			</td>
		{/each}
	</tr>
{/snippet}

{#snippet emptyFallback()}
	<span>No record yet.</span>
{/snippet}

<div class="card">
	<!-- Card header -->
	<div class="card-header">
		<div class="header-left">
			<span class="card-title">{title}</span>
			<button
				class="refresh-btn"
				disabled={isRefreshing}
				onclick={handleRefresh}
				title="Refresh"
				aria-label="Refresh"
			>
				<img src="{refreshIcon}" class:spinning={isRefreshing} alt="" />
			</button>
			{@render titleActions?.()}
		</div>
		<div class="header-right">
			<div class="search-wrap">
				<span class="search-icon" uk-icon="icon: search; ratio: 0.8"></span>
				<input
					bind:value={searchInput}
					class="search-input"
					type="search"
					placeholder="Search…"
					aria-label="Search"
				/>
			</div>
		</div>
	</div>

	<!-- Filters -->
	{#if beforeTable}
		<div class="filters-row">
			{@render beforeTable()}
		</div>
	{/if}

	<!-- Table -->
	<div class="table-wrap">
		<table class="data-table">
			<thead>
				<tr>
					{#each headers as header}
						{@const key = header.sortKey ?? header.field}
						{@const isActive = sortField === key}
						<th
							class:sortable={header.sortable}
							onclick={() => toggleSort(header)}
						>
							{header.label}
							{#if header.sortable}
								<span
									class="sort-icon"
									class:active={isActive}
									uk-icon="icon: {isActive && sortDir === 'desc' ? 'chevron-down' : 'chevron-up'}; ratio: 0.75"
								></span>
							{/if}
						</th>
					{/each}
				</tr>
			</thead>
			<tbody>
				{#if isLoading && isFirstLoad}
					{#each { length: 5 } as _}
						<tr>
							{#each headers as _}
								<td><div class="skeleton-cell"></div></td>
							{/each}
						</tr>
					{/each}
				{:else if tableItems.length === 0}
					<tr>
						<td colspan="99" class="empty-cell">
							{@render (empty || emptyFallback)()}
						</td>
					</tr>
				{:else}
					{#each (clientSort ? displayItems : tableItems) as item}
						{@render (row || rowFallback)(item)}
					{/each}
				{/if}
			</tbody>
		</table>
	</div>

	{@render afterTable?.()}

	<!-- Footer -->
	<div class="card-footer">
		<span class="total-label">Total</span>
		<span class="total-count" use:countupInt={totalItems}></span>
	</div>

	<div bind:this={infiniteScrollEl}></div>
</div>

<style>
	/* Card */
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
		min-width: 0;
	}

	.header-right {
		flex-shrink: 0;
	}

	.card-title {
		font-size: 0.88rem;
		font-weight: 700;
		color: #1a1a1a;
		white-space: nowrap;
	}

	.refresh-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		background: none;
		border: none;
		padding: 4px;
		border-radius: 6px;
		cursor: pointer;
		color: #999;
		transition: background 0.15s;
		flex-shrink: 0;
	}

	.refresh-btn:hover:not(:disabled) {
		background: #f5f5f5;
		color: #333;
	}

	.refresh-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	/* Search */
	.search-wrap {
		position: relative;
		display: flex;
		align-items: center;
	}

	.search-icon {
		position: absolute;
		left: 8px;
		color: #bbb;
		pointer-events: none;
		display: flex;
		align-items: center;
	}

	.search-input {
		padding: 6px 10px 6px 28px;
		border: 1px solid #e8e8e8;
		border-radius: 7px;
		font-size: 0.8rem;
		font-family: inherit;
		color: #333;
		background: #fafafa;
		width: 180px;
		outline: none;
		transition: border-color 0.15s, background 0.15s;
	}

	.search-input:focus {
		border-color: #ccc;
		background: #fff;
	}

	.search-input::placeholder {
		color: #ccc;
	}

	/* Filters slot */
	.filters-row {
		padding: 10px 18px;
		border-bottom: 1px solid #f6f6f6;
		background: #fafafa;
	}

	/* Table */
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
		white-space: nowrap;
		background: #fafafa;
		user-select: none;
	}

	.data-table th.sortable {
		cursor: pointer;
	}

	.data-table th.sortable:hover {
		color: #555;
	}

	.sort-icon {
		display: inline-flex;
		vertical-align: middle;
		margin-left: 2px;
		opacity: 0.25;
	}

	.sort-icon.active {
		opacity: 1;
		color: var(--color-theme-1);
	}

	.data-table td {
		padding: 10px 16px;
		color: #333;
		border-bottom: 1px solid #f6f6f6;
		vertical-align: middle;
	}

	.data-table tbody tr:last-child td {
		border-bottom: none;
	}

	.data-table tbody tr:hover td {
		background: #fafafa;
	}

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
	.spinning {
		animation: spin 0.7s linear infinite;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	.skeleton-cell {
		height: 14px;
		border-radius: 4px;
		background: linear-gradient(90deg, #f0f0f0 25%, #e8e8e8 50%, #f0f0f0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.4s infinite;
	}

	@keyframes shimmer {
		0% { background-position: 200% 0; }
		100% { background-position: -200% 0; }
	}
</style>
