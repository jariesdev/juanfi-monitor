<script lang="ts">
    import Chart from 'chart.js/auto';
    import {onMount, onDestroy} from "svelte";
    import {apiUrl} from "$lib/store";
    import map from "lodash/map";
    import groupBy from "lodash/groupBy";

    let chartData: any[] = []
    let canvas: HTMLCanvasElement
    let baseApiUrl: string = ''
    let intervalId: number
    let isLoading: boolean = false
    let chart: Chart
    let controller: AbortController | undefined = undefined

    interface iDailySale {
        date: string
        total: number
        vendo_id: number
        vendo_name: string
    }

    function renderChart(): void {
        chartData = {
            datasets: []
        };
        chart = new Chart(canvas, {
            type: 'line',
            data: chartData,
            options: {
                parsing: {
                    xAxisKey: 'date',
                    yAxisKey: 'total'
                },
                scales: {
                    y: {
                        ticks: {
                            callback: function (value: string) {
                                return 'PHP ' + value;
                            }
                        }
                    }
                }
            }
        });
    }

    function loadChartData(): void {
        controller = new AbortController()
        const signal = controller.signal
        const request = new Request(`${baseApiUrl}/daily-sales`, {method: "GET", signal: signal})
        fetch(request)
            .then((response) => {
                if (response.status === 200) {
                    return response.json();
                } else {
                    throw new Error("Something went wrong on API server!");
                }
            })
            .then(({data}) => {
                if (chart) {
                    const byVendo = groupBy(data, "vendo_id")
                    const datasets = map(byVendo, (vendoSales: iDailySale[]) => {
                        const data = map(vendoSales, (d: iDailySale) => {
                            const dt = new Date(Date.parse(d.date))
                            const m = dt.getMonth() + 1
                            const dy = dt.getDate()
                            const df = String(dy).padStart(2, '0')
                            const mf = String(m).padStart(2, '0')
                            return {
                                date: `${mf}-${df}`,
                                total: d.total
                            }
                        });
                        const vendoName = vendoSales[0].vendo_name;
                        return {
                            label: vendoName,
                            data: data,
                            borderWidth: 1,
                            tension: 0.4,
                        }
                    })

                    chart.data.datasets = datasets
                    chart.update();
                }
            })
            .catch((error) => {
                console.error(error);
            })
            .finally(() => {
                isLoading = false
            })
    }

    apiUrl.subscribe(function (value) {
        baseApiUrl = value
    })

    onMount(() => {
        renderChart()
        loadChartData()
        setInterval(() => loadChartData(), 30 * 1000)
    })

    onDestroy(() => {
        controller && controller.abort('component destroyed')
        if (intervalId) {
            clearInterval(intervalId)
        }
    })
</script>


<canvas bind:this={canvas}></canvas>