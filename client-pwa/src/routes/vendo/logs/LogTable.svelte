<script lang="ts">
    import {onMount} from "svelte";

    interface iVendoLog {
        id: number
        log_time: string
        description: string
        created_at: string
    }

    let logs: iVendoLog[] = []
    let isLoading: boolean = false
    const request = new Request("http://192.46.225.21:8000/logs", {method: "GET"});

    export async function loadData(): Promise<void> {
        isLoading = false
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
    }

    function humanizeTime(time: string): string {
        return new Date(Date.parse(time)).toLocaleString()
    }

    onMount(() => {
        loadData()
    })

</script>

<div class="uk-card uk-card-default uk-card-body">
    <h3 class="uk-card-title">System Logs</h3>
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
