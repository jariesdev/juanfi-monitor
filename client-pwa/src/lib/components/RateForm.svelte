<script lang="ts">
	import type { iVendoRate } from '$lib/types/models';

	interface Props {
		rate?: iVendoRate | null;
		vendoId?: number | null;
		onsuccess?: () => void;
		oncancel?: () => void;
	}

	const { rate = null, vendoId = null, onsuccess, oncancel }: Props = $props();

	let isProcessing = $state(false);
	let error = $state('');

	let name = $state(rate?.name ?? '');
	let price = $state(rate?.price ?? 0);
	let minutes = $state(rate?.minutes ?? 0);
	let validityMinutes = $state(rate?.validity_minutes ?? 0);
	let dataLimitMb = $state<number | null>(rate?.data_limit_mb ?? null);
	let userProfile = $state(rate?.user_profile ?? '');

	async function submit() {
		error = '';
		if (!name.trim()) {
			error = 'Rate name is required.';
			return;
		}
		if (!price || !minutes || !validityMinutes) {
			error = 'Price, minutes, and validity are required.';
			return;
		}

		isProcessing = true;
		try {
			const url = rate
				? `/x-api/vendo-rates/${rate.id}`
				: vendoId
					? `/x-api/vendo-machines/${vendoId}/rates`
					: '/x-api/vendo-rates/default';
			const method = rate ? 'PUT' : 'POST';
			const res = await fetch(url, {
				method,
				headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
				body: JSON.stringify({
					name: name.trim(),
					price,
					minutes,
					validity_minutes: validityMinutes,
					data_limit_mb: dataLimitMb || null,
					user_profile: userProfile.trim() || null
				})
			});
			if (!res.ok) {
				const data = await res.json().catch(() => ({}));
				error = data.detail ?? 'Failed to save rate.';
				return;
			}
			onsuccess?.();
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
		<label class="uk-form-label" for="rate-name">Rate Name</label>
		<input
			bind:value={name}
			id="rate-name"
			class="uk-input"
			type="text"
			placeholder="e.g. Basic: 1 PHP / 20 Mins"
			required
		/>
	</div>

	<div class="uk-margin-small-bottom">
		<label class="uk-form-label" for="rate-price">Price</label>
		<input
			bind:value={price}
			id="rate-price"
			class="uk-input"
			type="number"
			min="0"
			step="any"
			required
		/>
	</div>

	<div class="uk-margin-small-bottom">
		<label class="uk-form-label" for="rate-minutes">Minutes</label>
		<input bind:value={minutes} id="rate-minutes" class="uk-input" type="number" min="0" required />
	</div>

	<div class="uk-margin-small-bottom">
		<label class="uk-form-label" for="rate-validity">Validity in minutes</label>
		<input
			bind:value={validityMinutes}
			id="rate-validity"
			class="uk-input"
			type="number"
			min="0"
			required
		/>
	</div>

	<div class="uk-margin-small-bottom">
		<label class="uk-form-label" for="rate-data-limit">Data (MB) Limit Usage (Optional)</label>
		<input
			bind:value={dataLimitMb}
			id="rate-data-limit"
			class="uk-input"
			type="number"
			min="0"
			placeholder="Data Limit Usage"
		/>
	</div>

	<div class="uk-margin-small-bottom">
		<label class="uk-form-label" for="rate-profile"
			>User Profile (Optional, it overrides the default profile set in system config)</label
		>
		<input
			bind:value={userProfile}
			id="rate-profile"
			class="uk-input"
			type="text"
			placeholder="e.g. basic"
		/>
	</div>

	<div class="uk-text-center uk-margin-top">
		<button type="button" class="uk-modal-close uk-button" onclick={oncancel}>Cancel</button>
		<button
			type="submit"
			class="uk-button uk-button-primary uk-margin-left"
			disabled={isProcessing}
		>
			{isProcessing ? 'Saving…' : rate ? 'Update Rate' : 'Add Rate'}
		</button>
	</div>
</form>
