<script lang="ts">
	import {browser} from '$app/environment';
	import {onDestroy, onMount, type Snippet} from 'svelte';
	import debounce from 'lodash/debounce';
	import get from 'lodash/get';
	import type {Filter, RowItem, TableHeader} from '$lib/types/datatable';
	import refreshIcon from '$lib/icons/refresh.svg'

	// props
	interface Props {
		url: string
		headers: TableHeader[]
		filters: Filter
		title: string
		perPage?: number

		// snippets
		beforeTable: Snippet|undefined
		afterTable: Snippet|undefined
		empty: Snippet|undefined
		row: Snippet|undefined
		cell: Snippet|undefined
	}

	const {url, headers = [], filters = {}, title = 'Table Records', perPage = 15, beforeTable, afterTable, row, cell, empty}: Props = $props()

	// states
	let isRefreshing: boolean = $state(false);
	let currentPage: number = $state(1);
	let maxPage: number = $state(1);
	let totalItems: number = $state(1);
	let isLoading: boolean = $state(true);
	let tableItems: RowItem[] = $state([]);
	let searchInput: string = $state('');
	let controller: AbortController | undefined = undefined;
	let infiniteScrollEl: HTMLDivElement;

	const queryParams: Filter = $derived.by(() => {
		let params: Filter = {};
		// remove empty params
		Object.keys(filters).forEach((k) => {
			if (filters[k] != undefined && filters[k] != null && filters[k] != '') {
				params[k] = filters[k];
			}
		});

		return {
			...params,
			q: searchInput,
			page: currentPage,
			size: perPage
		}
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
					totalItems = response.total;
					maxPage = response.pages || 1;
					const items = response.items || [];
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

<div class="uk-card uk-card-default uk-card-body">
	<div class="uk-margin-small-top uk-grid uk-grid-small" style="row-gap: 15px;">
		<div class="uk-width-2-3@s uk-flex uk-flex-middle" style="gap: 8px;">
			<h3 class="uk-card-title" style="margin: 0;">{title}</h3>
				<button
					class="uk-icon-button"
					disabled={isRefreshing}
					onclick={handleRefresh}
					title="Refresh"
					style:border="none"
				>
					<img src="{refreshIcon}" class:spinning={isRefreshing}  alt="Refresh" style:margin="3px"/>
				</button>
		</div>
		<div class="uk-width-1-3@s">
			<input
				bind:value={searchInput}
				class="uk-input uk-form-small"
				type="search"
				placeholder="Search"
				aria-label="Input"
			/>
		</div>
	</div>
	<div>
		{@render beforeTable?.()}
	</div>
	<div class="uk-overflow-auto uk-margin-bottom">
		<table class="uk-table uk-table-divider">
			<thead>
			<tr>
				{#each headers as header}
					<th>{header.label}</th>
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
					<td colspan="99" class="uk-text-center uk-text-italic uk-text-muted uk-text-small">
						{@render (empty || emptyFallback)()}
					</td>
				</tr>
			{:else}
				{#each tableItems as item}
					{@render (row || rowFallback)(item) }
				{/each}
			{/if}

			</tbody>
		</table>
	</div>
	<div>
		{@render afterTable?.()}
	</div>
	<div class="uk-text-muted">
		Total items: {totalItems}
	</div>
	<div bind:this={infiniteScrollEl}></div>
</div>

<style>
	.spinning {
		animation: spin 0.7s linear infinite;
	}
	@keyframes spin {
		to { transform: rotate(360deg); }
	}
	.skeleton-cell {
		height: 16px;
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
