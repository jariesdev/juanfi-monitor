<script>
</script>

<div class="guest-layout">
	<slot />
	<footer>
		<p class="uk-text-italic uk-text-muted uk-text-small uk-text-light">Vendo Monitoring</p>
	</footer>
</div>

<style>
	.guest-layout {
		display: flex;
		flex-direction: column;
		min-height: 100dvh;
		width: 100%;
		max-width: 64rem;
		margin: 0 auto;
		padding: 0 16px;
		box-sizing: border-box;
	}

	footer {
		display: flex;
		justify-content: center;
		align-items: center;
		padding: 16px 0;
		margin-top: auto;
	}
</style>
