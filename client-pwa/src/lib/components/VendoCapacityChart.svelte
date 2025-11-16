<script lang="ts">
	import Chart from 'chart.js/auto';
	import { onDestroy, onMount } from 'svelte';
	import { baseApiUrl } from '$lib/env';
	import type { ChartConfiguration } from 'chart.js';
	import 'chartjs-adapter-moment';
	import type { Plugin } from 'chart.js/dist/types';

	let chartData: any[] = [];
	let canvas: HTMLCanvasElement;
	let intervalId: any;
	let isLoading: boolean = false;
	let chart: Chart;
	let controller: AbortController | undefined = undefined;

	const horizontalLinePlugin: Plugin = {
		afterDraw: function(chartInstance) {
			const { ctx, canvas } = chartInstance;
			const yScale = chartInstance.scales.y;

			const { horizontalLine } = chartInstance.options;
			if (horizontalLine && yScale) {
				let textOffset = -2;
				for (let index = 0; index < horizontalLine.length; index++) {
					let line = horizontalLine[index];
					let style;
					if (!line.style) {
						style = 'rgba(169,169,169, .6)';
					} else {
						style = line.style;
					}
					let yValue;
					if (line.y) {
						yValue = yScale.getPixelForValue(line.y);
					} else {
						yValue = 0;
					}
					ctx.lineWidth = 1;
					if (yValue) {
						ctx.beginPath();
						ctx.setLineDash([5, 3]);
						ctx.moveTo(0, yValue);
						ctx.lineTo(canvas.width, yValue);
						ctx.strokeStyle = style;
						ctx.stroke();
					}
					if (line.text) {
						ctx.fillStyle = style;
						ctx.fillText(line.text, 0, yValue + ctx.lineWidth + textOffset);
					}
				}
				return;
			}
		}
	};

	function renderChart(): void {

		chartData = {
			labels: [],
			datasets: []
		};
		const chartConfig: ChartConfiguration = {
			type: 'bar',
			data: chartData,
			options: {
				responsive: true,
				horizontalLine: [
					{
						y: 5000,
						style: 'rgba(255, 0, 0, .9)',
						text: 'FULL'
					},
					{
						y: 4000,
						style: 'orange',
						text: ''
					}
				],
				scales: {
					y: {
						min: 0,
						max: 5500
					}
				},
				interaction: {
					intersect: false,
					mode: 'index'
				},
				plugins: {
					legend: false
					// 	tooltip: {
					// 		enabled: true,
					// 		position: 'nearest',
					// 		callbacks: {
					// 			title: function(tooltipItems: TooltipItem[]) {
					// 				const { raw } = tooltipItems[0];
					// 				return moment(raw.time).format('MMMM DD, Y h:mm A');
					// 			},
					// 		}
					// 	}
				}
			},
			plugins: [horizontalLinePlugin]
		};
		chart = new Chart(canvas, chartConfig);
	}

	function loadChartData(): void {
		controller = new AbortController();
		const signal = controller.signal;

		const request = new Request(
			`${baseApiUrl}/vendo-machines?` + (new URLSearchParams({is_active: true}).toString()),
			{
				method: 'GET',
				signal: signal
			}
		);
		fetch(request)
			.then((response) => {
				if (response.status === 200) {
					return response.json();
				} else {
					throw new Error('Something went wrong on API server!');
				}
			})
			.then(({ data }) => {
				if (chart) {
					const datasets = [{
						label: 'Current Sales',
						data: data.map((d) => d?.recent_status.current_sales | 0)
					}];
					chart.data.labels = data.map((d) => d.name);
					chart.data.datasets = datasets;
					chart.update();
				}
			})
			.catch((error) => {
				console.error(error);
			})
			.finally(() => {
				isLoading = false;
			});
	}

	onMount(() => {
		renderChart();
		loadChartData();
		intervalId = setInterval(() => loadChartData(), 30 * 1000);
	});

	onDestroy(() => {
		controller && controller.abort('component destroyed');
		if (intervalId) {
			clearInterval(intervalId);
		}
	});
</script>

<canvas bind:this={canvas} />
