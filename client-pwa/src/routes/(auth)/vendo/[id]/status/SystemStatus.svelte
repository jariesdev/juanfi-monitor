<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import moment from 'moment';

	import {changeVendoStatus, getVendoInfo} from "$lib/remote/vendo.remote";
	import type {iVendo} from "$lib/types/models";

	interface iStatus {
		key: string;
		label: string;
		text: string;
	}

	// export let vendoId: number;
	const {vendoId} = $props()
	let statuses: iStatus[] = $state([]);
	let isLoading: boolean = $state(true);
	let systemUptime: number = $state(0);
	let serverTime: number = $state(0);
	let controller: AbortController | undefined = undefined;
	let intervalId: any;
	let timeIntervalId: any;
	let isWithdrawing: boolean = $state(false);

	let vendo: iVendo|null = $derived(await getVendoInfo(+vendoId))

	function loadStatuses(): void {
		isLoading = true;
		controller = new AbortController();
		const signal = controller.signal;
		const request = new Request(`/x-api/vendo-machines/${vendoId}/status?nosw=1`, {
			method: 'GET',
			signal: signal
		});
		fetch(request)
			.then((response) => {
				if (response.status === 200) {
					return response.json();
				} else {
					throw new Error('Something went wrong on API server!');
				}
			})
			.then((response) => {
				systemUptime = response.system_uptime_ms;
				serverTime = response.server_time;
				const list: iStatus[] = [];
				Object.keys(response).forEach((key): void => {
					if (!['system_uptime_ms', 'server_time'].includes(key)) {
						list.push({
							key: key,
							label: titleCase(key),
							text: response[key]
						});
					}
				});
				statuses = list;
				startTimer();
			})
			.catch((error) => {
				console.error(error);
			})
			.finally(() => {
				isLoading = false;
			});
	}

	function titleCase(s: string) {
		return s.replace(/^_*(.)|_+(.)/g, (s: string, c: string, d: string) =>
			c ? c.toUpperCase() : ' ' + d.toUpperCase()
		);
	}

	function startTimer(): void {
		if (timeIntervalId) return;
		timeIntervalId = setInterval(() => {
			systemUptime += 1000;
		}, 1000);
	}

	function withdrawCurrenSale(): void {
		const confirmed = confirm('This will reset the current sales counter to 0. Proceed?');

		if (!confirmed) {
			return;
		}

		isWithdrawing = true;
		let url = `/x-api/vendo-machines/${vendoId}/withdraw-current-sales`;
		controller = new AbortController();
		const signal = controller.signal;
		const request: Request = new Request(url, { method: 'POST', signal });
		fetch(request)
			.then((response) => {
				if (response.ok) {
					return response.json();
				}
				throw new Error(response.statusText);
			})
			.finally(() => {
				isWithdrawing = false;
			});
	}

	function toRelativeTime(time: number): string {
		const d = moment().diff(Date.now() - time, 'days');
		const h = moment().diff(Date.now() - time, 'hours') % 24;
		const m = moment().diff(Date.now() - time, 'minutes') % 60;
		const s = moment().diff(Date.now() - time, 'seconds') % 60;
		const mf = String(m).padStart(2, '0');
		const sf = String(s).padStart(2, '0');

		if (d > 0) {
			return `${d}d ${h}:${mf}:${sf}`;
		} else if (h > 0) {
			return `${h}:${mf}:${sf}`;
		} else if (m > 0) {
			return `${mf}:${sf}`;
		}
		return `${sf}`;
	}

	// $: serverTimeString = (): string => {
	// 	return moment(serverTime).format();
	// };
	let serverTimeString = $derived(():string => moment(serverTime).format())

	onMount(() => {
		loadStatuses();
		// refresh status
		intervalId = setInterval(() => loadStatuses(), 30 * 1000);
	});
	onDestroy(() => {
		controller && controller.abort('component destroyed');
		if (intervalId) {
			clearInterval(intervalId);
		}
		if (timeIntervalId) {
			clearInterval(timeIntervalId);
		}
	});
</script>

<div class="card">
	<div class="card-header">
		<span class="header-icon" uk-icon="icon: settings; ratio: 1"></span>
		<span class="card-title">System Status</span>
	</div>

	{#if isLoading}
		<div class="status-list">
			{#each { length: 6 } as _}
				<div class="status-row">
					<div class="skeleton skeleton-label"></div>
					<div class="skeleton skeleton-value"></div>
				</div>
			{/each}
		</div>
	{:else if statuses.length > 0}
		<div class="status-list">
			<div class="status-row">
				<span class="status-key">System Uptime</span>
				<span class="status-val">{toRelativeTime(systemUptime)}</span>
			</div>
			<div class="status-row">
				<span class="status-key">Server Time</span>
				<span class="status-val">{serverTimeString()}</span>
			</div>
			{#each statuses as status}
				<div class="status-row">
					<span class="status-key">{status.label}</span>
					<div class="status-val-wrap">
						{#if status.key === 'current_coin_count'}
							<button
								disabled={isWithdrawing || parseFloat(status.text || '0') === 0}
								class="withdraw-btn"
								type="button"
								onclick={withdrawCurrenSale}
							>
								<span uk-icon="icon: credit-card; ratio: 0.8"></span>
								Withdraw
							</button>
							<span class="status-val">{status.text}</span>
						{:else}
							<span class="status-val">{status.text}</span>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{:else}
		<p class="status-empty">Status not available yet.</p>
	{/if}
</div>

{#if vendo}
	<div class="card toggle-card">
		<button
			class="toggle-btn"
			class:toggle-btn-danger={vendo.is_active}
			type="button"
			onclick={async () => { changeVendoStatus({id: vendo.id, status: !vendo.is_active}).then(() => getVendoInfo(+vendoId).refresh())}}
		>
			<span uk-icon="icon: {vendo.is_active ? 'ban' : 'check'}; ratio: 0.85"></span>
			{vendo.is_active ? 'Disable' : 'Enable'} {vendo.name}
		</button>
	</div>
{/if}

<style>
	.card {
		background: #fff;
		border: 1px solid #e8e8e8;
		border-radius: 10px;
		overflow: hidden;
		margin-bottom: 12px;
	}

	.card-header {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 14px 18px;
		border-bottom: 1px solid #f0f0f0;
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

	/* Status rows */
	.status-list {
		padding: 4px 0;
	}

	.status-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 10px 18px;
		border-bottom: 1px solid #f6f6f6;
		gap: 12px;
	}

	.status-row:last-child {
		border-bottom: none;
	}

	.status-key {
		font-size: 0.82rem;
		color: #666;
		font-weight: 500;
	}

	.status-val {
		font-size: 0.82rem;
		color: #1a1a1a;
		font-weight: 600;
		text-align: right;
	}

	.status-val-wrap {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.status-empty {
		padding: 20px 18px;
		font-size: 0.82rem;
		color: #bbb;
		font-style: italic;
		margin: 0;
	}

	/* Withdraw button */
	.withdraw-btn {
		display: inline-flex;
		align-items: center;
		gap: 5px;
		padding: 5px 10px;
		background: var(--color-theme-1);
		color: #fff;
		border: none;
		border-radius: 6px;
		font-size: 0.78rem;
		font-weight: 600;
		font-family: inherit;
		cursor: pointer;
		transition: opacity 0.15s;
	}

	.withdraw-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	/* Toggle card */
	.toggle-card {
		padding: 14px 18px;
	}

	.toggle-btn {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 7px 14px;
		background: #22a06b;
		color: #fff;
		border: none;
		border-radius: 7px;
		font-size: 0.82rem;
		font-weight: 600;
		font-family: inherit;
		cursor: pointer;
		transition: opacity 0.15s;
	}

	.toggle-btn-danger {
		background: #c0392b;
	}

	.toggle-btn:hover {
		opacity: 0.88;
	}

	/* Skeletons */
	.skeleton {
		background: linear-gradient(90deg, #f0f0f0 25%, #e8e8e8 50%, #f0f0f0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.4s infinite;
		border-radius: 4px;
	}

	.skeleton-label { width: 130px; height: 13px; }
	.skeleton-value { width: 70px; height: 13px; }

	@keyframes shimmer {
		0% { background-position: 200% 0; }
		100% { background-position: -200% 0; }
	}
</style>