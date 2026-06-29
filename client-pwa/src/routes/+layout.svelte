<script>
	import './styles.css';
	import '$lib/css/uikit.css';
	import '$lib/css/override.css';
	import '$lib/js/uikit.min.js';
	import '$lib/js/uikit-icons.min.js';
	import OfflineAlert from '$lib/components/OfflineAlert.svelte';
	import { onMount } from 'svelte';

	onMount(() => {
		import('$lib/js/uikit.min.js');
	});
</script>

<slot />
<OfflineAlert />
