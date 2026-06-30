<svelte:head>
	<title>Login</title>
</svelte:head>

<script lang="ts">
	import LoginForm from './LoginForm.svelte';
</script>

<div class="login-wrap">
	<div class="card">
		<div class="card-header">
			<div class="brand">Vendo Monitoring</div>
			<p class="subtitle">Sign in to your account</p>
		</div>
		<div class="card-body">
			<LoginForm />
		</div>
	</div>
</div>

<style>
	.login-wrap {
		min-height: 100dvh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 24px 16px;
	}

	.card {
		background: #fff;
		border: 1px solid #e8e8e8;
		border-radius: 12px;
		width: 100%;
		max-width: 360px;
		overflow: hidden;
		box-shadow: 0 4px 24px rgba(0, 0, 0, 0.07);
	}

	.card-header {
		padding: 28px 28px 20px;
		border-bottom: 1px solid #f0f0f0;
		text-align: center;
	}

	.brand {
		font-size: 1.1rem;
		font-weight: 700;
		color: var(--color-theme-1);
		letter-spacing: -0.01em;
	}

	.subtitle {
		font-size: 0.78rem;
		color: #aaa;
		margin: 4px 0 0;
	}

	.card-body {
		padding: 24px 28px 28px;
	}
</style>
