<script lang="ts">
	let isOnline: boolean = $state(true);
</script>

<svelte:window bind:online={isOnline} />

{#if !isOnline}
	<div
		class="uk-position-bottom-center uk-box-shadow-small uk-animation-fade uk-alert-danger" uk-alert
	>
			Not connected to the internet. Data may not accurate.
	</div>
{/if}
