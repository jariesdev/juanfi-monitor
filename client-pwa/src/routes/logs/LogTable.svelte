<script lang="ts">
    import {onMount} from "svelte";
    import debounce from 'lodash/debounce'

    interface iVendoLog {
        id: number
        log_time: string
        description: string
        created_at: string
    }

    let logs: iVendoLog[] = []
    let searchInput: string = ''
    let isLoading: boolean = false

    export const loadData: Function = debounce(async (): Promise<void> => {
        isLoading = false

        let url = "http://192.46.225.21:8000/logs"
        if (!!searchInput) {
            url = `${url}?q=${searchInput}`
        }
        const request = new Request(url, {method: "GET"});

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

    onMount(() => {
        loadData()
    })

</script>

<div class="uk-card uk-card-default uk-card-body">
    <div class="uk-flex uk-flex-between">
        <h3 class="uk-card-title">System Logs</h3>
        <div>
            <input bind:value={searchInput} class="uk-input uk-form-width-medium uk-form-small" type="search"
                   placeholder="Search" aria-label="Input">
        </div>
    </div>
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
