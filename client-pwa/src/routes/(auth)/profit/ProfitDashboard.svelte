<script lang="ts">
	import { onMount } from 'svelte';

	interface MonthlyRow {
		month: string;
		revenue: number;
		commission: number;
		adjustments: number;
		one_time: number;
		recurring: number;
		expenses: number;
		net: number;
		cumulative_net: number;
	}
	interface YearlyRow {
		year: string;
		revenue: number;
		expenses: number;
		net: number;
		cumulative_net: number;
	}
	interface Report {
		monthly: MonthlyRow[];
		yearly: YearlyRow[];
		summary: {
			total_revenue: number;
			total_commission: number;
			total_adjustments: number;
			total_expenses: number;
			net: number;
			status: string;
		};
	}
	interface Forecast {
		avg_monthly_revenue: number;
		monthly_recurring_expense: number;
		net_monthly: number;
		cumulative_net: number;
		status: string;
		months_to_break_even: number | null;
		projected_break_even_month: string | null;
		projection: { month: string; projected_cumulative_net: number }[];
	}

	let report: Report | null = $state(null);
	let forecast: Forecast | null = $state(null);
	let isLoading = $state(true);
	let error = $state('');

	function fmt(n: number) {
		const abs = Math.abs(n);
		return (
			(n < 0 ? '-₱' : '₱') +
			new Intl.NumberFormat(undefined, {
				minimumFractionDigits: 2,
				maximumFractionDigits: 2
			}).format(abs)
		);
	}

	function fmtMonth(s: string) {
		const [y, m] = s.split('-');
		return new Date(Number(y), Number(m) - 1, 1).toLocaleDateString(undefined, {
			month: 'short',
			year: 'numeric'
		});
	}

	// Formats a month count as years + months, e.g. 27 -> "2 years 3 months".
	function fmtDuration(months: number) {
		if (months <= 0) return '0 months';
		const y = Math.floor(months / 12);
		const m = months % 12;
		const parts: string[] = [];
		if (y) parts.push(`${y} year${y > 1 ? 's' : ''}`);
		if (m) parts.push(`${m} month${m > 1 ? 's' : ''}`);
		return parts.join(' ');
	}

	const statusMeta: Record<string, { label: string; cls: string }> = {
		profit: { label: 'Profitable', cls: 'good' },
		break_even: { label: 'Break-even', cls: 'neutral' },
		loss: { label: 'Operating at a Loss', cls: 'bad' }
	};

	async function load() {
		isLoading = true;
		error = '';
		try {
			const [rRes, fRes] = await Promise.all([
				fetch('/x-api/profit-report'),
				fetch('/x-api/profit-forecast')
			]);
			if (rRes.ok) report = await rRes.json();
			else error = 'Failed to load report.';
			if (fRes.ok) forecast = await fRes.json();
		} catch {
			error = 'An unexpected error occurred.';
		} finally {
			isLoading = false;
		}
	}

	onMount(load);
</script>

{#if error}
	<div class="uk-alert uk-alert-danger" role="alert">{error}</div>
{/if}

{#if isLoading}
	<div class="loading">Loading profit report…</div>
{:else if report}
	<!-- Summary cards -->
	<div class="cards">
		<div class="stat">
			<span class="stat-label">Total Revenue</span>
			<span class="stat-value revenue">{fmt(report.summary.total_revenue)}</span>
		</div>
		{#if report.summary.total_commission !== 0}
			<div class="stat">
				<span class="stat-label">Commission</span>
				<span class="stat-value expense">-{fmt(report.summary.total_commission)}</span>
			</div>
		{/if}
		{#if report.summary.total_adjustments !== 0}
			<div class="stat">
				<span class="stat-label">Adjustments</span>
				<span
					class="stat-value"
					class:pos={report.summary.total_adjustments > 0}
					class:neg={report.summary.total_adjustments < 0}
				>
					{fmt(report.summary.total_adjustments)}
				</span>
			</div>
		{/if}
		<div class="stat">
			<span class="stat-label">Total Expenses</span>
			<span class="stat-value expense">{fmt(report.summary.total_expenses)}</span>
		</div>
		<div class="stat">
			<span class="stat-label">Net {report.summary.net < 0 ? 'Loss' : 'Profit'}</span>
			<span
				class="stat-value"
				class:pos={report.summary.net > 0}
				class:neg={report.summary.net < 0}
			>
				{fmt(report.summary.net)}
			</span>
		</div>
		<div class="stat">
			<span class="stat-label">Status</span>
			<span class="status-pill {statusMeta[report.summary.status]?.cls ?? 'neutral'}">
				{statusMeta[report.summary.status]?.label ?? report.summary.status}
			</span>
		</div>
	</div>

	<!-- Forecast -->
	{#if forecast}
		<div class="forecast-card {forecast.status}">
			<div class="forecast-head">
				<span uk-icon="icon: bolt"></span>
				<span class="forecast-title">Break-even Forecast</span>
			</div>
			{#if forecast.status === 'recovered'}
				<p class="forecast-headline good">✓ Investment recovered — you're in profit.</p>
				<p class="forecast-sub">Cumulative net is positive at {fmt(forecast.cumulative_net)}.</p>
			{:else if forecast.status === 'on_track'}
				<p class="forecast-headline">
					Break-even in ~<strong>{fmtDuration(forecast.months_to_break_even ?? 0)}</strong>
					{#if forecast.projected_break_even_month}
						<span class="est">(est. {fmtMonth(forecast.projected_break_even_month)})</span>
					{/if}
				</p>
				<p class="forecast-sub">
					That's about {forecast.months_to_break_even}
					{forecast.months_to_break_even === 1 ? 'month' : 'months'} at the current pace — you still
					need to recover {fmt(-forecast.cumulative_net)}.
				</p>
			{:else}
				<p class="forecast-headline bad">Not recovering at the current rate.</p>
				<p class="forecast-sub">
					Net monthly cash flow is {fmt(forecast.net_monthly)} — revenue isn't outpacing recurring costs,
					so the investment won't be recovered without a change.
				</p>
			{/if}

			<div class="forecast-metrics">
				<div>
					<span class="m-label">Avg monthly revenue</span><span class="m-val"
						>{fmt(forecast.avg_monthly_revenue)}</span
					>
				</div>
				<div>
					<span class="m-label">Monthly recurring cost</span><span class="m-val"
						>{fmt(forecast.monthly_recurring_expense)}</span
					>
				</div>
				<div>
					<span class="m-label">Net monthly</span><span
						class="m-val"
						class:pos={forecast.net_monthly > 0}
						class:neg={forecast.net_monthly < 0}>{fmt(forecast.net_monthly)}</span
					>
				</div>
				<div>
					<span class="m-label">Cumulative net to date</span><span
						class="m-val"
						class:pos={forecast.cumulative_net > 0}
						class:neg={forecast.cumulative_net < 0}>{fmt(forecast.cumulative_net)}</span
					>
				</div>
			</div>
		</div>
	{/if}

	<!-- Monthly table -->
	<div class="table-card">
		<div class="table-title">Monthly Breakdown</div>
		<div class="table-wrap">
			<table class="data-table">
				<thead>
					<tr>
						<th>Month</th>
						<th class="ta-right">Revenue</th>
						<th class="ta-right">Commission</th>
						<th class="ta-right">Adjust.</th>
						<th class="ta-right">One-time</th>
						<th class="ta-right">Recurring</th>
						<th class="ta-right">Expenses</th>
						<th class="ta-right">Net</th>
						<th class="ta-right">Cumulative</th>
					</tr>
				</thead>
				<tbody>
					{#if report.monthly.length === 0}
						<tr><td colspan="9" class="empty-cell">No data in range.</td></tr>
					{:else}
						{#each report.monthly as row}
							<tr>
								<td class="month">{fmtMonth(row.month)}</td>
								<td class="ta-right">{fmt(row.revenue)}</td>
								<td
									class="ta-right"
									class:neg={row.commission > 0}
									class:muted={row.commission === 0}
									>{row.commission > 0 ? '-' + fmt(row.commission) : fmt(0)}</td
								>
								<td
									class="ta-right"
									class:pos={row.adjustments > 0}
									class:neg={row.adjustments < 0}
									class:muted={row.adjustments === 0}>{fmt(row.adjustments)}</td
								>
								<td class="ta-right muted">{fmt(row.one_time)}</td>
								<td class="ta-right muted">{fmt(row.recurring)}</td>
								<td class="ta-right">{fmt(row.expenses)}</td>
								<td class="ta-right" class:pos={row.net > 0} class:neg={row.net < 0}
									>{fmt(row.net)}</td
								>
								<td
									class="ta-right bold"
									class:pos={row.cumulative_net > 0}
									class:neg={row.cumulative_net < 0}>{fmt(row.cumulative_net)}</td
								>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>

	<!-- Yearly table -->
	<div class="table-card">
		<div class="table-title">Yearly Summary</div>
		<div class="table-wrap">
			<table class="data-table">
				<thead>
					<tr>
						<th>Year</th>
						<th class="ta-right">Revenue</th>
						<th class="ta-right">Expenses</th>
						<th class="ta-right">Net</th>
						<th class="ta-right">Cumulative</th>
					</tr>
				</thead>
				<tbody>
					{#if report.yearly.length === 0}
						<tr><td colspan="5" class="empty-cell">No data in range.</td></tr>
					{:else}
						{#each report.yearly as row}
							<tr>
								<td class="month">{row.year}</td>
								<td class="ta-right">{fmt(row.revenue)}</td>
								<td class="ta-right">{fmt(row.expenses)}</td>
								<td class="ta-right" class:pos={row.net > 0} class:neg={row.net < 0}
									>{fmt(row.net)}</td
								>
								<td
									class="ta-right bold"
									class:pos={row.cumulative_net > 0}
									class:neg={row.cumulative_net < 0}>{fmt(row.cumulative_net)}</td
								>
							</tr>
						{/each}
					{/if}
				</tbody>
			</table>
		</div>
	</div>
{/if}

<style>
	.loading {
		padding: 40px;
		text-align: center;
		color: #999;
	}

	.cards {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
		gap: 12px;
		margin-bottom: 16px;
	}
	.stat {
		background: #fff;
		border: 1px solid #e8e8e8;
		border-radius: 10px;
		padding: 14px 16px;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.stat-label {
		font-size: 0.72rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: #aaa;
	}
	.stat-value {
		font-size: 1.25rem;
		font-weight: 700;
		color: #1a1a1a;
		font-variant-numeric: tabular-nums;
	}
	.stat-value.revenue {
		color: #059669;
	}
	.stat-value.expense {
		color: #dc2626;
	}
	.pos {
		color: #059669 !important;
	}
	.neg {
		color: #dc2626 !important;
	}

	.status-pill {
		align-self: flex-start;
		padding: 4px 12px;
		border-radius: 20px;
		font-size: 0.82rem;
		font-weight: 700;
	}
	.status-pill.good {
		background: #d1fae5;
		color: #047857;
	}
	.status-pill.bad {
		background: #fee2e2;
		color: #b91c1c;
	}
	.status-pill.neutral {
		background: #f1f5f9;
		color: #475569;
	}

	.forecast-card {
		background: #fff;
		border: 1px solid #e8e8e8;
		border-radius: 10px;
		padding: 18px 20px;
		margin-bottom: 16px;
		border-left: 4px solid #94a3b8;
	}
	.forecast-card.recovered {
		border-left-color: #059669;
	}
	.forecast-card.on_track {
		border-left-color: #2563eb;
	}
	.forecast-card.not_recovering {
		border-left-color: #dc2626;
	}

	.forecast-head {
		display: flex;
		align-items: center;
		gap: 8px;
		color: #64748b;
		margin-bottom: 8px;
	}
	.forecast-title {
		font-size: 0.75rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	.forecast-headline {
		font-size: 1.15rem;
		font-weight: 700;
		color: #1a1a1a;
		margin: 0 0 4px;
	}
	.forecast-headline.good {
		color: #047857;
	}
	.forecast-headline.bad {
		color: #b91c1c;
	}
	.forecast-headline .est {
		font-size: 0.85rem;
		font-weight: 500;
		color: #64748b;
	}
	.forecast-sub {
		font-size: 0.85rem;
		color: #64748b;
		margin: 0 0 14px;
	}

	.forecast-metrics {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		gap: 10px;
	}
	.forecast-metrics > div {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.m-label {
		font-size: 0.7rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		color: #aaa;
	}
	.m-val {
		font-size: 0.95rem;
		font-weight: 700;
		color: #1a1a1a;
		font-variant-numeric: tabular-nums;
	}

	.table-card {
		background: #fff;
		border: 1px solid #e8e8e8;
		border-radius: 10px;
		overflow: hidden;
		margin-bottom: 16px;
	}
	.table-title {
		font-size: 0.88rem;
		font-weight: 700;
		color: #1a1a1a;
		padding: 14px 18px;
		border-bottom: 1px solid #f0f0f0;
	}
	.table-wrap {
		overflow-x: auto;
	}
	.data-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.85rem;
	}
	.data-table th {
		padding: 9px 16px;
		text-align: left;
		font-size: 0.72rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: #aaa;
		background: #fafafa;
		white-space: nowrap;
	}
	.data-table td {
		padding: 10px 16px;
		border-bottom: 1px solid #f6f6f6;
		white-space: nowrap;
		font-variant-numeric: tabular-nums;
	}
	.data-table tbody tr:last-child td {
		border-bottom: none;
	}
	.data-table tbody tr:hover td {
		background: #fafafa;
	}
	.ta-right {
		text-align: right;
	}
	.month {
		color: #333;
		font-weight: 600;
	}
	.muted {
		color: #999;
	}
	.bold {
		font-weight: 700;
	}
	.empty-cell {
		text-align: center;
		color: #bbb;
		font-style: italic;
		padding: 28px 16px !important;
	}
</style>
