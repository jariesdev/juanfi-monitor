<script lang="ts">
    import {onMount} from "svelte";
    import moment from 'moment'

    interface iStatus {
        label: string
        text: string
    }

    let statuses: iStatus[] = []
    let isLoading: boolean = false
    let systemUptime: number = 0

    function loadStatuses(): void {
        const request = new Request("http://192.46.225.21:8000/vendo_status", {method: "GET"})
        fetch(request)
            .then((response) => {
                if (response.status === 200) {
                    return response.json();
                } else {
                    throw new Error("Something went wrong on API server!");
                }
            })
            .then((response) => {
                systemUptime = response.system_uptime_ms
                const list: iStatus[] = []
                Object.keys(response).forEach((key): void => {
                    if (!['system_uptime_ms'].includes(key)) {
                        list.push({
                            label: titleCase(key),
                            text: response[key]
                        })
                    }

                })
                statuses = list
                startTimer()
            })
            .catch((error) => {
                console.error(error);
            })
            .finally(() => {
                isLoading = false
            })
    }

    function titleCase(s: string) {
        return s.replace(/^_*(.)|_+(.)/g, (s: string, c: string, d: string) => c ? c.toUpperCase() : ' ' + d.toUpperCase())
    }

    function startTimer(): void {
        setInterval(() => {
            systemUptime += 1000
        }, 1000)
    }

    function toRelativeTime(time: number): string {
        return moment(Date.now() - time).fromNow()
    }

    onMount(() => {
        loadStatuses()
    })
</script>

<div class="uk-card uk-card-default uk-card-body">
    {#if statuses.length > 0}
        <ul class="uk-list uk-list-divider">
            <li class="uk-flex uk-flex-between">
                <span>System Uptime</span>
                <span>{toRelativeTime(systemUptime)}</span>
            </li>
            {#each statuses as status}
                <li class="uk-flex uk-flex-between">
                    <span>{status.label}</span>
                    <span>{status.text}</span>
                </li>
            {/each}
        </ul>
    {:else }
        <p class="uk-text-info">
            Status not available yet.
        </p>
    {/if}
</div>