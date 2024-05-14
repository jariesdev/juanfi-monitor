<script lang="ts">
	import { apiUrl, currentUser } from '$lib/store';
	import type { iUser } from '$lib/interfaces';
	import { createEventDispatcher } from 'svelte';

	let baseApiUrl: string = '';
	let isProcessing: boolean = false;
	let username: string = '';
	let password: string = '';
	let user: iUser;
	const dispatcher = createEventDispatcher();

	function submit(): void {
		isProcessing = true;
		const formData = new FormData();
		formData.append('username', username);
		formData.append('password', password);
		const request = new Request(`${baseApiUrl}/token`, { method: 'POST', body: formData });
		fetch(request)
			.then(async (response) => {
				if (!response.ok) {
					throw new Error(response.statusText);
				}

				const data = await response.json();
				localStorage.setItem('auth_token', data.access_token);
				currentUser.set(data.user);
				console.log(data);
				dispatcher('success', { user });
			})
			.catch(() => {
				localStorage.removeItem('auth_token');
				currentUser.set(null);
			})
			.finally(() => {
				isProcessing = false;
			});
	}

	apiUrl.subscribe(function (value) {
		baseApiUrl = value;
	});

	currentUser.subscribe(function (value) {
		user = value;
	});
</script>

<form on:submit|preventDefault={submit}>
	<div class="uk-grid uk-grid-small uk-child-width-1-1" style="row-gap: 20px">
		<div>
			<input
				bind:value={username}
				type="text"
				class="uk-input"
				name="username"
				placeholder="Username"
				disabled={isProcessing}
			/>
		</div>
		<div>
			<input
				bind:value={password}
				type="password"
				class="uk-input"
				name="password"
				placeholder="Password"
				disabled={isProcessing}
			/>
		</div>

		<div class="uk-text-center">
			<button type="submit" class="uk-button uk-button-primary" disabled={isProcessing}>
				Login
			</button>
		</div>
	</div>
</form>
