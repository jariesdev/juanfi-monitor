<script lang="ts">
	interface Props {
		online: boolean;
		active: boolean;
		/** Live data-pull progress 0..100, or null/undefined when idle. */
		progress?: number | null;
	}

	const { online, active, progress = null }: Props = $props();

	// The travelling "snake" only runs for active, online machines that are
	// currently being pulled by the scheduler.
	const showSnake = $derived(active && online && progress != null);
</script>

<span class="badge-wrap" class:refreshing={showSnake} style="--progress:{progress ?? 0}">
	<span
		class="status-badge"
		class:online={active && online}
		class:offline={active && !online}
		class:inactive={!active}
	>
		{online ? 'Online' : 'Offline'}
	</span>
</span>

<style>
	@property --progress {
		syntax: '<number>';
		initial-value: 0;
		inherits: true;
	}

	.badge-wrap {
		position: relative;
		display: inline-block;
		border-radius: 12px;
		--progress: 0;
		/* Snake colours: darker shade of the online badge background + brighter head. */
		--snake: #16a34a;
		--snake-head: #4ade80;
	}

	/* Tween the arc between scheduler milestones so the snake glides. */
	.badge-wrap.refreshing {
		transition: --progress 0.35s linear;
	}

	/* Masked conic-gradient ring: a darker arc of the badge background grows
	   clockwise from top-center (0%) to --progress%, with a brighter leading head.
	   At 100% the ring is full, closing back at top-center. */
	.badge-wrap.refreshing::before {
		content: '';
		position: absolute;
		inset: -2px;
		border-radius: inherit;
		padding: 2px;
		background: conic-gradient(
			var(--snake) 0,
			var(--snake) calc(var(--progress) * 1% - 6%),
			var(--snake-head) calc(var(--progress) * 1%),
			transparent 0
		);
		-webkit-mask:
			linear-gradient(#000 0 0) content-box,
			linear-gradient(#000 0 0);
		mask:
			linear-gradient(#000 0 0) content-box,
			linear-gradient(#000 0 0);
		-webkit-mask-composite: xor;
		mask-composite: exclude;
		pointer-events: none;
	}

	.status-badge {
		display: inline-block;
		padding: 2px 8px;
		border-radius: 12px;
		font-size: 0.78rem;
		font-weight: 600;
	}

	.status-badge.online {
		background: #dcfce7;
		color: #166534;
	}

	.status-badge.offline {
		background: #fee2e2;
		color: #991b1b;
	}

	.status-badge.inactive {
		background: #e5e7eb;
		color: #6b7280;
	}
</style>
