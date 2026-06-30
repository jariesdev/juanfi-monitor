<script lang="ts">
	interface iProps {
		message: string,
		onHide: Function
	}

	const {message, onHide}: iProps = $props()

	let timer: number = 10
	let intervalId: ReturnType<typeof setTimeout>
	let isVisible: boolean = $derived(!!message)

	$effect(() => {
		if (message) {
			intervalId = setTimeout(() => {
				clearTimeout(intervalId)

				onHide()
			}, timer * 1000)
		}
	})
</script>

	<div class="uk-alert-primary uk-animation-slide-bottom-small uk-padding-small notification-toast" class:uk-hidden={!message} uk-alert>{message}</div>

<style>
	.notification-toast {
		position: fixed;
		bottom: 20px;
		left: 50%;
		transform: translateX(-50%);
		z-index: 1000;
		max-width: 90vw;
	}
</style>
