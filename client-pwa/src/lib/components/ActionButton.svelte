<script lang="ts">
	interface Props {
		/** uk-icon name, e.g. "download". */
		icon: string;
		/** Visible label; hidden on mobile, where the button collapses to icon-only. */
		label: string;
		/** When set, the button renders as an anchor link instead of a <button>. */
		href?: string;
		onclick?: () => void;
		disabled?: boolean;
		/** Filled accent styling for the primary action. */
		primary?: boolean;
		/** Tooltip / accessible name; defaults to the label. */
		title?: string;
		type?: 'button' | 'submit';
	}

	const {
		icon,
		label,
		href,
		onclick,
		disabled = false,
		primary = false,
		title,
		type = 'button'
	}: Props = $props();
</script>

{#if href}
	<a class="action-btn" class:action-btn-primary={primary} {href} title={title ?? label}>
		<span uk-icon={`icon: ${icon}; ratio: 0.8`}></span>
		<span class="btn-label">{label}</span>
	</a>
{:else}
	<button
		class="action-btn"
		class:action-btn-primary={primary}
		{type}
		{onclick}
		{disabled}
		title={title ?? label}
	>
		<span uk-icon={`icon: ${icon}; ratio: 0.8`}></span>
		<span class="btn-label">{label}</span>
	</button>
{/if}

<style>
	.action-btn {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 6px 12px;
		border: 1px solid #e8e8e8;
		border-radius: 7px;
		background: #fff;
		color: #555;
		font-size: 0.78rem;
		font-weight: 600;
		font-family: inherit;
		text-decoration: none;
		cursor: pointer;
		transition:
			background 0.15s,
			opacity 0.15s;
	}

	.action-btn:hover:not(:disabled) {
		background: #f5f5f5;
	}

	.action-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.action-btn-primary {
		background: var(--color-theme-1);
		border-color: var(--color-theme-1);
		color: #fff;
	}

	.action-btn-primary:hover:not(:disabled) {
		opacity: 0.88;
		background: var(--color-theme-1);
	}

	/* On mobile, collapse to icon-only. */
	@media (max-width: 640px) {
		.btn-label {
			display: none;
		}

		.action-btn {
			padding: 6px 10px;
		}
	}
</style>
