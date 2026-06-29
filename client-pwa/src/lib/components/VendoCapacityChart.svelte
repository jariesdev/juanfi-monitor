<script lang="ts">
	import Chart from 'chart.js/auto';
	import { onDestroy, onMount } from 'svelte';

	import type { ChartConfiguration, Plugin } from 'chart.js';
	import 'chartjs-adapter-moment';

	let chartData: any = {};
	let canvas: HTMLCanvasElement;
	let intervalId: any;
	let isLoading: boolean = true;
	let chart: Chart;
	let controller: AbortController | undefined = undefined;
	let lastDataHash: string = '';

	const horizontalLinePlugin: Plugin = {
		id: 'horizontalLine',
		afterDraw: function(chartInstance: Chart) {
			const { ctx, canvas } = chartInstance;
			const yScale = chartInstance.scales.y;

			const { horizontalLine } = (chartInstance.options as any);
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
		const chartConfig = {
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
						max: 5500,
							ticks: {
								stepSize: 1,
								callback: function(value: string | number) {
									const n = Number(value)
									return '₱ ' + n.toLocaleString();
								}
							}
					}
				},
				interaction: {
					intersect: false,
					mode: 'index'
				},
				plugins: {
					legend: { display: false }
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
		chart = new Chart(canvas, chartConfig as any);
	}

	function loadChartData(): void {
		controller = new AbortController();
		const signal = controller.signal;

		const request = new Request(
			`/x-api/vendo-machines?` + (new URLSearchParams({is_active: 'true'}).toString()),
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
					const dataHash = JSON.stringify(data);
					if (dataHash === lastDataHash) return;
					lastDataHash = dataHash;

					const newLabels = data.map((d: any) => d.name);
					const newValues = data.map((d: any) => d?.recent_status.current_sales | 0);

					chart.data.labels = newLabels;

					if (chart.data.datasets.length > 0) {
						chart.data.datasets[0].data = newValues;
					} else {
						chart.data.datasets = [{ label: 'Current Sales', data: newValues }];
					}

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

<div class="chart-wrapper">
	<canvas bind:this={canvas}></canvas>
	{#if isLoading}
		<div class="chart-skeleton"></div>
	{/if}
</div>

<style>
	.chart-wrapper {
		position: relative;
		min-height: 200px;
	}
	.chart-skeleton {
		position: absolute;
		inset: 0;
		background: linear-gradient(90deg, #f0f0f0 25%, #e8e8e8 50%, #f0f0f0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.4s infinite;
		border-radius: 6px;
	}
	@keyframes shimmer {
		0% { background-position: 200% 0; }
		100% { background-position: -200% 0; }
	}
</style>
