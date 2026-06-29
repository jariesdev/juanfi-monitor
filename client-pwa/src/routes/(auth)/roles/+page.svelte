<script lang="ts">
	import SimpleTable from '$lib/components/SimpleTable.svelte';
	import RoleForm from '$lib/components/RoleForm.svelte';
	import type { iRole } from '$lib/types/models';
	import type { RowItem, SimpleTableHeader } from '$lib/types/datatable';

	let roles: iRole[] = $state([]);
	let isLoading = $state(true);
	let editingRole: iRole | null = $state(null);
	let showModal = $state(false);
	let pageError = $state('');

	const headers: SimpleTableHeader[] = [
		{ label: 'ID',          field: 'id',          sortable: true  },
		{ label: 'Name',        field: 'name',        sortable: true  },
		{ label: 'Permissions', field: 'permissions', sortable: false },
		{ label: 'Actions',     field: 'actions',     sortable: false }
	];

	async function loadRoles() {
		isLoading = true;
		try {
			const res = await fetch('/x-api/roles');
			if (res.ok) {
				const data = await res.json();
				roles = data.data ?? [];
			}
		} finally {
			isLoading = false;
		}
	}

	loadRoles();

	function openAdd() {
		editingRole = null;
		showModal = true;
	}

	function openEdit(role: iRole) {
		editingRole = role;
		showModal = true;
	}

	function onFormSuccess() {
		showModal = false;
		loadRoles();
	}

	async function deleteRole(role: iRole) {
		pageError = '';
		if (!confirm(`Delete role "${role.name}"?`)) return;
		const res = await fetch(`/x-api/roles/${role.id}`, { method: 'DELETE' });
		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			pageError = body.detail ?? 'Failed to delete role.';
			return;
		}
		loadRoles();
	}
</script>

<div class="uk-section uk-section-xsmall">
	<div class="page-header">
		<h2 class="uk-h3" style="margin: 0;">Roles</h2>
		<button
			type="button"
			class="uk-button uk-button-primary uk-button-small"
			onclick={openAdd}
		>
			+ Add Role
		</button>
	</div>

	{#if pageError}
		<div class="uk-alert uk-alert-danger uk-margin-small-top" role="alert">
			{pageError}
			<button
				type="button"
				class="uk-alert-close"
				onclick={() => (pageError = '')}
				aria-label="Close"
			></button>
		</div>
	{/if}

	{#if isLoading}
		<div class="uk-text-center uk-padding">
			<div uk-spinner></div>
		</div>
	{:else}
		<SimpleTable {headers} items={roles} emptyText="No roles yet. Create one to get started.">
			{#snippet cell(item: RowItem, header: SimpleTableHeader, getCellValue: Function)}
				{#if header.field === 'permissions'}
					<div class="perm-badges">
						{#each (item as iRole).permissions as perm}
							<span class="perm-badge">{perm}</span>
						{:else}
							<span class="uk-text-muted uk-text-small">none</span>
						{/each}
					</div>
				{:else if header.field === 'actions'}
					<div class="action-btns">
						<button
							type="button"
							class="uk-icon-button"
							uk-icon="pencil"
							title="Edit"
							onclick={() => openEdit(item as iRole)}
						></button>
						<button
							type="button"
							class="uk-icon-button uk-icon-button-danger"
							uk-icon="trash"
							title="Delete"
							onclick={() => deleteRole(item as iRole)}
						></button>
					</div>
				{:else}
					{getCellValue(item, header)}
				{/if}
			{/snippet}
		</SimpleTable>
	{/if}
</div>

{#if showModal}
	<div id="role-modal" class="uk-modal uk-open" style="display: block;">
		<div class="uk-modal-dialog uk-modal-body">
			<button
				class="uk-modal-close-default"
				type="button"
				uk-close
				onclick={() => (showModal = false)}
			></button>
			<h2 class="uk-modal-title">{editingRole ? 'Edit Role' : 'New Role'}</h2>
			<RoleForm
				role={editingRole}
				onsuccess={onFormSuccess}
				oncancel={() => (showModal = false)}
			/>
		</div>
	</div>
	<div class="uk-modal-overlay" onclick={() => (showModal = false)}></div>
{/if}

<style>
	.page-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 16px;
	}

	.perm-badges {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
	}

	.perm-badge {
		display: inline-block;
		padding: 2px 7px;
		border-radius: 10px;
		font-size: 0.72rem;
		font-weight: 600;
		background: #f0fdf4;
		color: #166534;
		border: 1px solid #bbf7d0;
	}

	.action-btns {
		display: flex;
		gap: 4px;
		align-items: center;
	}

	.uk-icon-button-danger {
		color: #dc2626;
	}

	.uk-modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.4);
		z-index: 999;
	}

	#role-modal {
		position: fixed;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}

	#role-modal .uk-modal-dialog {
		max-width: 480px;
		width: 90%;
	}
</style>
