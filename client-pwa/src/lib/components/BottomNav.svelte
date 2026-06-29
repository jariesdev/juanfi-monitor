<script lang="ts">
	import { page } from '$app/stores';
	import { getVisibleNavItems, type NavItem } from '$lib/nav';

	interface Props {
		permissions?: string[];
	}
	let { permissions = [] }: Props = $props();

	const visibleItems = $derived(getVisibleNavItems(permissions));

	let activeSheet: NavItem | null = $state(null);
	let tooltip: { label: string; x: number } | null = $state(null);
	let tooltipTimer: ReturnType<typeof setTimeout>;

	function onNavClick(item: NavItem) {
		if (item.children) {
			activeSheet = activeSheet?.label === item.label ? null : item;
		}
	}

	function closeSheet() {
		activeSheet = null;
	}

	function isActive(item: NavItem): boolean {
		if (item.href) return $page.url.pathname === item.href;
		return item.children?.some((c) => $page.url.pathname === c.href) ?? false;
	}

	function touchStart(label: string, event: TouchEvent) {
		const touch = event.touches[0];
		const x = touch.clientX;
		tooltipTimer = setTimeout(() => {
			tooltip = { label, x };
		}, 500);
	}

	function touchEnd() {
		clearTimeout(tooltipTimer);
		setTimeout(() => {
			tooltip = null;
		}, 1000);
	}

	function touchMove() {
		clearTimeout(tooltipTimer);
		tooltip = null;
	}
</script>

<!-- Sheet backdrop -->
{#if activeSheet}
	<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
	<div class="sheet-backdrop" onclick={closeSheet}></div>
	<div class="sheet" role="dialog" aria-label={activeSheet.label}>
		<div class="sheet-header">
			<span class="sheet-title">{activeSheet.label}</span>
			<button class="sheet-close" onclick={closeSheet} aria-label="Close">
				<span uk-icon="icon: close"></span>
			</button>
		</div>
		<div class="sheet-body">
			{#each activeSheet.children ?? [] as child}
				<a
					href={child.href}
					class="sheet-item"
					class:active={$page.url.pathname === child.href}
					onclick={closeSheet}
				>
					{child.label}
				</a>
			{/each}
		</div>
	</div>
{/if}

<!-- Long-press tooltip -->
{#if tooltip}
	<div
		class="lp-tooltip"
		style="left: {Math.min(tooltip.x, (typeof window !== 'undefined' ? window.innerWidth : 375) - 80)}px"
	>
		{tooltip.label}
	</div>
{/if}

<!-- Bottom nav bar -->
<nav class="bottom-nav" aria-label="Main navigation">
	{#each visibleItems as item}
		{#if item.children}
			<button
				class="nav-btn"
				class:active={isActive(item) || activeSheet?.label === item.label}
				onclick={() => onNavClick(item)}
				ontouchstart={(e) => touchStart(item.label, e)}
				ontouchend={touchEnd}
				ontouchmove={touchMove}
				aria-label={item.label}
			>
				<span uk-icon={`icon: ${item.icon}`}></span>
				{#if isActive(item) || activeSheet?.label === item.label}
					<span class="active-dot"></span>
				{/if}
			</button>
		{:else}
			<a
				href={item.href}
				class="nav-btn"
				class:active={isActive(item)}
				ontouchstart={(e) => touchStart(item.label, e)}
				ontouchend={touchEnd}
				ontouchmove={touchMove}
				aria-label={item.label}
			>
				<span uk-icon={`icon: ${item.icon}`}></span>
				{#if isActive(item)}
					<span class="active-dot"></span>
				{/if}
			</a>
		{/if}
	{/each}

	<!-- Logout -->
	<form action="/logout" method="POST" class="nav-form">
		<button
			type="submit"
			class="nav-btn"
			ontouchstart={(e) => touchStart('Logout', e)}
			ontouchend={touchEnd}
			ontouchmove={touchMove}
			aria-label="Logout"
		>
			<span uk-icon="icon: sign-out"></span>
		</button>
	</form>
</nav>

<style>
	/* Bottom nav */
	.bottom-nav {
		position: fixed;
		bottom: 0;
		left: 0;
		right: 0;
		height: calc(56px + env(safe-area-inset-bottom));
		padding-bottom: env(safe-area-inset-bottom);
		background: #fff;
		border-top: 1px solid #e8e8e8;
		display: flex;
		align-items: center;
		z-index: 200;
		box-shadow: 0 -1px 8px rgba(0, 0, 0, 0.06);
	}

	.nav-btn {
		flex: 1;
		position: relative;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 56px;
		color: #999;
		text-decoration: none;
		background: none;
		border: none;
		cursor: pointer;
		-webkit-tap-highlight-color: transparent;
		transition: color 0.15s;
	}

	.nav-btn.active {
		color: var(--color-theme-1);
	}

	.nav-btn:active {
		background: rgba(0, 0, 0, 0.04);
	}

	.active-dot {
		position: absolute;
		top: 6px;
		left: 50%;
		transform: translateX(-50%);
		width: 28px;
		height: 3px;
		background: var(--color-theme-1);
		border-radius: 0 0 3px 3px;
	}

	.nav-form {
		flex: 1;
		display: flex;
	}

	.nav-form .nav-btn {
		flex: 1;
	}

	/* Sub-menu sheet */
	.sheet-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.35);
		z-index: 300;
	}

	.sheet {
		position: fixed;
		left: 0;
		right: 0;
		bottom: calc(56px + env(safe-area-inset-bottom));
		background: #fff;
		border-radius: 14px 14px 0 0;
		z-index: 301;
		overflow: hidden;
		animation: slideUp 0.22s ease;
		box-shadow: 0 -4px 24px rgba(0, 0, 0, 0.12);
	}

	@keyframes slideUp {
		from {
			transform: translateY(100%);
			opacity: 0;
		}
		to {
			transform: translateY(0);
			opacity: 1;
		}
	}

	.sheet-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 16px 20px 12px;
		border-bottom: 1px solid #f0f0f0;
	}

	.sheet-title {
		font-size: 0.9rem;
		font-weight: 700;
		color: #222;
	}

	.sheet-close {
		background: none;
		border: none;
		cursor: pointer;
		color: #888;
		padding: 4px;
		border-radius: 6px;
		display: flex;
		align-items: center;
	}

	.sheet-body {
		padding: 8px 0 12px;
	}

	.sheet-item {
		display: flex;
		align-items: center;
		padding: 15px 24px;
		font-size: 1rem;
		color: #333;
		text-decoration: none;
		transition: background 0.15s;
	}

	.sheet-item:hover {
		background: #f8f8f8;
		text-decoration: none;
	}

	.sheet-item.active {
		color: var(--color-theme-1);
		font-weight: 600;
	}

	/* Long-press tooltip */
	.lp-tooltip {
		position: fixed;
		transform: translateX(-50%);
		bottom: calc(64px + env(safe-area-inset-bottom));
		background: rgba(0, 0, 0, 0.72);
		color: #fff;
		font-size: 0.78rem;
		padding: 5px 10px;
		border-radius: 5px;
		z-index: 400;
		pointer-events: none;
		white-space: nowrap;
		animation: fadeIn 0.15s ease;
	}

	@keyframes fadeIn {
		from { opacity: 0; transform: translateX(-50%) translateY(4px); }
		to { opacity: 1; transform: translateX(-50%) translateY(0); }
	}
</style>
