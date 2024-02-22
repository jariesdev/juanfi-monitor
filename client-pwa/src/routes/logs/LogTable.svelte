<script lang="ts">
    import {onMount, onDestroy} from "svelte";
    import debounce from 'lodash/debounce'
    import {apiUrl} from "$lib/store";
    import moment from "moment/moment";

    interface iVendoLog {
        id: number
        log_time: string
        description: string
        created_at: string
    }

    let logs: iVendoLog[] = []
    let searchInput: string = ''
    let date: string = moment().format('Y-MM-DD')
    let isLoading: boolean = false
    let baseApiUrl: string = ''
    let controller: AbortController | undefined = undefined
    let showFilter: boolean = true

    export const loadData: Function = debounce(async (): Promise<void> => {
        isLoading = false

        const queryParams = []

        let url = `${baseApiUrl}/logs`
        if (!!searchInput) {
            queryParams.push({
                name: 'q',
                value: searchInput
            })
        }

        if (date) {
            queryParams.push({
                name: 'date',
                value: date
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
                logs = response.data
            })
            .catch((error) => {
                console.error(error);
                logs = []
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
    <div class="uk-column-1-2@s">
        <h3 class="uk-card-title">System Logs</h3>
        <div class="uk-text-right">
            <div class="uk-inline">
                <button class="uk-form-icon uk-form-icon-flip" uk-icon="icon: {showFilter ? 'chevron-up' : 'chevron-down'}"
                   on:click|preventDefault={() => showFilter=!showFilter} title="Show filters"></button>
                <input bind:value={searchInput} class="uk-input uk-form-small" type="search"
                       placeholder="Search" aria-label="Input">
            </div>
        </div>
    </div>
    {#if showFilter}
        <div class="uk-margin-small-top">
            <div class="uk-column-1-2@s uk-column-1-3@m">
                <input bind:value={date} type="date" name="date" id="date"
                       class="uk-input uk-form-small" on:change={loadData}>
            </div>
        </div>
    {/if}
    <div class="uk-overflow-auto">
        <table class="uk-table uk-table-divider">
            <thead>
            <tr>
                <th>Time</th>
                <th>Description</th>
                <th>Created At</th>
            </tr>
            </thead>
            <tbody>
            {#each logs as log}
                <tr>
                    <td title="{humanizeTime(log.log_time)}">{log.log_time}</td>
                    <td>{log.description}</td>
                    <td title="{humanizeTime(log.created_at)}">{log.created_at}</td>
                </tr>
            {/each}
            </tbody>
        </table>
    </div>
</div>
