<script lang="ts">
    import {onMount, onDestroy} from "svelte";
    import debounce from 'lodash/debounce'
    import {apiUrl} from "$lib/store";
    import type {iSale} from "$lib/interfaces";
    import DateTime from "$lib/components/DateTime.svelte";


    let withdrawals: iSale[] = []
    let isLoading: boolean = false
    let baseApiUrl: string = ''
    let controller: AbortController | undefined = undefined
    let isHumanizeDate: boolean = false

    export const loadData: Function = debounce(async (): Promise<void> => {
        isLoading = false

        let url = `${baseApiUrl}/withdrawals`
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
                withdrawals = response.data
            })
            .catch((error) => {
                console.error(error);
                withdrawals = []
            })
            .finally(() => {
                isLoading = false
            })
    }, 250, {maxWait: 1000})

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
            <h3 class="uk-card-title">Withdrawals</h3>
        </div>
    </div>
    <div class="uk-overflow-auto">
        <table class="uk-table uk-table-divider">
            <thead>
            <tr>
                <th>Vendo</th>
                <th>Amount</th>
                <th>Date</th>
            </tr>
            </thead>
            <tbody>
            {#each withdrawals as item}
                <tr>
                    <td>{item.vendo?.name}</td>
                    <td>{item.amount}</td>
                    <td>
                        <DateTime bind:humanized={isHumanizeDate} date="{item.created_at}" on:click={() => { isHumanizeDate = !isHumanizeDate }} class="uk-text-nowrap" />
                    </td>
                </tr>
            {/each}
            </tbody>
        </table>
    </div>
</div>
