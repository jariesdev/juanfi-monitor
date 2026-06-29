<svelte:head>
	<title>Withdraw Sales</title>
</svelte:head>

<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { toast } from '$lib/store';

	const { data } = $props();
	const vendoId = $derived(data.id);

	let currentCoinCount: number = $state(0);
	let isLoading: boolean = $state(true);
	let isWithdrawing: boolean = $state(false);
	let showModal: boolean = $state(false);
	let controller: AbortController | undefined;
	let pollInterval: ReturnType<typeof setInterval> | undefined;

	function loadStatus() {
		controller?.abort();
		controller = new AbortController();

		fetch(`/x-api/vendo-machines/${vendoId}/status?nosw=1`, { signal: controller.signal })
			.then((r) => (r.ok ? r.json() : Promise.reject(r.statusText)))
			.then((data) => {
				currentCoinCount = data.current_coin_count ?? 0;
			})
			.catch(() => {})
			.finally(() => {
				isLoading = false;
			});
	}

	function confirmWithdraw() {
		if (currentCoinCount === 0) return;
		showModal = true;
	}

	function cancelWithdraw() {
		showModal = false;
	}

	function executeWithdraw() {
		isWithdrawing = true;
		showModal = false;

		fetch(`/x-api/vendo-machines/${vendoId}/withdraw-current-sales`, { method: 'POST' })
			.then((r) => {
				if (r.ok) return r.json();
				throw new Error(r.statusText);
			})
			.then(() => {
				toast.set({ message: 'Sales withdrawn successfully.', type: 'success' });
				goto('/withdrawals');
			})
			.catch(() => {
				toast.set({ message: 'Withdrawal failed. Please try again.', type: 'error' });
				isWithdrawing = false;
			});
	}

	onMount(() => {
		loadStatus();
		pollInterval = setInterval(loadStatus, 5000);
	});

	onDestroy(() => {
		controller?.abort();
		clearInterval(pollInterval);
	});
</script>

<div class="uk-section">
	<div class="uk-container">
		<div class="page-header">
			<a href="/vendo/{vendoId}/status" class="back-btn">
				<span uk-icon="icon: arrow-left; ratio: 0.9"></span>
				Back
			</a>
			<h1 class="page-title">Withdraw Sales</h1>
		</div>

		<!-- Live current sales card -->
		<div class="card sales-card">
			<div class="sales-label">Current Sales</div>
			{#if isLoading}
				<div class="sales-skeleton"></div>
			{:else}
				<div class="sales-amount">₱{currentCoinCount.toLocaleString('en-PH', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</div>
				<div class="sales-unit">current sales</div>
			{/if}
			<div class="live-badge">
				<span class="live-dot"></span>
				Live
			</div>
		</div>

		<!-- Warning card -->
		<div class="card warning-card">
			<div class="warning-header">
				<span uk-icon="icon: warning; ratio: 1"></span>
				<span class="warning-title">Before you proceed</span>
			</div>
			<p class="warning-text">
				Withdrawing will permanently reset the current sales counter to zero. Ensure you have
				physically collected the coins or recorded the amount before continuing — this
				cannot be undone.
			</p>
		</div>

		<!-- Action -->
		<div class="card action-card">
			<button
				class="withdraw-btn"
				type="button"
				disabled={isLoading || isWithdrawing || currentCoinCount === 0}
				onclick={confirmWithdraw}
			>
				{#if isWithdrawing}
					<span uk-icon="icon: refresh; ratio: 0.9"></span>
					Processing…
				{:else}
					<span uk-icon="icon: credit-card; ratio: 0.9"></span>
					Withdraw {currentCoinCount} coins
				{/if}
			</button>
			{#if currentCoinCount === 0 && !isLoading}
				<p class="no-sales-note">There are no coins to withdraw at this time.</p>
			{/if}
		</div>
	</div>
</div>

<!-- Confirmation modal -->
{#if showModal}
	<div
		class="modal-backdrop"
		role="presentation"
		onclick={cancelWithdraw}
		onkeydown={(e) => { if (e.key === 'Escape') cancelWithdraw(); }}
	>
		<div class="modal" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()} role="dialog" aria-modal="true" tabindex="-1">
			<div class="modal-header">
				<span uk-icon="icon: credit-card; ratio: 1"></span>
				<span class="modal-title">Confirm Withdrawal</span>
			</div>
			<div class="modal-body">
				<p>
					You are about to withdraw <strong>{currentCoinCount} coins</strong> from the current sales
					counter.
				</p>
				<p class="modal-note">This action cannot be undone.</p>
			</div>
			<div class="modal-footer">
				<button class="btn-cancel" type="button" onclick={cancelWithdraw}>Cancel</button>
				<button class="btn-confirm" type="button" onclick={executeWithdraw}>
					<span uk-icon="icon: check; ratio: 0.85"></span>
					Confirm Withdrawal
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.page-header {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 16px;
	}

	.back-btn {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		padding: 6px 12px;
		border: 1px solid #e8e8e8;
		border-radius: 7px;
		font-size: 0.82rem;
		font-weight: 500;
		color: #555;
		text-decoration: none;
		background: #fff;
		transition: background 0.15s;
	}

	.back-btn:hover {
		background: #f5f5f5;
		color: #333;
	}

	.page-title {
		font-size: 1rem;
		font-weight: 700;
		color: #1a1a1a;
		margin: 0;
	}

	/* Cards */
	.card {
		background: #fff;
		border: 1px solid #e8e8e8;
		border-radius: 10px;
		overflow: hidden;
		margin-bottom: 12px;
	}

	/* Sales card */
	.sales-card {
		padding: 28px 24px 20px;
		position: relative;
		text-align: center;
	}

	.sales-label {
		font-size: 0.78rem;
		font-weight: 600;
		color: #999;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		margin-bottom: 8px;
	}

	.sales-amount {
		font-size: 3rem;
		font-weight: 800;
		color: #1a1a1a;
		line-height: 1;
	}

	.sales-unit {
		font-size: 0.8rem;
		color: #999;
		margin-top: 4px;
	}

	.sales-skeleton {
		width: 100px;
		height: 48px;
		margin: 0 auto 4px;
		background: linear-gradient(90deg, #f0f0f0 25%, #e8e8e8 50%, #f0f0f0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.4s infinite;
		border-radius: 6px;
	}

	.live-badge {
		position: absolute;
		top: 14px;
		right: 14px;
		display: inline-flex;
		align-items: center;
		gap: 5px;
		font-size: 0.72rem;
		font-weight: 600;
		color: #1a7a4a;
		background: #e6f7ee;
		border-radius: 20px;
		padding: 3px 8px;
	}

	.live-dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: #1a7a4a;
		animation: pulse 1.5s infinite;
	}

	/* Warning card */
	.warning-card {
		padding: 16px 18px;
		background: #fffbf0;
		border-color: #f5d87a;
	}

	.warning-header {
		display: flex;
		align-items: center;
		gap: 8px;
		color: #b07d00;
		margin-bottom: 8px;
	}

	.warning-title {
		font-size: 0.85rem;
		font-weight: 700;
	}

	.warning-text {
		font-size: 0.82rem;
		color: #7a5c00;
		line-height: 1.6;
		margin: 0;
	}

	/* Action card */
	.action-card {
		padding: 16px 18px;
	}

	.withdraw-btn {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		width: 100%;
		justify-content: center;
		padding: 12px 20px;
		background: var(--color-theme-1);
		color: #fff;
		border: none;
		border-radius: 8px;
		font-size: 0.9rem;
		font-weight: 700;
		font-family: inherit;
		cursor: pointer;
		transition: opacity 0.15s;
	}

	.withdraw-btn:hover:not(:disabled) {
		opacity: 0.9;
	}

	.withdraw-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.no-sales-note {
		text-align: center;
		font-size: 0.78rem;
		color: #bbb;
		margin: 10px 0 0;
	}

	/* Modal */
	.modal-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.45);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
		padding: 16px;
	}

	.modal {
		background: #fff;
		border-radius: 12px;
		width: 100%;
		max-width: 380px;
		box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
		animation: modal-in 0.18s ease;
	}

	.modal-header {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 18px 20px 14px;
		border-bottom: 1px solid #f0f0f0;
		color: var(--color-theme-1);
	}

	.modal-title {
		font-size: 0.95rem;
		font-weight: 700;
		color: #1a1a1a;
	}

	.modal-body {
		padding: 16px 20px;
	}

	.modal-body p {
		font-size: 0.85rem;
		color: #333;
		margin: 0 0 8px;
		line-height: 1.5;
	}

	.modal-body p:last-child {
		margin-bottom: 0;
	}

	.modal-note {
		font-size: 0.78rem !important;
		color: #e74c3c !important;
		font-weight: 600;
	}

	.modal-footer {
		display: flex;
		gap: 10px;
		padding: 14px 20px 18px;
		justify-content: flex-end;
	}

	.btn-cancel {
		padding: 8px 16px;
		background: #f5f5f5;
		border: 1px solid #e8e8e8;
		border-radius: 7px;
		font-size: 0.82rem;
		font-weight: 600;
		font-family: inherit;
		color: #555;
		cursor: pointer;
		transition: background 0.15s;
	}

	.btn-cancel:hover {
		background: #ebebeb;
	}

	.btn-confirm {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 8px 16px;
		background: var(--color-theme-1);
		color: #fff;
		border: none;
		border-radius: 7px;
		font-size: 0.82rem;
		font-weight: 700;
		font-family: inherit;
		cursor: pointer;
		transition: opacity 0.15s;
	}

	.btn-confirm:hover {
		opacity: 0.9;
	}

	@keyframes pulse {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.4; }
	}

	@keyframes shimmer {
		0% { background-position: 200% 0; }
		100% { background-position: -200% 0; }
	}

	@keyframes modal-in {
		from { opacity: 0; transform: scale(0.96); }
		to   { opacity: 1; transform: scale(1); }
	}
</style>
