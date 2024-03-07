<script lang="ts">
    import {onMount, onDestroy} from "svelte";
    import moment from 'moment'
    import {apiUrl} from "$lib/store";

    interface iStatus {
        key: string
        label: string
        text: string
    }

    export let vendoId: number
    let statuses: iStatus[] = []
    let isLoading: boolean = false
    let systemUptime: number = 0
    let serverTime: number = 0
    let baseApiUrl: string = ''
    let controller: AbortController | undefined = undefined
    let intervalId: any
    let timeIntervalId: any
    let isWithdrawing: boolean = false

    function loadStatuses(): void {
        controller = new AbortController()
        const signal = controller.signal
        const request = new Request(`${baseApiUrl}/vendo-machines/${vendoId}/status?nosw=1`, {
            method: "GET",
            signal: signal
        })
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
                            key: key,
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
        if (timeIntervalId) return;
        timeIntervalId = setInterval(() => {
            systemUptime += 1000
        }, 1000)
    }

    function withdrawCurrenSale(): void {

        const confirmed = confirm('This will reset the current sales counter to 0. Proceed?')

        if (!confirmed) {
            return
        }

        isWithdrawing = true
        let url = `${baseApiUrl}/vendo-machines/${vendoId}/withdraw-current-sales`;
        controller = new AbortController()
        const signal = controller.signal
        const request: Request = new Request(url, {method: 'POST', signal})
        fetch(request)
            .then((response) => {
                if (response.ok) {
                    return response.json()
                }
                throw new Error(response.statusText)
            })
            .finally(() => {
                isWithdrawing = false
            })
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
       intervalId = setInterval(() => loadStatuses(), 30 * 1000)
    })
    onDestroy(() => {
        controller && controller.abort('component destroyed')
        if (intervalId) {
            clearInterval(intervalId)
        }
        if (timeIntervalId) {
            clearInterval(timeIntervalId)
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
                    <div>{status.label}</div>
                    <div class="uk-flex uk-flex-center uk-flex-right uk-flex-wrap">
                        {#if status.key === 'current_coin_count'}
                            <button disabled="{isWithdrawing || parseFloat(status.text || '0') === 0}"
                                    class="uk-button uk-button-primary uk-button-small uk-border-rounded"
                                    type="button"
                                    on:click={withdrawCurrenSale}>
                                <i uk-icon="icon: credit-card" class="uk-margin-small-right"></i>
                                Withdraw
                            </button>
                            <span class="uk-margin-small-left">{status.text}</span>
                        {:else}
                            <span>{status.text}</span>
                        {/if}

                    </div>
                </li>
            {/each}
        </ul>
    {:else }
        <p class="uk-text-info">
            Status not available yet.
        </p>
    {/if}
</div>