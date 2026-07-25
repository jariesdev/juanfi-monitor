<script>
	import './styles.css';
	import 'uikit/dist/css/uikit.min.css';
	import '$lib/css/override.css';
	import UIkit from 'uikit';
	import UIkitIcons from 'uikit/dist/js/uikit-icons.js';
	import OfflineAlert from '$lib/components/OfflineAlert.svelte';
	import { onMount } from 'svelte';

	UIkit.use(UIkitIcons);

	onMount(() => {
		UIkit.update();
	});
</script>

<slot />
<OfflineAlert />
