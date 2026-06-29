<svelte:head>
	<title>Active Users</title>
</svelte:head>

<script lang="ts">
	import { onMount, onDestroy } from 'svelte';

	interface ActiveUser {
		node_id: string;
		user: string;
		mac_address: string;
		session_left: string;
	}

	const { data } = $props();
	const vendoId = $derived(data.id);

	let users: ActiveUser[] = $state([]);
	let isLoading = $state(true);
	let controller: AbortController | undefined;

	function loadData() {
		controller?.abort();
		controller = new AbortController();
		isLoading = true;

		fetch(`/x-api/vendo-machines/${vendoId}/active-users`, { signal: controller.signal })
			.then((r) => (r.ok ? r.json() : Promise.reject(r.statusText)))
			.then(({ data }) => {
				users = data ?? [];
			})
			.catch(() => {})
			.finally(() => {
				isLoading = false;
			});
	}

	onMount(loadData);
	onDestroy(() => controller?.abort());
</script>

<div class="uk-section">
	<div class="uk-container">
		<div class="card">
			<div class="card-header">
				<span class="header-icon" uk-icon="icon: users; ratio: 1"></span>
				<div class="header-left">
					<span class="card-title">Active Users</span>
					<button class="refresh-btn" onclick={loadData} disabled={isLoading} title="Refresh">
						<span uk-icon="icon: refresh; ratio: 0.9"></span>
					</button>
				</div>
			</div>

			{#if isLoading}
				<table class="users-table">
					<thead>
						<tr>
							<th>User</th>
							<th>MAC Address</th>
							<th>Session Left</th>
						</tr>
					</thead>
					<tbody>
						{#each { length: 4 } as _}
							<tr>
								<td><div class="skeleton" style="width:80px;height:13px"></div></td>
								<td><div class="skeleton" style="width:140px;height:13px"></div></td>
								<td><div class="skeleton" style="width:70px;height:13px"></div></td>
							</tr>
						{/each}
					</tbody>
				</table>
			{:else}
				<table class="users-table">
					<thead>
						<tr>
							<th>User</th>
							<th>MAC Address</th>
							<th>Session Left</th>
						</tr>
					</thead>
					<tbody>
						{#if users.length === 0}
							<tr>
								<td colspan="3" class="empty">No active users.</td>
							</tr>
						{:else}
							{#each users as u (u.node_id)}
								<tr>
									<td class="user-cell">{u.user}</td>
									<td class="mono">{u.mac_address}</td>
									<td>{u.session_left}</td>
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			{/if}
		</div>
	</div>
</div>

<style>
	.card {
		background: #fff;
		border: 1px solid #e8e8e8;
		border-radius: 10px;
		overflow: hidden;
	}

	.card-header {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 14px 18px;
		border-bottom: 1px solid #f0f0f0;
	}

	.header-left {
		display: flex;
		align-items: center;
		gap: 8px;
		flex: 1;
	}

	.header-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 30px;
		height: 30px;
		background: #fff4f1;
		border-radius: 7px;
		color: var(--color-theme-1);
		flex-shrink: 0;
	}

	.card-title {
		font-size: 0.88rem;
		font-weight: 700;
		color: #1a1a1a;
	}

	.refresh-btn {
		background: none;
		border: 1px solid #e8e8e8;
		border-radius: 6px;
		padding: 5px 8px;
		cursor: pointer;
		color: #888;
		display: flex;
		align-items: center;
		transition: background 0.15s, color 0.15s;
	}

	.refresh-btn:hover:not(:disabled) {
		background: #f5f5f5;
		color: #333;
	}

	.refresh-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.users-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.82rem;
	}

	.users-table th {
		text-align: left;
		padding: 10px 18px;
		color: #aaa;
		font-weight: 600;
		font-size: 0.72rem;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		border-bottom: 1px solid #f0f0f0;
		background: #fafafa;
	}

	.users-table td {
		padding: 11px 18px;
		border-bottom: 1px solid #f6f6f6;
		color: #1a1a1a;
	}

	.users-table tbody tr:last-child td {
		border-bottom: none;
	}

	.user-cell {
		font-weight: 600;
	}

	.mono {
		font-family: ui-monospace, monospace;
		color: #555;
	}

	.empty {
		padding: 20px 18px;
		font-size: 0.82rem;
		color: #bbb;
		font-style: italic;
		text-align: center;
	}

	.skeleton {
		background: linear-gradient(90deg, #f0f0f0 25%, #e8e8e8 50%, #f0f0f0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.4s infinite;
		border-radius: 4px;
	}

	@keyframes shimmer {
		0% { background-position: 200% 0; }
		100% { background-position: -200% 0; }
	}
</style>
