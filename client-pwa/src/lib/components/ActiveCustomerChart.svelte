<script lang="ts">
	import Chart from 'chart.js/auto';
	import { onMount, onDestroy } from 'svelte';
	import map from 'lodash/map';
	import groupBy from 'lodash/groupBy';
	import moment from 'moment';
	import minBy from 'lodash/minBy';
	import maxBy from 'lodash/maxBy';
	import keyBy from 'lodash/keyBy';

	import type { ChartConfiguration } from 'chart.js';
	import 'chartjs-adapter-moment';

	let chartData: any = {};
	let canvas: HTMLCanvasElement;
	let intervalId: any;
	let isLoading: boolean = true;
	let chart: Chart;
	let controller: AbortController | undefined = undefined;
	let lastDataHash: string = '';

	interface iDailySale {
		date: string;
		time: string;
		total: number;
		vendo_id: number;
		vendo_name: string;
		average_active_users: number;
	}

	function renderChart(): void {
		chartData = {
			labels: [],
			datasets: []
		};
		const chartConfig: ChartConfiguration = {
			type: 'line',
			data: chartData,
			options: {
				parsing: {
					xAxisKey: 'time',
					yAxisKey: 'users'
				},
				interaction: {
					intersect: false,
					mode: 'index'
				},
				elements: {
					point: {
						radius: 0
					}
				},
				scales: {
					x: {
						type: 'time',
					},
					y: {
						ticks: {
							stepSize: 1
						},
					}
				},
				plugins: {
					tooltip: {
						enabled: true,
						position: 'nearest',
						callbacks: {
							title: function(tooltipItems: any[]) {
								const { raw } = tooltipItems[0];
								return moment(raw.time).format('MMMM DD, Y h:mm A');
							},
						}
					}
				},
			}
		}
		chart = new Chart(canvas, chartConfig);
	}

	function loadChartData(): void {
		controller = new AbortController();
		const signal = controller.signal;
		const from = moment().subtract(1, 'month').format('Y-MM-DD');
		const to = moment().format('Y-MM-DD');
		const request = new Request(
			`/x-api/vendo-status-history?from_date=${from}&to_date=${to}&active_only=true`,
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

					const minTime = minBy(data, (o: any) => o.time)?.time;
					const maxTime = maxBy(data, (o: any) => o.time)?.time;
					const byVendo = groupBy(data, 'vendo_name');
					const newDatasets = map(byVendo, (vendoSales: iDailySale[]) => {
						const points = [];

						const vendoSales2 = keyBy(vendoSales, (o: iDailySale) => (new Date(Date.parse(o.time)).toISOString().substring(0, 16).replace('T', ' ')));

						const sTime = new Date(Date.parse(minTime));
						const eTime = new Date(Date.parse(maxTime));
						while (sTime.getTime() <= eTime.getTime()) {
							const dKey = sTime.toISOString().substring(0, 16).replace('T', ' ');
							const dt = sTime;
							if (vendoSales2[dKey]) {
								points.push({
									time: moment(dt).startOf('hour').toDate(),
									users: vendoSales2[dKey].average_active_users
								});
							} else {
								points.push({
									time: moment(dt).startOf('hour').toDate(),
									users: null
								});
							}
							sTime.setHours(sTime.getHours() + 1);
						}

						const vendoName = vendoSales[0].vendo_name;
						return {
							label: vendoName,
							data: points,
							borderWidth: 1,
							tension: 0.4
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
