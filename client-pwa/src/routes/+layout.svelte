<script>
	import './styles.css';
	import '$lib/css/uikit.css';
	import '$lib/js/uikit.min.js';
	import '$lib/js/uikit-icons.min.js';
	import OfflineAlert from '$lib/components/OfflineAlert.svelte';
</script>

<div class="app">
	<main>
		<slot />
	</main>

	<footer>
		<p class="uk-text-italic uk-text-muted uk-text-small uk-text-light">Vendo Reports</p>
	</footer>
	<OfflineAlert />
</div>

<style>
	.app {
		display: flex;
		flex-direction: column;
		min-height: 100vh;
	}

	main {
		flex: 1;
		display: flex;
		flex-direction: column;
		width: 100%;
		max-width: 64rem;
		margin: 0 auto;
		box-sizing: border-box;
	}

	footer {
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		padding: 12px;
	}

	@media (min-width: 480px) {
		footer {
			padding: 12px 0;
		}
	}
</style>
