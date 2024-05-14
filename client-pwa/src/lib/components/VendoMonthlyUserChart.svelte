<script lang="ts">
	import Chart from 'chart.js/auto';
	import { onMount, onDestroy } from 'svelte';
	import { apiUrl } from '$lib/store';
	import map from 'lodash/map';
	import groupBy from 'lodash/groupBy';
	import moment from 'moment';
	import minBy from 'lodash/minBy';
	import maxBy from 'lodash/maxBy';
	import keyBy from 'lodash/keyBy';
	import { baseApiUrl } from '$lib/env';

	let chartData: any[] = [];
	let canvas: HTMLCanvasElement;
	let intervalId: any;
	let isLoading: boolean = false;
	let chart: Chart;
	let controller: AbortController | undefined = undefined;

	interface iDailySale {
		date: string;
		total: number;
		vendo_id: number;
		vendo_name: string;
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
					xAxisKey: 'time',
					yAxisKey: 'users'
				},
				elements: {
					point: {
						radius: 0
					}
				}
			}
		});
	}

	function loadChartData(): void {
		controller = new AbortController();
		const signal = controller.signal;
		const from = moment().subtract(1, 'month').format('Y-MM-DD');
		const to = moment().format('Y-MM-DD');
		const request = new Request(
			`${baseApiUrl}/vendo-status-history?from_date=${from}&to_date=${to}`,
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
					const minTime = minBy(data, (o: any) => o.time)?.time;
					const maxTime = maxBy(data, (o: any) => o.time)?.time;
					const byVendo = groupBy(data, 'vendo_name');
					const datasets = map(byVendo, (vendoSales: iDailySale[]) => {
						const data = [];

						const vendoSales2 = keyBy(vendoSales, (o: iDailySale) => o.time);
						const sTime = new Date(Date.parse(minTime));
						const eTime = new Date(Date.parse(maxTime));
						while (sTime.getTime() <= eTime.getTime()) {
							const dKey = sTime.toISOString().substring(0, 16).replace('T', ' ');
							const dt = sTime;
							if (vendoSales2[dKey]) {
								data.push({
									time: moment(dt).format('MM-DD ha'),
									users: vendoSales2[dKey].average_active_users
								});
							} else {
								data.push({
									time: moment(dt).format('MM-DD ha'),
									users: null
								});
							}
							sTime.setHours(sTime.getHours() + 1);
						}

						// const data = map(vendoSales, (d: any) => {
						//     const dt = new Date(Date.parse(d.time))
						//     return {
						//         time: moment(dt).format('MM-DD ha'),
						//         users: d.average_active_users
						//     }
						// });
						const vendoName = vendoSales[0].vendo_name;
						return {
							label: vendoName,
							data: data,
							borderWidth: 1,
							tension: 0.4
						};
					});

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
