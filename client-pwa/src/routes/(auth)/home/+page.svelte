<svelte:head>
	<title>Dashboard</title>
</svelte:head>

<script lang="ts">
	import ActiveUsersWidget from '$lib/components/ActiveUsersWidget.svelte';
	import DailySaleChart from '$lib/components/DailySaleChart.svelte';
	import ActiveCustomerChart from '$lib/components/ActiveCustomerChart.svelte';
	import VendoCapacityChart from '$lib/components/VendoCapacityChart.svelte';
	import MonthlySaleChart from '$lib/components/MonthlySaleChart.svelte';
	import { hasPermission } from '$lib/acl.svelte.js';
	import { onMount } from 'svelte';

	onMount(() => {
		window.Notification.requestPermission();
	});
</script>

<section class="uk-section">
	<ActiveUsersWidget />

	{#if hasPermission('sales')}
		<h3 class="uk-text-light uk-text-center">Daily Sales</h3>
		<DailySaleChart />
	{/if}
	<h3 class="uk-text-light uk-text-center">Vendo Capacity Counter</h3>
	<VendoCapacityChart />
	{#if hasPermission('sales')}
		<h3 class="uk-text-light uk-text-center">Monthly Sales</h3>
		<MonthlySaleChart />
	{/if}
	<h3 class="uk-text-light uk-text-center">Active User History</h3>
	<ActiveCustomerChart />
</section>
