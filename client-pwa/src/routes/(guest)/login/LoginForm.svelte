<script lang="ts">
	import {enhance} from '$app/forms';
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

    const handleSubmit = () => {
        isProcessing = true; // Set to true when submission starts
        return async ({ update }) => {
            await update(); // Wait for the form update (data, status, etc.)
            isProcessing = false; // Set to false when submission finishes
        };
    }
</script>

<form method="POST" action="/login" use:enhance={handleSubmit}>
	<div class="uk-grid uk-grid-small uk-child-width-1-1" style="row-gap: 20px">
		<div>
			<div class="uk-inline uk-display-block">
				<span class="uk-form-icon" uk-icon="icon: user"></span>
				<input
					bind:value={username}
					type="text"
					class="uk-input"
					name="username"
					placeholder="Username"
					disabled={isProcessing}
					class:uk-form-danger={page.form?.invalid}
				/>

				<!-- Display the error message if the login failed -->
				{#if page.form?.invalid}
					<p class="error-message uk-text-small uk-hidden">{page.form.message}</p>
				{/if}
			</div>
		</div>

		<div>
			<div class="uk-inline uk-display-block">
				<span class="uk-form-icon" uk-icon="icon: lock"></span>
				<input
					bind:value={password}
					type="password"
					class="uk-input"
					name="password"
					placeholder="Password"
					disabled={isProcessing}
					class:uk-form-danger={page.form?.invalid}
				/>
			</div>
		</div>

		<div class="uk-text-center">
			<button type="submit" class="uk-button uk-button-primary" disabled={isProcessing}>
				Login
			</button>
		</div>
	</div>
</form>
