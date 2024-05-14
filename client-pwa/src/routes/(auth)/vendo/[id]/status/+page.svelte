<script lang="ts">
	import SystemStatus from './SystemStatus.svelte';

	export let data;
</script>

<div class="uk-section">
	<div class="uk-container">
		<SystemStatus bind:vendoId={data.id} />
	</div>
</div>
