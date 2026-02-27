<script>
	import Header from '$lib/components/Header.svelte';
	import Notifications from "$lib/components/Notifications.svelte";
</script>

<div class="app-main-inner">
	<Header />

	<div class="app-content">
		<slot />
	</div>

	<Notifications />
</div>

