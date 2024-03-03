<script lang="ts">
    import {onMount, onDestroy} from "svelte";
    import debounce from 'lodash/debounce'
    import {apiUrl} from "$lib/store";
    import type {iSale} from "$lib/interfaces";


    let sales: iSale[] = []
    let searchInput: string = ''
    let isLoading: boolean = false
    let baseApiUrl: string = ''
    let controller: AbortController | undefined = undefined

    export const loadData: Function = debounce(async (): Promise<void> => {
        isLoading = false

        let url = `${baseApiUrl}/sales`
        if (!!searchInput) {
            url = `${url}?q=${searchInput}`
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

    function humanizeTime(time: string): string {
        return new Date(Date.parse(time)).toLocaleString()
    }

    $: searchInput, loadData();

    apiUrl.subscribe(function (value) {
        baseApiUrl = value
    })

    onMount(() => {
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
                    <td title="{humanizeTime(log.sale_time)}">{log.sale_time}</td>
                    <td>{log.vendo?.name}</td>
                    <td>{log.mac_address}</td>
                    <td>{log.voucher}</td>
                    <td>{log.amount}</td>
                    <td title="{humanizeTime(log.created_at)}">{log.created_at}</td>
                </tr>
            {/each}
            </tbody>
        </table>
    </div>
</div>
