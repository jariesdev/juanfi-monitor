<script lang="ts">
	import { toast } from '$lib/store';

	let visible = $state(false);
	let message = $state('');
	let type: 'success' | 'error' = $state('success');
	let timerId: ReturnType<typeof setTimeout> | undefined;

	toast.subscribe((value) => {
		if (!value) return;
		message = value.message;
		type = value.type;
		visible = true;
		clearTimeout(timerId);
		timerId = setTimeout(() => {
			visible = false;
			toast.set(null);
		}, 4000);
	});
</script>

{#if visible}
	<div class="toast toast-{type}" role="alert">
		<span class="toast-icon" uk-icon="icon: {type === 'success' ? 'check' : 'warning'}; ratio: 0.9"></span>
		<span class="toast-message">{message}</span>
		<button
			class="toast-close"
			type="button"
			aria-label="Dismiss"
			onclick={() => { visible = false; toast.set(null); }}
		>
			<span uk-icon="icon: close; ratio: 0.8"></span>
		</button>
	</div>
{/if}

<style>
	.toast {
		position: fixed;
		bottom: 80px;
		left: 50%;
		transform: translateX(-50%);
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px 18px;
		border-radius: 10px;
		font-size: 0.85rem;
		font-weight: 600;
		z-index: 9999;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
		animation: slide-up 0.2s ease;
		max-width: calc(100vw - 32px);
	}

	.toast-success {
		background: #1a7a4a;
		color: #fff;
	}

	.toast-error {
		background: #c0392b;
		color: #fff;
	}

	.toast-icon {
		display: flex;
		align-items: center;
		flex-shrink: 0;
	}

	.toast-message {
		flex: 1;
	}

	.toast-close {
		background: none;
		border: none;
		color: inherit;
		cursor: pointer;
		padding: 0;
		display: flex;
		align-items: center;
		opacity: 0.7;
		flex-shrink: 0;
	}

	.toast-close:hover {
		opacity: 1;
	}

	@keyframes slide-up {
		from { opacity: 0; transform: translateX(-50%) translateY(12px); }
		to   { opacity: 1; transform: translateX(-50%) translateY(0); }
	}

	@media (min-width: 768px) {
		.toast {
			bottom: 32px;
		}
	}
</style>
