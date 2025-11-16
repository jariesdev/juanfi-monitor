<script lang="ts">
	import {browser} from '$app/environment';
	import {onDestroy, onMount} from 'svelte';
	import debounce from 'lodash/debounce';
	import get from 'lodash/get';
	import type {Filter, QueryParameters, RowItem, TableHeader} from '$lib/types/datatable';

	// props
	interface Props {
		url: string,
		headers: TableHeader[],
		filters: Filter,
		title: string,
		perPage?: number
	}

	const {url, headers = [], filters = {}, title = 'Table Records', perPage = 15}: Props = $props()

	// states
	let currentPage: number = $state(1);
	let maxPage: number = $state(1);
	let totalItems: number = $state(1);
	let isLoading: boolean = $state(false);
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
				});
		},
		250,
		{maxWait: 1000}
	);

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
	$effect(() => {
		if (browser && infiniteScrollEl) {
			let options = {
				rootMargin: '0px',
				threshold: 0
			};
			let callback = (entries: any[]) => {
				entries.forEach((e) => {
					if (currentPage < maxPage && !isLoading && e.isIntersecting) {
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
	})

	onMount(() => {
		// load data from url
		loadData();
	});

	onDestroy(() => {
		// on component destroy, cancel ongoing HTTP request
		controller && controller.abort('component destroyed');
	});
</script>

<div class="uk-card uk-card-default uk-card-body">
	<div class="uk-margin-small-top uk-grid uk-grid-small" style="row-gap: 15px;">
		<div class="uk-width-2-3@s">
			<h3 class="uk-card-title">{title}</h3>
		</div>
		<div class="uk-width-1-3@s">
			<input
				bind:value={searchInput}
				class="uk-input uk-form-small"
				type="search"
				placeholder="Search"
				aria-label="Input"
				onchange={()=>resetTableQuery()}
			/>
		</div>
	</div>
	<div>
		<slot name="before-table" />
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
			{#if tableItems.length === 0}
				<tr>
					<td colspan="99" class="uk-text-center uk-text-italic uk-text-muted uk-text-small">
						<slot name="empty">No record yet.</slot>
					</td>
				</tr>
			{/if}
			{#each tableItems as item}
				<slot name="item" {item}>
					<tr>
						{#each headers as header}
							<td>
								<!-- svelte 4 does not support dynamic slot names -->
								<slot name="cell" {item} {header} {getCellValue}>
									{getCellValue(item, header)}
								</slot>
							</td>
						{/each}
					</tr>
				</slot>
			{/each}
			</tbody>
		</table>
	</div>
	<div>
		<slot name="after-table" />
	</div>
	<div class="uk-text-muted">
		Total items: {totalItems}
	</div>
	<div bind:this={infiniteScrollEl} />
</div>
