<script lang="ts">
	import type { iRole, iUser, iVendo } from '$lib/types/models';

	interface Props {
		user?: iUser | null;
		roles?: iRole[];
		vendos?: iVendo[];
		onsuccess?: () => void;
		oncancel?: () => void;
	}

	const { user = null, roles = [], vendos = [], onsuccess, oncancel }: Props = $props();

	let isProcessing = $state(false);
	let error = $state('');

	let form = $state({
		username: user?.username ?? '',
		password: '',
		role_ids: user?.roles?.map((r) => r.id) ?? [] as number[],
		is_active: user?.is_active ?? true,
		vendo_ids: user?.vendos?.map((v) => v.id) ?? [] as number[]
	});

	const allVendosSelected = $derived(
		vendos.length > 0 && vendos.every((v) => form.vendo_ids.includes(v.id))
	);

	function toggleVendo(id: number) {
		if (form.vendo_ids.includes(id)) {
			form.vendo_ids = form.vendo_ids.filter((v) => v !== id);
		} else {
			form.vendo_ids = [...form.vendo_ids, id];
		}
	}

	function toggleAllVendos() {
		if (allVendosSelected) {
			form.vendo_ids = [];
		} else {
			form.vendo_ids = vendos.map((v) => v.id);
		}
	}

	async function submit() {
		error = '';
		isProcessing = true;

		const body: Record<string, unknown> = {
			username: form.username,
			role_ids: form.role_ids,
			is_active: form.is_active,
			vendo_ids: form.vendo_ids
		};
		if (form.password) {
			body.password = form.password;
		} else if (!user) {
			error = 'Password is required for new users.';
			isProcessing = false;
			return;
		}

		try {
			const url = user ? `/x-api/users/${user.id}` : '/x-api/users';
			const method = user ? 'PUT' : 'POST';
			const res = await fetch(url, {
				method,
				headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
				body: JSON.stringify(body)
			});
			if (!res.ok) {
				const data = await res.json().catch(() => ({}));
				error = data.detail ?? 'Failed to save user.';
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

<form onsubmit={(e) => { e.preventDefault(); submit(); }}>
	{#if error}
		<div class="uk-alert uk-alert-danger">{error}</div>
	{/if}

	<div class="uk-margin-small-bottom">
		<label class="uk-form-label" for="uf-username">Username</label>
		<input
			bind:value={form.username}
			id="uf-username"
			class="uk-input"
			type="text"
			placeholder="e.g. jdoe"
			required
		/>
	</div>

	<div class="uk-margin-small-bottom">
		<label class="uk-form-label" for="uf-password">
			Password {#if user}<span class="uk-text-muted uk-text-small">(leave blank to keep current)</span>{/if}
		</label>
		<input
			bind:value={form.password}
			id="uf-password"
			class="uk-input"
			type="password"
			placeholder={user ? 'Leave blank to keep current' : 'Required'}
		/>
	</div>

	<div class="uk-margin-small-bottom">
		<label class="uk-form-label">Roles</label>
		<div class="vendo-list">
			{#each roles as role}
				<label class="vendo-item">
					<input
						type="checkbox"
						class="uk-checkbox"
						checked={form.role_ids.includes(role.id)}
						onchange={() => {
							if (form.role_ids.includes(role.id)) {
								form.role_ids = form.role_ids.filter((id) => id !== role.id);
							} else {
								form.role_ids = [...form.role_ids, role.id];
							}
						}}
					/>
					<span>{role.name}</span>
				</label>
			{:else}
				<span class="uk-text-muted uk-text-small">No roles available.</span>
			{/each}
		</div>
	</div>

	<div class="uk-margin-small-bottom">
		<label class="uk-form-label">
			<input
				type="checkbox"
				class="uk-checkbox"
				bind:checked={form.is_active}
				style="margin-right: 6px;"
			/>
			Active
		</label>
	</div>

	{#if vendos.length > 0}
		<div class="uk-margin-small-bottom">
			<div class="vendo-header">
				<label class="uk-form-label" style="margin: 0;">Assign Vendos</label>
				<button type="button" class="select-all-btn" onclick={toggleAllVendos}>
					{allVendosSelected ? 'Deselect All' : 'Select All'}
				</button>
			</div>
			<div class="vendo-list">
				{#each vendos as vendo}
					<label class="vendo-item">
						<input
							type="checkbox"
							class="uk-checkbox"
							checked={form.vendo_ids.includes(vendo.id)}
							onchange={() => toggleVendo(vendo.id)}
						/>
						<span>{vendo.name}</span>
					</label>
				{/each}
			</div>
		</div>
	{/if}

	<div class="uk-text-center uk-margin-top">
		<button type="button" class="uk-modal-close uk-button" onclick={oncancel}>Cancel</button>
		<button
			type="submit"
			class="uk-button uk-button-primary uk-margin-left"
			disabled={isProcessing}
		>
			{isProcessing ? 'Saving…' : user ? 'Update User' : 'Create User'}
		</button>
	</div>
</form>

<style>
	.vendo-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 6px;
	}

	.select-all-btn {
		background: none;
		border: none;
		padding: 0;
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--color-theme-1);
		cursor: pointer;
		text-decoration: underline;
		text-underline-offset: 2px;
	}

	.select-all-btn:hover {
		opacity: 0.75;
	}

	.vendo-list {
		display: flex;
		flex-direction: column;
		gap: 6px;
		max-height: 160px;
		overflow-y: auto;
		border: 1px solid #e8e8e8;
		border-radius: 4px;
		padding: 8px;
	}

	.vendo-item {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 0.875rem;
		cursor: pointer;
	}
</style>
