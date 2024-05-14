<script lang="ts">
	import Chart from 'chart.js/auto';
	import { onMount, onDestroy } from 'svelte';
	import { apiUrl } from '$lib/store';
	import map from 'lodash/map';
	import minBy from 'lodash/minBy';
	import maxBy from 'lodash/maxBy';
	import groupBy from 'lodash/groupBy';
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
					xAxisKey: 'date',
					yAxisKey: 'total'
				},
				elements: {
					point: {
						radius: 2.5
					}
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
		controller = new AbortController();
		const signal = controller.signal;
		const request = new Request(`${baseApiUrl}/daily-sales`, { method: 'GET', signal: signal });
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
					const minDate = minBy(data, (o: any) => o.date)?.date;
					const maxDate = maxBy(data, (o: any) => o.date)?.date;
					const byVendo = groupBy(data, 'vendo_id');
					const datasets = map(byVendo, (vendoSales: iDailySale[]) => {
						const data = [];
						const vendoSales2 = keyBy(vendoSales, (o: iDailySale) => o.date);
						const sDate = new Date(Date.parse(minDate));
						const eDate = new Date(Date.parse(maxDate));
						while (sDate.getTime() <= eDate.getTime()) {
							const dKey = sDate.toISOString().split('T')[0];
							const dt = sDate;
							const m = dt.getMonth() + 1;
							const dy = dt.getDate();
							const df = String(dy).padStart(2, '0');
							const mf = String(m).padStart(2, '0');

							if (vendoSales2[dKey]) {
								data.push({
									date: `${mf}-${df}`,
									total: vendoSales2[dKey].total
								});
							} else {
								data.push({
									date: `${mf}-${df}`,
									total: null
								});
							}
							sDate.setDate(sDate.getDate() + 1);
						}

						// const data = map(vendoSales, (d: iDailySale) => {
						//     const dt = new Date(Date.parse(d.date))
						//     const m = dt.getMonth() + 1
						//     const dy = dt.getDate()
						//     const df = String(dy).padStart(2, '0')
						//     const mf = String(m).padStart(2, '0')
						//     return {
						//         date: d.date,
						//         total: d.total
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
