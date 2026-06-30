<script lang="ts">
	import get from 'lodash/get';
	import type { Snippet } from 'svelte';
	import type { RowItem, SimpleTableHeader } from '$lib/types/datatable';

	interface Props {
		headers: SimpleTableHeader[];
		items: RowItem[];
		emptyText?: string;
		cell?: Snippet<[RowItem, SimpleTableHeader, Function]>;
	}

	const { headers = [], items = [], emptyText = 'No records yet.', cell }: Props = $props();

	let sortField: string | undefined = $state(undefined);
	let sortDir: 'asc' | 'desc' = $state('asc');

	function getCellValue(item: RowItem, header: SimpleTableHeader) {
		return get(item, header.field, '');
	}

	function toggleSort(header: SimpleTableHeader) {
		if (header.sortable === false) return;

		if (sortField === header.field) {
			sortDir = sortDir === 'asc' ? 'desc' : 'asc';
		} else {
			sortField = header.field;
			sortDir = 'asc';
		}
	}

	// click-to-sort: numbers sort numerically, everything else sorts as a string
	const sortedItems = $derived.by(() => {
		if (!sortField) return items;
		const field = sortField;
		const dir = sortDir === 'asc' ? 1 : -1;

		return [...items].sort((a, b) => {
			const av = get(a, field, '');
			const bv = get(b, field, '');
			if (typeof av === 'number' && typeof bv === 'number') return (av - bv) * dir;
			return String(av).localeCompare(String(bv)) * dir;
		});
	});
</script>

{#snippet cellFallback(item: RowItem, header: SimpleTableHeader, getCellValue: Function)}
	{getCellValue(item, header)}
{/snippet}

<table class="uk-table uk-table-divider uk-table-small simple-table">
	<thead>
		<tr>
			{#each headers as header}
				<th
					class:uk-text-center={header.align === 'center'}
					class:uk-text-right={header.align === 'right'}
					class:sortable={header.sortable !== false}
					onclick={() => toggleSort(header)}
				>
					{header.label}
					{#if header.sortable !== false}
						<span
							class="sort-icon"
							class:active={sortField === header.field}
							uk-icon={`icon: chevron-${sortField === header.field && sortDir === 'desc' ? 'down' : 'up'}`}
						></span>
					{/if}
				</th>
			{/each}
		</tr>
	</thead>
	<tbody>
		{#if sortedItems.length === 0}
			<tr>
				<td colspan={headers.length} class="uk-text-center uk-text-italic uk-text-muted uk-text-small">
					{emptyText}
				</td>
			</tr>
		{:else}
			{#each sortedItems as item}
				<tr>
					{#each headers as header}
						<td
							class:uk-text-center={header.align === 'center'}
							class:uk-text-right={header.align === 'right'}
						>
							{@render (cell || cellFallback)(item, header, getCellValue)}
						</td>
					{/each}
				</tr>
			{/each}
		{/if}
	</tbody>
</table>

<style>
	.simple-table th {
		user-select: none;
	}

	.simple-table th.sortable {
		cursor: pointer;
	}

	.simple-table th.sortable:hover {
		color: #222;
	}

	.sort-icon {
		opacity: 0.25;
		margin-left: 2px;
		vertical-align: middle;
	}

	.sort-icon.active {
		opacity: 1;
	}

	:global(.action-btns .uk-icon-button) {
		background-color: transparent;
	}
</style>
