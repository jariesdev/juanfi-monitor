<script lang="ts">
    import Chart from 'chart.js/auto';
    import {onMount, onDestroy} from "svelte";
    import {apiUrl} from "$lib/store";

    let chartData: any[] = []
    let canvas: HTMLCanvasElement
    let baseApiUrl: string = ''
    let intervalId: number
    let isLoading: boolean = false

    interface iDailySale {
        date: string
        total: number
    }

    function renderChart(data: any[]): void {
        const chart = new Chart(canvas, {
            type: 'line',
            data: data,
            options: {
                onClick: (e) => {
                    const canvasPosition = getRelativePosition(e, chart);

                    // Substitute the appropriate scale IDs
                    const dataX = chart.scales.x.getValueForPixel(canvasPosition.x);
                    const dataY = chart.scales.y.getValueForPixel(canvasPosition.y);
                }
            }
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
            const labels = data.map((d: iDailySale) => d.date);
            const saleData = data.map((d: iDailySale) => d.total);
            chartData = {
                labels: labels,
                datasets: [{
                    label: 'Daily Sales',
                    data: saleData,
                    fill: false,
                    borderColor: 'rgb(75, 192, 192)',
                    tension: 0.1
                }]
            };

            renderChart(chartData)
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
        loadChartData()
        setInterval(() => loadChartData(), 5000)
    })

    onDestroy(() => {
        if(intervalId) {
            clearInterval(intervalId)
        }
    })
</script>


<canvas bind:this={canvas}></canvas>