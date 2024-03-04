<script>
	import { page } from '$app/stores';
</script>

<header>
	<div class="corner" />

	<nav>
		<div class="left-wing">
			<svg viewBox="0 0 2 3" aria-hidden="true">
				<path d="M0,0 L1,2 C1.5,3 1.5,3 2,3 L2,0 Z" />
			</svg>
		</div>
		<ul>
			<li aria-current={$page.url.pathname === '/home' ? 'page' : undefined}>
				<a href="/home">Home</a>
			</li>
			<li aria-current={$page.url.pathname === '/vendo' ? 'page' : undefined}>
				<a href="/vendo">Vendo</a>
			</li>
			<li aria-current={$page.url.pathname === '/sales' ? 'page' : undefined}>
				<a href="/sales">Sales</a>
			</li>
			<li aria-current={$page.url.pathname === '/logs' ? 'page' : undefined}>
				<a href="/logs">Logs</a>
			</li>
			<li aria-current={$page.url.pathname === '/logout' ? 'page' : undefined} class="uk-flex uk-align-center">
				<form action="/logout" method="POST">
				  <button type="submit" class="uk-button-link" style="border: none; cursor: pointer">
					  <i uk-icon="icon: sign-out"></i>
				  </button>
				</form>
			</li>
		</ul>
		<div class="right-wing">
			<svg viewBox="0 0 2 3" aria-hidden="true">
				<path d="M0,0 L0,3 C0.5,3 0.5,3 1,2 L2,0 Z" />
			</svg>
		</div>
	</nav>

	<div class="corner" />
</header>

<style>
	header {
		display: flex;
		justify-content: space-between;
	}

	.corner {
		width: 3em;
		height: 3em;
	}

	nav {
		display: flex;
		justify-content: center;
		--background: rgba(255, 255, 255, 0.7);
	}

	svg {
		width: 2em;
		height: 3em;
		display: block;
	}

	path {
		fill: var(--background);
	}

	ul {
		position: relative;
		padding: 0;
		margin: 0;
		height: 3em;
		display: flex;
		justify-content: center;
		align-items: center;
		list-style: none;
		background: var(--background);
		background-size: contain;
	}

	li {
		position: relative;
		height: 100%;
	}

	li[aria-current='page']::before {
		--size: 6px;
		content: '';
		width: 0;
		height: 0;
		position: absolute;
		top: 0;
		left: calc(50% - var(--size));
		border: var(--size) solid transparent;
		border-top: var(--size) solid var(--color-theme-1);
	}

	nav a,
	nav form button {
		display: flex;
		height: 100%;
		align-items: center;
		padding: 0 0.5rem;
		color: var(--color-text);
		font-weight: 700;
		font-size: 0.8rem;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		text-decoration: none;
		transition: color 0.2s linear;
	}

	a:hover {
		color: var(--color-theme-1);
	}
	.left-wing,
	.right-wing {
		min-width: 32px;
	}
</style>
