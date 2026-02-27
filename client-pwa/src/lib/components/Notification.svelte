<script lang="ts">
	interface iProps {
		message: string,
		onHide: Function
	}

	const {message, onHide}: iProps = $props()

	let timer: number = 10
	let intervalId: number
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

<div class="uk-alert-primary uk-animation-slide-bottom-small uk-position-bottom-center uk-padding-small" class:uk-hidden={!message} uk-alert>{message}</div>
