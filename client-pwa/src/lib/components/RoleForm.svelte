<script lang="ts">
	import type { iRole } from '$lib/types/models';
	import { ALL_PERMISSIONS } from '$lib/types/models';

	interface Props {
		role?: iRole | null;
		onsuccess?: () => void;
		oncancel?: () => void;
	}

	const { role = null, onsuccess, oncancel }: Props = $props();

	let isProcessing = $state(false);
	let error = $state('');

	let name = $state(role?.name ?? '');
	let selectedPerms = $state<string[]>(role?.permissions ?? []);

	function togglePerm(perm: string) {
		if (selectedPerms.includes(perm)) {
			selectedPerms = selectedPerms.filter((p) => p !== perm);
		} else {
			selectedPerms = [...selectedPerms, perm];
		}
	}

	const PERM_LABELS: Record<string, string> = {
		dashboard: 'Dashboard',
		account: 'Account',
		vendos: 'Vendo Machines',
		sales: 'Sales',
		logs: 'Logs',
		withdrawals: 'Withdrawals',
		rates: 'Vendo Rates',
		users: 'User Management (Admin)'
	};

	async function submit() {
		error = '';
		if (!name.trim()) {
			error = 'Role name is required.';
			return;
		}
		isProcessing = true;
		try {
			const url = role ? `/x-api/roles/${role.id}` : '/x-api/roles';
			const method = role ? 'PUT' : 'POST';
			const res = await fetch(url, {
				method,
				headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
				body: JSON.stringify({ name: name.trim(), permissions: selectedPerms })
			});
			if (!res.ok) {
				const data = await res.json().catch(() => ({}));
				error = data.detail ?? 'Failed to save role.';
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
		<label class="uk-form-label" for="rf-name">Role Name</label>
		<input
			bind:value={name}
			id="rf-name"
			class="uk-input"
			type="text"
			placeholder="e.g. Operator"
			required
		/>
	</div>

	<div class="uk-margin-small-bottom">
		<label class="uk-form-label">Permissions</label>
		<div class="perm-list">
			{#each ALL_PERMISSIONS as perm}
				<label class="perm-item">
					<input
						type="checkbox"
						class="uk-checkbox"
						checked={selectedPerms.includes(perm)}
						onchange={() => togglePerm(perm)}
					/>
					<span>{PERM_LABELS[perm] ?? perm}</span>
				</label>
			{/each}
		</div>
	</div>

	<div class="uk-text-center uk-margin-top">
		<button type="button" class="uk-modal-close uk-button" onclick={oncancel}>Cancel</button>
		<button
			type="submit"
			class="uk-button uk-button-primary uk-margin-left"
			disabled={isProcessing}
		>
			{isProcessing ? 'Saving…' : role ? 'Update Role' : 'Create Role'}
		</button>
	</div>
</form>

<style>
	.perm-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
		border: 1px solid #e8e8e8;
		border-radius: 4px;
		padding: 10px 12px;
	}

	.perm-item {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 0.875rem;
		cursor: pointer;
	}
</style>
