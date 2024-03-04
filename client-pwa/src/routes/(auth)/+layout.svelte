<script>
	import '../styles.css';
	import '$lib/css/uikit.css'
	import '$lib/js/uikit.min.js'
	import '$lib/js/uikit-icons.min.js'
	import OfflineAlert from "$lib/components/OfflineAlert.svelte";
	import Header from "$lib/components/Header.svelte";
</script>

<div class="app-main">
	<Header />

	<div class="app-content">
		<slot />
	</div>

</div>
