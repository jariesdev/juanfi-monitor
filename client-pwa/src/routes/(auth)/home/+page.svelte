<svelte:head>
	<title>Dashboard</title>
</svelte:head>

<script>
	import DailySaleChart from '$lib/components/DailySaleChart.svelte';
	import ActiveCustomerChart from '$lib/components/ActiveCustomerChart.svelte';
	import VendoCapacityChart from '$lib/components/VendoCapacityChart.svelte';
	import MonthlySaleChart from '$lib/components/MonthlySaleChart.svelte';
	import {onMount} from "svelte";

	onMount(() => {
		window.Notification.requestPermission()
	})
</script>

<section class="uk-section">
	<h3 class="uk-text-light uk-text-center">Daily Sales</h3>
	<DailySaleChart />
	<h3 class="uk-text-light uk-text-center">Vendo Capacity Counter</h3>
	<VendoCapacityChart />
	<h3 class="uk-text-light uk-text-center">Monthly Sales</h3>
	<MonthlySaleChart />
	<h3 class="uk-text-light uk-text-center">Active User History</h3>
	<ActiveCustomerChart />
</section>
