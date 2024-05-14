<script lang="ts">
    import { browser } from '$app/environment';
    import debounce from "lodash/debounce";
    import {onDestroy, onMount} from "svelte";

    interface Filter {
        [key: string]: string | number | boolean | null | undefined
    }

    interface QueryParameters extends Filter {
        q?: string
        page: number
        perPage: number
    }

    interface TableHeader {
        label: string
        field: string
    }

    interface RowItem {
        [key: string]: any
    }

    // props
    export let url: string
    export let headers: TableHeader[] = []
    export let filters: Filter = {}
    export let title: string = "Table Records";
    export let perPage: number = 15;
    export let infiniteScroll: boolean = false;

    // variables
    let currentPage: number = 1
    let maxPage: number = 1
    let isLoading: boolean = false
    let tableItems: RowItem[] = []
    let searchInput: string = ''
    let controller: AbortController | undefined = undefined
    let infiniteScrollEl: HTMLDivElement

    // computed property queryParams
    $: ({queryParams} = ((): { queryParams: QueryParameters } => {
        let params: Filter = {}
        // remove empty params
        Object.keys(filters).forEach(k => {
            if (filters[k] != undefined && filters[k] != null) {
                params[k] = filters[k]
            }
        })

        return {
            queryParams: {
                ...params,
                q: searchInput,
                page: currentPage,
                perPage: perPage
            }
        }
    })())

    // load table data
    export const loadData: Function = debounce(async (): Promise<void> => {
        let localUrl = url
        isLoading = false

        if (Object.keys(queryParams).length > 0) {
            localUrl = localUrl + '?' + Object.keys(queryParams).map((k: string) => `${k}=${queryParams[k]}`).join('&')
        }
        controller = new AbortController()
        const signal = controller.signal
        const request = new Request(localUrl, {method: "GET", signal: signal});

        fetch(request)
            .then((response) => {
                if (response.status === 200) {
                    return response.json();
                } else {
                    throw new Error("Something went wrong on API server!");
                }
            })
            .then((response) => {
                maxPage = 1
                tableItems = response.data || []
            })
            .catch((error) => {
                console.error(error);
                tableItems = []
            })
            .finally(() => {
                isLoading = false
            })
    }, 250, {maxWait: 1000})

    function getCellValue(item: RowItem, header: TableHeader) {
        return item[header.field] || ''
    }

    // watch params then reload data
    $: queryParams, loadData();

    // reset when search
    $: searchInput, ()=>{
        currentPage = 1
    };
    $: filters, ()=>{
        currentPage = 1
    };

    // watcher infinite scroll
    $: {
        if (browser && infiniteScrollEl) {
            let options = {
                rootMargin: "0px",
                threshold: 1.0,
            };
            let callback = (entries: any[]) => {
                entries.forEach(() => {
                    if(currentPage < maxPage) {
                        currentPage += 1
                        loadData()
                    }
                });
            };

            let observer = new IntersectionObserver(callback, options);
            observer.observe(infiniteScrollEl);
        }
    }

    onMount(() => {
        // load data from url
        loadData()
    })

    onDestroy(() => {
        // on component destroy, cancel ongoing HTTP request
        controller && controller.abort('component destroyed')
    })
</script>


<div class="uk-card uk-card-default uk-card-body">
    <div class="uk-margin-small-top uk-grid uk-grid-small" style="row-gap: 15px;">
        <div class="uk-width-2-3@s">
            <h3 class="uk-card-title">{title}</h3>
        </div>
        <div class="uk-width-1-3@s">
            <input bind:value={searchInput}
                   class="uk-input uk-form-small"
                   type="search"
                   placeholder="Search"
                   aria-label="Input">
        </div>
    </div>
    <div>
        <slot name="before-table"/>
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
                        <slot name="empty">
                            No record yet.
                        </slot>
                    </td>
                </tr>
            {/if}
            {#each tableItems as item}
                <slot {item}>
                    <tr>
                        {#each headers as header}
                            <td>
                                <!-- svelte 4 does not support dynamic slot names -->
                                <slot name="cell" {item} {header}>
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
        <slot name="after-table"/>
    </div>
    <div bind:this={infiniteScrollEl}></div>
</div>

