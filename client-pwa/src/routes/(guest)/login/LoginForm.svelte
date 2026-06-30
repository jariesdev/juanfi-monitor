<script lang="ts">
	import {enhance} from '$app/forms';
	import type { SubmitFunction } from '@sveltejs/kit';
	import {currentUser} from '$lib/store';
	import type {iUser} from '$lib/types/models';
	import {page} from '$app/state';

	interface Props {
		onSuccess: (user: object) => void,
		onFailed: () => void
	}

	let isProcessing: boolean = $state(false);
	let username: string = $state('');
	let password: string = $state('');
	let user: iUser;
	// let {onSuccess, onFailed}: Props = $props();

	currentUser.subscribe(function (value) {
		user = value;
	});

    const handleSubmit: SubmitFunction = () => {
        isProcessing = true;
        return async ({ update }) => {
            await update();
            isProcessing = false;
        };
    }
</script>

<form method="POST" action="/login" use:enhance={handleSubmit} novalidate>
	{#if page.form?.invalid}
		<div class="error-banner">
			<span uk-icon="icon: warning; ratio: 0.85"></span>
			{page.form.message ?? 'Invalid username or password.'}
		</div>
	{/if}

	<div class="fields">
		<div class="field">
			<label for="login-username">Username</label>
			<div class="uk-inline uk-display-block">
				<span class="uk-form-icon" uk-icon="icon: user"></span>
				<input
					id="login-username"
					bind:value={username}
					type="text"
					class="uk-input"
					name="username"
					placeholder="Enter username"
					disabled={isProcessing}
					autocomplete="username"
					class:uk-form-danger={page.form?.invalid}
				/>
			</div>
		</div>

		<div class="field">
			<label for="login-password">Password</label>
			<div class="uk-inline uk-display-block">
				<span class="uk-form-icon" uk-icon="icon: lock"></span>
				<input
					id="login-password"
					bind:value={password}
					type="password"
					class="uk-input"
					name="password"
					placeholder="Enter password"
					disabled={isProcessing}
					autocomplete="current-password"
					class:uk-form-danger={page.form?.invalid}
				/>
			</div>
		</div>
	</div>

	<button type="submit" class="submit-btn" disabled={isProcessing}>
		{isProcessing ? 'Signing in…' : 'Sign In'}
	</button>
</form>

<style>
	.error-banner {
		display: flex;
		align-items: center;
		gap: 7px;
		background: #fff4f1;
		color: #c0392b;
		font-size: 0.8rem;
		font-weight: 500;
		padding: 9px 12px;
		border-radius: 7px;
		margin-bottom: 16px;
	}

	.fields {
		display: flex;
		flex-direction: column;
		gap: 14px;
		margin-bottom: 20px;
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
	}

	.submit-btn {
		display: block;
		width: 100%;
		padding: 10px;
		background: var(--color-theme-1);
		color: #fff;
		border: none;
		border-radius: 7px;
		font-size: 0.875rem;
		font-weight: 600;
		font-family: inherit;
		cursor: pointer;
		transition: opacity 0.15s;
	}

	.submit-btn:hover:not(:disabled) {
		opacity: 0.9;
	}

	.submit-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
</style>
