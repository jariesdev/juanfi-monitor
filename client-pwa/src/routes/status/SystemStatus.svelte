<script lang="ts">
    import {onMount, onDestroy} from "svelte";
    import moment from 'moment'
    import {apiUrl} from "$lib/store";

    interface iStatus {
        label: string
        text: string
    }

    let statuses: iStatus[] = []
    let isLoading: boolean = false
    let systemUptime: number = 0
    let serverTime: number = 0
    let baseApiUrl: string = ''
    let controller: AbortController | undefined = undefined
    let intervalId: number = 0

    function loadStatuses(): void {
        controller = new AbortController()
        const signal = controller.signal
        const request = new Request(`${baseApiUrl}/vendo_status?nosw=1`, {method: "GET", signal: signal})
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
                serverTime = response.server_time
                const list: iStatus[] = []
                Object.keys(response).forEach((key): void => {
                    if (!['system_uptime_ms', 'server_time'].includes(key)) {
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
        if(intervalId) return;
        intervalId = setInterval(() => {
            systemUptime += 1000
        }, 1000)
    }

    function toRelativeTime(time: number): string {
        const d = moment().diff(Date.now() - time, 'days')
        const h = moment().diff(Date.now() - time, 'hours') % 24
        const m = moment().diff(Date.now() - time, 'minutes') % 60
        const s = moment().diff(Date.now() - time, 'seconds') % 60
        const mf = String(m).padStart(2, '0')
        const sf = String(s).padStart(2, '0')

        if (d > 0) {
            return `${d}d ${h}:${mf}:${sf}`
        } else if (h > 0) {
            return `${h}:${mf}:${sf}`
        } else if (m > 0) {
            return `${mf}:${sf}`
        }
        return `${sf}`
    }

    $: serverTimeString = (): string => {
        return moment(serverTime).format()
    }

    apiUrl.subscribe(function (value) {
        baseApiUrl = value
    })

    onMount(() => {
        loadStatuses()
        // refresh status
        setInterval(() => loadStatuses(), 30 * 1000)
    })
    onDestroy(() => {
        controller && controller.abort('component destroyed')
        if (intervalId) {
            clearInterval(intervalId)
        }
    })
</script>

<div class="uk-card uk-card-default uk-card-body">
    {#if statuses.length > 0}
        <ul class="uk-list uk-list-divider">
            <li class="uk-flex uk-flex-between">
                <span>System Uptime</span>
                <span>{toRelativeTime(systemUptime)}</span>
            </li>
            <li class="uk-flex uk-flex-between">
                <span>Server Time</span>
                <span>{serverTimeString()}</span>
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