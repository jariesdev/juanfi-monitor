<script lang="ts">
	import { Chart, type ChartConfiguration } from 'chart.js';
	import { onMount, onDestroy } from 'svelte';
	import map from 'lodash/map';
	import groupBy from 'lodash/groupBy';
	import keyBy from 'lodash/keyBy';
	import { baseApiUrl } from '$lib/env';
	import moment from 'moment';
	import 'chartjs-adapter-moment';

	let chartData: any[] = [];
	let canvas: HTMLCanvasElement;
	let intervalId: any;
	let isLoading: boolean = false;
	let chart: Chart;
	let controller: AbortController | undefined = undefined;

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
							title: function(tooltipItems: TooltipItem[]) {
								const { raw } = tooltipItems[0];
								return moment(raw.date).format('MMMM Y');
							},
							footer: function(tooltipItems: TooltipItem[]) {
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
						ticks: {
							beginAtZero: true,
							stepSize: 1,
							callback: function(value: string) {
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
					const byVendo = groupBy(data, 'vendo_id');
					const datasets = map(byVendo, (sales: iMonthlySale[]) => {
						const data = [];
						const vendoSales2 = keyBy(sales, (o: iMonthlySale) => o.month);
						const sDate = new Date(fromDate);
						const eDate = new Date(toDate);
						while (sDate <= eDate) {
							const dKey = `${sDate.getFullYear()}-${String(sDate.getMonth() + 1).padStart(2, '0')}`;
							const dt = sDate;

							if (vendoSales2[dKey]) {
								data.push({
									date: moment(dt).startOf('day').toDate(),
									total: vendoSales2[dKey].total
								});
							} else {
								data.push({
									date: moment(dt).startOf('day').toDate(),
									total: null
								});
							}
							sDate.setMonth(sDate.getMonth() + 1);
						}

						const vendoName = sales[0].vendo_name;
						return {
							label: vendoName,
							data: data,
							borderWidth: 1,
							tension: 0.4,
							fill: false
						};
					});

					// const startDate = moment(fromDate).startOf('day');
					// chart.data.labels = [];
					// while (startDate.isSameOrBefore(toDate)) {
					// 	chart.data.labels.push(startDate.toDate());
					// 	startDate.add(1, 'day');
					// }
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
