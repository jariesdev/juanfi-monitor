<script lang="ts">
	import { page } from '$app/stores';
	import { navCollapsed } from '$lib/store';
	import { navItems, type NavItem } from '$lib/nav';

	interface Props {
		permissions?: string[];
	}
	let { permissions = [] }: Props = $props();

	const visibleItems = $derived(
		navItems.filter((item) => !item.requiredPermission || permissions.includes(item.requiredPermission))
	);

	let openGroups: Record<string, boolean> = $state({});

	function toggleGroup(label: string) {
		openGroups[label] = !openGroups[label];
	}

	function isActive(item: NavItem): boolean {
		if (item.href) return $page.url.pathname === item.href;
		return item.children?.some((c) => $page.url.pathname === c.href) ?? false;
	}

	// Auto-open the group that contains the current page
	$effect(() => {
		visibleItems.forEach((item) => {
			if (item.children?.some((c) => $page.url.pathname === c.href)) {
				openGroups[item.label] = true;
			}
		});
	});
</script>

<aside class="sidebar" class:collapsed={$navCollapsed} aria-label="Main navigation">
	<!-- Header -->
	<div class="sidebar-top">
		{#if !$navCollapsed}
			<span class="brand">VendoReport</span>
		{/if}
		<button
			class="toggle-btn"
			onclick={() => navCollapsed.update((v) => !v)}
			title={$navCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
		>
			<span uk-icon={$navCollapsed ? 'icon: chevron-right' : 'icon: chevron-left'}></span>
		</button>
	</div>

	<!-- Nav items -->
	<nav>
		{#each visibleItems as item}
			{#if item.children}
				<!-- Parent with children -->
				<div class="nav-group" class:collapsed={$navCollapsed}>
					<button
						class="nav-item"
						class:active={isActive(item)}
						onclick={() => { if (!$navCollapsed) toggleGroup(item.label); }}
						title={item.label}
					>
						<span class="nav-icon" uk-icon={`icon: ${item.icon}`}></span>
						{#if !$navCollapsed}
							<span class="nav-label">{item.label}</span>
							<span
								class="nav-caret"
								uk-icon={openGroups[item.label] ? 'icon: chevron-up' : 'icon: chevron-down'}
							></span>
						{/if}
					</button>

					{#if $navCollapsed}
						<!-- Collapsed: CSS hover popout to the right -->
						<div class="popout" role="menu">
							<div class="popout-title">{item.label}</div>
							{#each item.children as child}
								<a
									href={child.href}
									class="popout-item"
									class:active={$page.url.pathname === child.href}
									role="menuitem"
								>
									{child.label}
								</a>
							{/each}
						</div>
					{:else if openGroups[item.label]}
						<!-- Expanded: inline accordion -->
						<div class="nav-children">
							{#each item.children as child}
								<a
									href={child.href}
									class="nav-child"
									class:active={$page.url.pathname === child.href}
								>
									<span class="child-dot"></span>
									{child.label}
								</a>
							{/each}
						</div>
					{/if}
				</div>
			{:else}
				<a
					href={item.href}
					class="nav-item"
					class:active={isActive(item)}
					title={item.label}
				>
					<span class="nav-icon" uk-icon={`icon: ${item.icon}`}></span>
					{#if !$navCollapsed}
						<span class="nav-label">{item.label}</span>
					{/if}
				</a>
			{/if}
		{/each}
	</nav>

	<!-- Logout -->
	<div class="sidebar-footer">
		<form action="/logout" method="POST">
			<button type="submit" class="nav-item logout-btn" title="Logout">
				<span class="nav-icon" uk-icon="icon: sign-out"></span>
				{#if !$navCollapsed}
					<span class="nav-label">Logout</span>
				{/if}
			</button>
		</form>
	</div>
</aside>

<style>
	:root {
		--sidebar-w: 220px;
		--sidebar-collapsed-w: 56px;
	}

	.sidebar {
		position: fixed;
		top: 0;
		left: 0;
		height: 100dvh;
		width: var(--sidebar-w);
		background: #fff;
		border-right: 1px solid #e8e8e8;
		display: flex;
		flex-direction: column;
		transition: width 0.2s ease;
		z-index: 200;
		overflow: visible;
	}

	.sidebar.collapsed {
		width: var(--sidebar-collapsed-w);
	}

	/* Top */
	.sidebar-top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 10px;
		height: 56px;
		border-bottom: 1px solid #e8e8e8;
		flex-shrink: 0;
		overflow: hidden;
	}

	.collapsed .sidebar-top {
		justify-content: center;
		padding: 0;
	}

	.brand {
		font-size: 0.85rem;
		font-weight: 700;
		color: var(--color-theme-1);
		white-space: nowrap;
		padding-left: 4px;
	}

	.toggle-btn {
		background: none;
		border: none;
		cursor: pointer;
		padding: 6px;
		border-radius: 6px;
		color: #666;
		display: flex;
		align-items: center;
		flex-shrink: 0;
		transition: background 0.15s;
	}

	.toggle-btn:hover {
		background: #f2f2f2;
		color: #333;
	}

	/* Nav */
	nav {
		flex: 1;
		overflow-y: auto;
		overflow-x: visible;
		padding: 8px 0;
		scrollbar-width: thin;
	}

	.sidebar.collapsed nav {
		overflow: visible;
	}

	.nav-item {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 9px 10px;
		margin: 1px 6px;
		width: calc(100% - 12px);
		color: #444;
		font-size: 0.875rem;
		font-weight: 500;
		text-decoration: none;
		border: none;
		background: none;
		cursor: pointer;
		border-radius: 6px;
		white-space: nowrap;
		transition: background 0.15s, color 0.15s;
		box-sizing: border-box;
	}

	.nav-item:hover {
		background: #f5f5f5;
		color: #111;
		text-decoration: none;
	}

	.nav-item.active {
		background: #fff0ed;
		color: var(--color-theme-1);
	}

	.nav-icon {
		flex-shrink: 0;
		width: 20px;
		height: 20px;
		display: flex;
		align-items: center;
	}

	.nav-label {
		flex: 1;
		text-align: left;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.nav-caret {
		flex-shrink: 0;
		width: 16px;
		height: 16px;
		opacity: 0.4;
	}

	/* Accordion children */
	.nav-children {
		display: flex;
		flex-direction: column;
		padding: 2px 0 4px 36px;
	}

	.nav-child {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 7px 10px;
		margin: 1px 6px;
		font-size: 0.82rem;
		color: #555;
		text-decoration: none;
		border-radius: 5px;
		transition: background 0.15s, color 0.15s;
	}

	.nav-child:hover {
		background: #f5f5f5;
		color: #111;
		text-decoration: none;
	}

	.nav-child.active {
		color: var(--color-theme-1);
		font-weight: 600;
	}

	.child-dot {
		width: 5px;
		height: 5px;
		border-radius: 50%;
		background: currentColor;
		opacity: 0.4;
		flex-shrink: 0;
	}

	/* Collapsed hover popout */
	.nav-group {
		position: relative;
	}

	/* Invisible bridge so hover is not lost between icon and popout */
	.nav-group.collapsed::after {
		content: '';
		position: absolute;
		left: 100%;
		top: 0;
		width: 8px;
		height: 100%;
	}

	.popout {
		display: none;
		position: absolute;
		left: 100%;
		padding-left: 6px;
		top: 0;
		background: transparent;
		border: none;
		box-shadow: none;
		flex-direction: column;
		min-width: 166px;
		z-index: 300;
		overflow: visible;
	}

	.nav-group.collapsed:hover .popout {
		display: flex;
	}

	.popout-title,
	.popout-item {
		background: #fff;
		border-left: 1px solid #e8e8e8;
		border-right: 1px solid #e8e8e8;
	}

	.popout-title {
		padding: 8px 14px 6px;
		font-size: 0.72rem;
		font-weight: 700;
		color: #999;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		border-top: 1px solid #e8e8e8;
		border-bottom: 1px solid #f0f0f0;
		border-radius: 8px 8px 0 0;
		box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.04);
	}

	.popout-item:last-child {
		border-bottom: 1px solid #e8e8e8;
		border-radius: 0 0 8px 8px;
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
	}

	.popout-item {
		padding: 10px 14px;
		font-size: 0.875rem;
		color: #444;
		text-decoration: none;
		transition: background 0.15s;
	}

	.popout-item:hover {
		background: #f5f5f5;
		text-decoration: none;
	}

	.popout-item.active {
		color: var(--color-theme-1);
		font-weight: 600;
	}

	/* Footer / logout */
	.sidebar-footer {
		border-top: 1px solid #e8e8e8;
		padding: 6px 0;
		flex-shrink: 0;
	}

	.sidebar-footer form {
		display: contents;
	}

	.logout-btn {
		color: #c0392b;
	}

	.logout-btn:hover {
		background: #fff0ee;
		color: #c0392b;
	}
</style>
