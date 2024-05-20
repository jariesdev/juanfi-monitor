<script>
	import welcome from '$lib/images/svelte-welcome.webp';
	import welcome_fallback from '$lib/images/svelte-welcome.png';
	import DailySaleChart from '$lib/components/DailySaleChart.svelte';
	import VendoMonthlyUserChart from '$lib/components/VendoMonthlyUserChart.svelte';
	import VendoCapacityChart from '$lib/components/VendoCapacityChart.svelte';
</script>

<svelte:head>
	<title>Home</title>
	<meta name="description" content="Vendor Report App" />
</svelte:head>

<section class="uk-section">
	<h3 class="uk-text-light uk-text-center">Daily Sales</h3>
	<DailySaleChart />
	<h3 class="uk-text-light uk-text-center">Vendo Capacity Counter</h3>
	<VendoCapacityChart />
	<h3 class="uk-text-light uk-text-center">Active User History</h3>
	<VendoMonthlyUserChart />
</section>
