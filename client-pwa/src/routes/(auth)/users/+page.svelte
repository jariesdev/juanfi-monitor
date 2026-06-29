<script lang="ts">
	import DataTable from '$lib/components/DataTable.svelte';
	import UserForm from '$lib/components/UserForm.svelte';
	import type { iRole, iUser, iVendo } from '$lib/types/models';
	import type { RowItem, TableHeader } from '$lib/types/datatable';

	let dataTable: DataTable;
	let editingUser: iUser | null = $state(null);
	let showModal = $state(false);
	let roles: iRole[] = $state([]);
	let vendos: iVendo[] = $state([]);
	let deleteError = $state('');

	const tableHeaders: TableHeader[] = [
		{ label: 'ID',         field: 'id',         sortable: true  },
		{ label: 'Username',   field: 'username',   sortable: true  },
		{ label: 'Role',       field: 'role'                        },
		{ label: 'Active',     field: 'is_active',  sortable: true  },
		{ label: 'Created',    field: 'created_at', sortable: true  },
		{ label: '',           field: 'actions'                     }
	];

	async function openAdd() {
		await loadFormData();
		editingUser = null;
		showModal = true;
	}

	async function openEdit(user: iUser) {
		await loadFormData();
		editingUser = user;
		showModal = true;
	}

	async function loadFormData() {
		const [rolesRes, vendosRes] = await Promise.all([
			fetch('/x-api/roles'),
			fetch('/x-api/vendo-machines')
		]);
		if (rolesRes.ok) {
			const r = await rolesRes.json();
			roles = r.data ?? [];
		}
		if (vendosRes.ok) {
			const v = await vendosRes.json();
			vendos = v.data ?? [];
		}
	}

	async function deleteUser(user: iUser) {
		deleteError = '';
		if (!confirm(`Delete user "${user.username}"? This cannot be undone.`)) return;

		const res = await fetch(`/x-api/users/${user.id}`, { method: 'DELETE' });
		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			deleteError = body.detail ?? 'Failed to delete user.';
			return;
		}
		dataTable.refresh();
	}

	function onFormSuccess() {
		showModal = false;
		dataTable.refresh();
	}
</script>

{#if deleteError}
	<div class="uk-alert uk-alert-danger uk-margin-small-bottom" role="alert">
		{deleteError}
		<button
			type="button"
			class="uk-alert-close"
			onclick={() => (deleteError = '')}
			aria-label="Close"
		></button>
	</div>
{/if}

<DataTable
	bind:this={dataTable}
	url="/x-api/users"
	headers={tableHeaders}
	filters={{}}
	title="Users"
>
	{#snippet titleActions()}
		<button
			type="button"
			class="uk-icon-button"
			uk-icon="icon: plus-circle"
			style="border: none;"
			title="Add user"
			aria-label="Add user"
			onclick={openAdd}
		></button>
	{/snippet}

	{#snippet cell(item: RowItem, header: TableHeader, getCellValue: Function)}
		{#if header.field === 'role'}
			{#if (item as iUser).role}
				<span class="role-badge">{(item as iUser).role!.name}</span>
			{:else}
				<span class="uk-text-muted uk-text-small">—</span>
			{/if}
		{:else if header.field === 'is_active'}
			{#if (item as iUser).is_active}
				<span class="status-badge active">Active</span>
			{:else}
				<span class="status-badge inactive">Inactive</span>
			{/if}
		{:else if header.field === 'created_at'}
			{new Date((item as iUser).created_at).toLocaleDateString()}
		{:else if header.field === 'actions'}
			<div class="action-btns">
				<button
					type="button"
					class="uk-icon-button"
					uk-icon="pencil"
					title="Edit"
					onclick={() => openEdit(item as iUser)}
				></button>
				<button
					type="button"
					class="uk-icon-button uk-icon-button-danger"
					uk-icon="trash"
					title="Delete"
					onclick={() => deleteUser(item as iUser)}
				></button>
			</div>
		{:else}
			{getCellValue(item, header)}
		{/if}
	{/snippet}

	{#snippet afterTable()}
		{#if showModal}
			<div id="user-modal" uk-modal class="uk-modal uk-open" style="display: block;">
				<div class="uk-modal-dialog uk-modal-body">
					<button
						class="uk-modal-close-default"
						type="button"
						uk-close
						onclick={() => (showModal = false)}
					></button>
					<h2 class="uk-modal-title">{editingUser ? 'Edit User' : 'New User'}</h2>
					<UserForm
						user={editingUser}
						{roles}
						{vendos}
						onsuccess={onFormSuccess}
						oncancel={() => (showModal = false)}
					/>
				</div>
			</div>
			<div class="uk-modal-overlay" onclick={() => (showModal = false)}></div>
		{/if}
	{/snippet}
</DataTable>

<style>
	.action-btns {
		display: flex;
		gap: 4px;
		align-items: center;
	}

	.role-badge {
		display: inline-block;
		padding: 2px 8px;
		border-radius: 12px;
		font-size: 0.78rem;
		font-weight: 600;
		background: #eef2ff;
		color: #3730a3;
	}

	.status-badge {
		display: inline-block;
		padding: 2px 8px;
		border-radius: 12px;
		font-size: 0.78rem;
		font-weight: 600;
	}

	.status-badge.active {
		background: #dcfce7;
		color: #166534;
	}

	.status-badge.inactive {
		background: #fee2e2;
		color: #991b1b;
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

	:global(#user-modal) {
		z-index: 1000;
	}
</style>
