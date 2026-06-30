<script lang="ts">
	import type { iVendoVoucher } from '$lib/types/models';

	interface Props {
		vendoId: number;
		onsuccess?: (created: iVendoVoucher[]) => void;
		oncancel?: () => void;
	}

	const { vendoId, onsuccess, oncancel }: Props = $props();

	let prefix = $state('VC');
	let amount = $state(10);
	let quantity = $state(1);
	let addToSales = $state(false);
	let printThermal = $state(false);
	let isProcessing = $state(false);
	let error = $state('');

	const prefixPattern = /^[A-Za-z][A-Za-z0-9]?$/;

	async function submit() {
		error = '';
		if (!prefixPattern.test(prefix.trim())) {
			error = 'Prefix must be 1-2 characters and start with a letter.';
			return;
		}
		if (!amount || amount <= 0) {
			error = 'Amount must be greater than 0.';
			return;
		}
		if (quantity < 1 || quantity > 15) {
			error = 'Quantity must be between 1 and 15.';
			return;
		}

		isProcessing = true;
		try {
			const res = await fetch(`/x-api/vendo-machines/${vendoId}/vouchers/generate`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
				body: JSON.stringify({
					prefix: prefix.trim(),
					amount,
					quantity,
					add_to_sales: addToSales,
					print_thermal: printThermal
				})
			});
			const body = await res.json().catch(() => ({}));
			if (!res.ok) {
				error = body.detail ?? 'Failed to generate vouchers.';
				return;
			}
			onsuccess?.((body.data ?? []) as iVendoVoucher[]);
		} catch {
			error = 'An unexpected error occurred.';
		} finally {
			isProcessing = false;
		}
	}
</script>

<form
	onsubmit={(e) => {
		e.preventDefault();
		submit();
	}}
>
	{#if error}
		<div class="uk-alert uk-alert-danger">{error}</div>
	{/if}

	<div class="uk-margin-small-bottom">
		<label class="uk-form-label" for="voucher-prefix">Voucher Prefix</label>
		<input
			id="voucher-prefix"
			class="uk-input"
			type="text"
			maxlength="2"
			bind:value={prefix}
			placeholder="VC"
			required
		/>
	</div>

	<div class="uk-margin-small-bottom">
		<label class="uk-form-label" for="voucher-amount">Amount (Price)</label>
		<input
			id="voucher-amount"
			class="uk-input"
			type="number"
			min="0"
			step="any"
			bind:value={amount}
			required
		/>
	</div>

	<div class="uk-margin-small-bottom">
		<label class="uk-form-label" for="voucher-qty">Quantity (PC)</label>
		<input
			id="voucher-qty"
			class="uk-input"
			type="number"
			min="1"
			max="15"
			bind:value={quantity}
			required
		/>
		<div class="helper-text">Avoid more than 15; generation can take longer on the device.</div>
	</div>

	<div class="uk-margin-small-bottom">
		<label class="uk-form-label" for="voucher-sales">Add to Sales</label>
		<select id="voucher-sales" class="uk-select" bind:value={addToSales}>
			<option value={false}>No</option>
			<option value={true}>Yes</option>
		</select>
	</div>

	<div class="uk-margin-small-bottom">
		<label class="uk-form-label" for="voucher-print">Print on Thermal Printer</label>
		<select id="voucher-print" class="uk-select" bind:value={printThermal}>
			<option value={false}>No</option>
			<option value={true}>Yes</option>
		</select>
	</div>

	<div class="uk-text-center uk-margin-top">
		<button type="button" class="uk-modal-close uk-button" onclick={oncancel}>Cancel</button>
		<button
			type="submit"
			class="uk-button uk-button-primary uk-margin-left"
			disabled={isProcessing}
		>
			{isProcessing ? 'Generating…' : 'Generate'}
		</button>
	</div>
</form>

<style>
	.helper-text {
		font-size: 0.78rem;
		color: #777;
		margin-top: 4px;
	}
</style>
