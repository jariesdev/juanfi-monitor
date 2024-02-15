<script lang="ts">
    import Chart from 'chart.js/auto';
    import {onMount, onDestroy} from "svelte";
    import {apiUrl} from "$lib/store";
    import map from "lodash/map";

    let chartData: any[] = []
    let canvas: HTMLCanvasElement
    let baseApiUrl: string = ''
    let intervalId: number
    let isLoading: boolean = false
    let chart: Chart

    interface iDailySale {
        date: string
        total: number
    }

    function renderChart(saleList: any[]): void {
        chartData = {
            labels: [],
            datasets: [{
                label: 'Daily Sales',
                data: [],
                fill: false,
                borderColor: 'rgb(75, 192, 192)',
                tension: 0.1
            }]
        };
        chart = new Chart(canvas, {
            type: 'line',
            data: chartData
        });
    }

    function loadChartData(): void {
        const request = new Request(`${baseApiUrl}/daily-sales`, {method: "GET"})
        fetch(request)
            .then((response) => {
                if (response.status === 200) {
                    return response.json();
                } else {
                    throw new Error("Something went wrong on API server!");
                }
            })
            .then(({data}) => {
                const labels = map(data, (d: iDailySale) => {
                    const dt = new Date(Date.parse(d.date))
                    const m = dt.getMonth() + 1
                    const dy = dt.getDate()
                    return `${m}-${dy}`
                });
                const saleData = map(data, (d: iDailySale) => d.total);
                if (chart) {
                    chart.data.labels = labels
                    chart.data.datasets = [{
                        label: 'Sales',
                        data: saleData,
                        borderWidth: 1
                    }]
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
        setInterval(() => loadChartData(), 10000)
    })

    onDestroy(() => {
        if (intervalId) {
            clearInterval(intervalId)
        }
    })
</script>


<canvas bind:this={canvas}></canvas>