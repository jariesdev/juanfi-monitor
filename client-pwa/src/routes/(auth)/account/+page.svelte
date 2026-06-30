<svelte:head>
	<title>Account</title>
</svelte:head>

<script lang="ts">
	import { enhance } from '$app/forms';
	import { page } from '$app/state';
	import type { SubmitFunction } from '@sveltejs/kit';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	let isProcessing = $state(false);
	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');

	const mismatch = $derived(confirmPassword.length > 0 && newPassword !== confirmPassword);

	const handleSubmit: SubmitFunction = () => {
		isProcessing = true;
		return async ({ update }) => {
			await update();
			isProcessing = false;
			if (page.form?.success) {
				currentPassword = '';
				newPassword = '';
				confirmPassword = '';
			}
		};
	};
</script>

<div class="page">
	<div class="card">
		<!-- Card header -->
		<div class="card-header">
			<span class="header-icon" uk-icon="icon: lock; ratio: 1.1"></span>
			<div>
				<div class="card-title">Change Password</div>
				<div class="card-subtitle">Signed in as <strong>{data.user?.username}</strong></div>
			</div>
		</div>

		<!-- Feedback banners -->
		{#if page.form?.success}
			<div class="banner banner-success">
				<span uk-icon="icon: check; ratio: 0.9"></span>
				Password updated successfully.
			</div>
		{/if}
		{#if page.form?.error}
			<div class="banner banner-error">
				<span uk-icon="icon: warning; ratio: 0.9"></span>
				{page.form.error}
			</div>
		{/if}

		<form method="POST" use:enhance={handleSubmit} novalidate>
			<div class="fields">
				<div class="field">
					<label for="current_password">Current password</label>
					<input
						id="current_password"
						bind:value={currentPassword}
						type="password"
						name="current_password"
						class="uk-input"
						placeholder="••••••••"
						disabled={isProcessing}
						autocomplete="current-password"
					/>
				</div>

				<div class="divider"></div>

				<div class="field">
					<label for="new_password">New password</label>
					<input
						id="new_password"
						bind:value={newPassword}
						type="password"
						name="new_password"
						class="uk-input"
						placeholder="••••••••"
						disabled={isProcessing}
						autocomplete="new-password"
					/>
					{#if newPassword.length > 0 && newPassword.length < 6}
						<span class="hint hint-warn">At least 6 characters required</span>
					{/if}
				</div>

				<div class="field">
					<label for="confirm_password">Confirm new password</label>
					<input
						id="confirm_password"
						bind:value={confirmPassword}
						type="password"
						name="confirm_password"
						class="uk-input"
						class:uk-form-danger={mismatch}
						placeholder="••••••••"
						disabled={isProcessing}
						autocomplete="new-password"
					/>
					{#if mismatch}
						<span class="hint hint-error">Passwords don't match</span>
					{/if}
				</div>
			</div>

			<button
				type="submit"
				class="submit-btn"
				disabled={isProcessing || mismatch}
			>
				{#if isProcessing}
					<span uk-spinner="ratio: 0.6"></span>
					Saving…
				{:else}
					Update Password
				{/if}
			</button>
		</form>
	</div>
</div>

<style>
	.page {
		padding-top: 8px;
	}

	.card {
		background: #fff;
		border: 1px solid #e8e8e8;
		border-radius: 10px;
		max-width: 420px;
		overflow: hidden;
	}

	/* Header */
	.card-header {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 18px 20px 16px;
		border-bottom: 1px solid #f0f0f0;
	}

	.header-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 36px;
		height: 36px;
		background: #fff4f1;
		border-radius: 8px;
		color: var(--color-theme-1);
		flex-shrink: 0;
	}

	.card-title {
		font-size: 0.9rem;
		font-weight: 700;
		color: #1a1a1a;
		line-height: 1.2;
	}

	.card-subtitle {
		font-size: 0.75rem;
		color: #aaa;
		margin-top: 2px;
	}

	/* Feedback banners */
	.banner {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 20px;
		font-size: 0.82rem;
		font-weight: 500;
	}

	.banner-success {
		background: #f0faf4;
		color: #1a7f4b;
		border-bottom: 1px solid #d0edd9;
	}

	.banner-error {
		background: #fff4f1;
		color: #c0392b;
		border-bottom: 1px solid #fad5cc;
	}

	/* Form */
	form {
		padding: 20px;
	}

	.fields {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}

	.field {
		display: flex;
		flex-direction: column;
		gap: 5px;
	}

	.field label {
		font-size: 0.78rem;
		font-weight: 600;
		color: #666;
		letter-spacing: 0.01em;
	}

	.divider {
		height: 1px;
		background: #f2f2f2;
		margin: 2px 0;
	}

	.hint {
		font-size: 0.72rem;
		margin-top: 2px;
	}

	.hint-warn {
		color: #e67e22;
	}

	.hint-error {
		color: #c0392b;
	}

	/* Submit button */
	.submit-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		width: 100%;
		margin-top: 20px;
		padding: 10px;
		background: var(--color-theme-1);
		color: #fff;
		border: none;
		border-radius: 7px;
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
		transition: opacity 0.15s, background 0.15s;
	}

	.submit-btn:hover:not(:disabled) {
		background: #e03500;
	}

	.submit-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
</style>
