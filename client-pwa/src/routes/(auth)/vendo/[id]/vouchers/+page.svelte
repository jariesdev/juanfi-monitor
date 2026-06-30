<script lang="ts">
	import VouchersTable from '$lib/components/VouchersTable.svelte';

	const { data } = $props();
	const vendoId = $derived(Number(data.id));
</script>

<svelte:head>
	<title>Vouchers</title>
</svelte:head>

<div class="uk-section">
	<div class="uk-container">
		<div class="page-header">
			<a href="/vendo/{vendoId}/status" class="back-btn">
				<span uk-icon="icon: arrow-left; ratio: 0.9"></span>
				Back
			</a>
			<h1 class="page-title">Vouchers</h1>
		</div>

		<VouchersTable {vendoId} />
	</div>
</div>

<style>
	.page-header {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 16px;
	}

	.back-btn {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		padding: 6px 12px;
		border: 1px solid #e8e8e8;
		border-radius: 7px;
		font-size: 0.82rem;
		font-weight: 500;
		color: #555;
		text-decoration: none;
		background: #fff;
		transition: background 0.15s;
	}

	.back-btn:hover {
		background: #f5f5f5;
		color: #333;
	}

	.page-title {
		font-size: 1rem;
		font-weight: 700;
		color: #1a1a1a;
		margin: 0;
	}
</style>
