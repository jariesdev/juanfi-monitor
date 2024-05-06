<script lang="ts">
    import {onMount, onDestroy} from "svelte";
    import debounce from 'lodash/debounce'
    import {apiUrl} from "$lib/store";
    import type {iSale, iVendo} from "$lib/interfaces";
    import DateTime from "$lib/components/DateTime.svelte";


    let sales: iSale[] = []
    let searchInput: string = ''
    let isLoading: boolean = false
    let baseApiUrl: string = ''
    let controller: AbortController | undefined = undefined
    let vendoId: number
    let vendos: iVendo[] = []

    export const loadData: Function = debounce(async (): Promise<void> => {
        const queryParams = []
        isLoading = false

        let url = `${baseApiUrl}/sales`
        if (!!searchInput) {
            queryParams.push({
                name: 'q',
                value: searchInput
            })
        }

        if (!!vendoId) {
            queryParams.push({
                name: 'vendo_id',
                value: vendoId
            })
        }

        if (queryParams.length > 0) {
            url = url + '?' + queryParams.map((o) => `${o.name}=${o.value}`).join('&')
        }

        controller = new AbortController()
        const signal = controller.signal
        const request = new Request(url, {method: "GET", signal: signal});

        fetch(request)
            .then((response) => {
                if (response.status === 200) {
                    return response.json();
                } else {
                    throw new Error("Something went wrong on API server!");
                }
            })
            .then((response) => {
                sales = response.data
            })
            .catch((error) => {
                console.error(error);
                sales = []
            })
            .finally(() => {
                isLoading = false
            })
    }, 250, {maxWait: 1000})

    function loadOptions(): void {
        let url = `${baseApiUrl}/vendo-machines`
        const request = new Request(url, {method: 'GET'});
        fetch(request)
            .then((response) => {
                if (response.status === 200) {
                    return response.json();
                } else {
                    throw new Error("Something went wrong on API server!");
                }
            })
            .then((response) => {
                vendos = response.data
            })
            .catch(() => {
                vendos = []
            })
    }

    $: searchInput, loadData();

    apiUrl.subscribe(function (value) {
        baseApiUrl = value
    })

    onMount(() => {
        loadOptions()
        loadData()
    })
    onDestroy(() => {
        controller && controller.abort('component destroyed')
    })
</script>

<div class="uk-card uk-card-default uk-card-body">
    <div class="uk-margin-small-top uk-grid uk-grid-small" style="row-gap: 15px;">
        <div class="uk-width-2-3@s">
            <h3 class="uk-card-title">Sales</h3>
        </div>
        <div class="uk-width-1-3@s">
            <input bind:value={searchInput} class="uk-input uk-form-small" type="search"
                   placeholder="Search" aria-label="Input">
        </div>
    </div>
    <div class="uk-margin-small-top uk-grid uk-grid-small uk-child-width-1-2@s uk-child-width-1-3@m" style="row-gap: 15px">
        <div>
            <select bind:value={vendoId}
                    name="vendo_id"
                    id="vendo_id"
                    class="uk-select uk-form-small uk-child-width-1-1"
                    on:change={loadData}>
                <option value="{undefined}">All</option>
                {#each vendos as vendo}
                    <option value="{vendo.id}">{vendo.name}</option>
                {/each}
            </select>
        </div>
    </div>
    <div class="uk-overflow-auto">
        <table class="uk-table uk-table-divider">
            <thead>
            <tr>
                <th>Time</th>
                <th>Vendo</th>
                <th>Mac Address</th>
                <th>Voucher</th>
                <th>Amount</th>
                <th>Created At</th>
            </tr>
            </thead>
            <tbody>
            {#each sales as log}
                <tr>
                    <td>
                        <DateTime date="{log.sale_time}" class="uk-text-nowrap" />
                    </td>
                    <td>{log.vendo?.name}</td>
                    <td>{log.mac_address}</td>
                    <td>{log.voucher}</td>
                    <td>{log.amount}</td>
                    <td>
                        <DateTime date="{log.created_at}" class="uk-text-nowrap" />
                    </td>
                </tr>
            {/each}
            </tbody>
        </table>
    </div>
</div>
