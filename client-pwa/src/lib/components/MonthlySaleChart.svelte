<script lang="ts">
	import { Chart, type ChartConfiguration } from 'chart.js';
	import { onMount, onDestroy } from 'svelte';
	import map from 'lodash/map';
	import groupBy from 'lodash/groupBy';
	import keyBy from 'lodash/keyBy';

	import moment from 'moment';
	import 'chartjs-adapter-moment';

	let chartData: any = {};
	let canvas: HTMLCanvasElement;
	let intervalId: any;
	let isLoading: boolean = true;
	let chart: Chart;
	let controller: AbortController | undefined = undefined;
	let lastDataHash: string = '';

	interface iMonthlySale {
		month: string;
		total: number;
		vendo_id: number;
		vendo_name: string;
	}


	function renderChart(): void {
		chartData = {
			labels: [],
			datasets: []
		};
		let chartConfig: ChartConfiguration = {
			type: 'line',
			data: chartData,
			options: {
				parsing: {
					xAxisKey: 'date',
					yAxisKey: 'total'
				},
				interaction: {
					intersect: false,
					mode: 'index'
				},
				plugins: {
					tooltip: {
						enabled: true,
						position: 'nearest',
						callbacks: {
							title: function(tooltipItems: any[]) {
								const { raw } = tooltipItems[0];
								return moment(raw.date).format('MMMM Y');
							},
							footer: function(tooltipItems: any[]) {
								const total = tooltipItems.map(i => i.raw.total)
									.reduce((carry: number, value: number) => carry + value, 0);
								const formatTotal = new Intl.NumberFormat().format(total);
								return `Total ${formatTotal}`;
							}
						}
					}
				},
				elements: {
					point: {
						radius: 2.5
					}
				},
				scales: {
					x: {
						type: 'time',
						time: {
							unit: 'month'
						}
					},
					y: {
						min: 0,
						ticks: {
							stepSize: 1,
							callback: function(value: string | number) {
								const n = Number(value)
								return '₱ ' + n.toLocaleString();
							}
						}
					}
				}
			}
		};

		chart = new Chart(canvas, chartConfig);
	}

	function loadChartData(): void {
		controller = new AbortController();
		const signal = controller.signal;

		// request
		const fromDate = new Date(new Date().getFullYear() - 1, new Date().getMonth(), 1)
		const fromDateStr = fromDate.toISOString().split('T')[0]
		const toDate = new Date(new Date().getFullYear(), new Date().getMonth(), 0)
		const toDateStr = toDate.toISOString().split('T')[0]
		let url = `/x-api/monthly-sales?from_date=${fromDateStr}&to_date=${toDateStr}`
		const request = new Request(url, { method: 'GET', signal: signal });

		// send request to API
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

					const byVendo = groupBy(data, 'vendo_id');
					const newDatasets = map(byVendo, (sales: iMonthlySale[]) => {
						const points = [];
						const vendoSales2 = keyBy(sales, (o: iMonthlySale) => o.month);
						const sDate = new Date(fromDate);
						const eDate = new Date(toDate);
						while (sDate <= eDate) {
							const dKey = `${sDate.getFullYear()}-${String(sDate.getMonth() + 1).padStart(2, '0')}`;
							const dt = sDate;

							if (vendoSales2[dKey]) {
								points.push({
									date: moment(dt).startOf('day').toDate(),
									total: vendoSales2[dKey].total
								});
							} else {
								points.push({
									date: moment(dt).startOf('day').toDate(),
									total: null
								});
							}
							sDate.setMonth(sDate.getMonth() + 1);
						}

						const vendoName = sales[0].vendo_name;
						return {
							label: vendoName,
							data: points,
							borderWidth: 1,
							tension: 0.4,
							fill: false
						};
					});

					const existingByLabel = keyBy(chart.data.datasets, 'label');
					chart.data.datasets = (newDatasets as any[]).map((newDs) => {
						const existing = existingByLabel[newDs.label];
						if (existing) {
							existing.data = newDs.data;
							return existing;
						}
						return newDs;
					});

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
