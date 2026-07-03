<script lang="ts">
	import Sidebar from '$lib/components/Sidebar.svelte';
	import BottomNav from '$lib/components/BottomNav.svelte';
	import Notifications from '$lib/components/Notifications.svelte';
	import Toast from '$lib/components/Toast.svelte';
	import { setPermissions } from '$lib/acl.svelte.js';
	import { navCollapsed } from '$lib/store';
	import type { Snippet } from 'svelte';

	interface LayoutData {
		user: { id?: number } | null;
		permissions: string[];
	}

	let { data, children }: { data: LayoutData; children: Snippet } = $props();

	$effect(() => {
		setPermissions(data.permissions);
	});
</script>

<!-- Desktop sidebar (hidden on mobile via CSS) -->
<div class="desktop-nav">
	<Sidebar />
</div>

<!-- Main content — offset matches sidebar width on desktop -->
<main
	class="auth-content"
	class:sidebar-expanded={!$navCollapsed}
	class:sidebar-collapsed={$navCollapsed}
>
	<div class="content-inner uk-margin-auto-left uk-margin-auto-right">
		{@render children()}
	</div>
</main>

<!-- Mobile bottom nav (hidden on desktop via CSS) -->
<div class="mobile-nav">
	<BottomNav />
</div>

<Notifications
	currentUserId={data.permissions?.includes('users') ? null : (data.user?.id ?? null)}
/>
<Toast />

<style>
	/* Desktop sidebar wrapper — visible only on desktop */
	.desktop-nav {
		display: none;
	}

	/* Mobile bottom nav — visible only on mobile */
	.mobile-nav {
		display: block;
	}

	/* Content area */
	.auth-content {
		min-height: 100dvh;
		transition: margin-left 0.2s ease;
		/* Mobile: add bottom padding so content isn't hidden behind bottom nav */
		padding-bottom: calc(56px + env(safe-area-inset-bottom) + 8px);
	}

	.content-inner {
		padding: 16px;
	}

	@media (min-width: 768px) {
		.desktop-nav {
			display: block;
		}

		.mobile-nav {
			display: none;
		}

		.auth-content {
			padding-bottom: 0;
		}

		.auth-content.sidebar-expanded {
			margin-left: 220px;
		}

		.auth-content.sidebar-collapsed {
			margin-left: 56px;
		}

		.content-inner {
			padding: 24px;
			max-width: 1400px;
		}
	}
</style>
