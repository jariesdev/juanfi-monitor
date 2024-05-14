<script>
	import welcome from '$lib/images/svelte-welcome.webp';
	import welcome_fallback from '$lib/images/svelte-welcome.png';
	import DailySaleChart from "$lib/components/DailySaleChart.svelte";
	import VendoMonthlyUserChart from "$lib/components/VendoMonthlyUserChart.svelte";
</script>

<svelte:head>
	<title>Home</title>
	<meta name="description" content="Vendor Report App" />
</svelte:head>

<section>
	<h1>
		<span class="welcome">
			<picture>
				<source srcset={welcome} type="image/webp" />
				<img src={welcome_fallback} alt="Welcome" />
			</picture>
		</span>
	</h1>


	<h3 class="uk-text-light uk-text-center">
		Daily Sales
	</h3>
	<DailySaleChart></DailySaleChart>
	<h3 class="uk-text-light uk-text-center">
		Active User History
	</h3>
	<VendoMonthlyUserChart></VendoMonthlyUserChart>
</section>
