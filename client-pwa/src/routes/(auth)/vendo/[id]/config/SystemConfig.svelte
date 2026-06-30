<script lang="ts">
	import { onMount, onDestroy } from 'svelte';

	interface SystemConfigData {
		[key: string]: string | number | boolean | string[] | undefined;
	}

	interface FieldDef {
		key: string;
		label: string;
		secret?: boolean;
		decode?: Record<string, string>;
	}

	interface SectionDef {
		title: string;
		fields: FieldDef[];
	}

	const { vendoId } = $props();

	let config: SystemConfigData | null = $state(null);
	let isLoading = $state(true);
	let loadError = $state('');
	let revealed: Record<string, boolean> = $state({});
	let controller: AbortController | undefined;

	const yesNo: Record<string, string> = { '1': 'Yes', '0': 'No' };

	const sections: SectionDef[] = [
		{
			title: 'Vendo Name',
			fields: [
				{ key: 'vendo_name', label: 'Vendo Name' },
				{ key: 'include_vendo_name', label: 'Include VendoName', decode: yesNo }
			]
		},
		{
			title: 'AP Setting',
			fields: [
				{ key: 'wifi_ssid', label: 'WiFi SSID' },
				{ key: 'wifi_password', label: 'WiFi Password', secret: true }
			]
		},
		{
			title: 'IP Address Setting',
			fields: [
				{ key: 'ip_address_mode', label: 'Mode', decode: { '0': 'DHCP', '1': 'Static' } },
				{ key: 'local_ip_address', label: 'Local IP Address' },
				{ key: 'subnet_mask', label: 'Subnet Mask' },
				{ key: 'gateway_ip', label: 'Gateway IP' },
				{ key: 'dns_server', label: 'DNS Server' }
			]
		},
		{
			title: 'MikroTik Setting',
			fields: [
				{ key: 'mikrotik_ip', label: 'Mikrotik Router IP' },
				{ key: 'mikrotik_username', label: 'Mikrotik Username' },
				{ key: 'mikrotik_password', label: 'Mikrotik Password', secret: true },
				{ key: 'connection_mode', label: 'Connection Mode' }
			]
		},
		{
			title: 'Admin Panel Setting',
			fields: [
				{ key: 'admin_username', label: 'Admin Username' },
				{ key: 'admin_password', label: 'Admin Password', secret: true }
			]
		},
		{
			title: 'Operator Panel Setting',
			fields: [
				{ key: 'operator_username', label: 'Operator Username' },
				{ key: 'operator_password', label: 'Operator Password', secret: true }
			]
		},
		{
			title: 'Coin Slot Setting',
			fields: [
				{ key: 'coin_slot_type', label: 'Coinslot Slot Type' },
				{ key: 'coin_slot_wait_time_sec', label: 'Coin slot insert wait time (Seconds)' },
				{ key: 'coin_slot_abuse_count', label: 'Coins Slot Abuse Count' },
				{ key: 'coin_slot_ban_minutes', label: 'Coins Slot Ban Minutes' },
				{ key: 'button_function', label: 'Button Function' },
				{ key: 'coin_multiplier', label: 'Coin Multiplier' },
				{ key: 'bill_acceptor_multiplier', label: 'Bill Acceptor Multiplier' }
			]
		},
		{
			title: 'Pin Setting',
			fields: [
				{ key: 'coin_slot_pin', label: 'Coinslot Pin' },
				{ key: 'bill_acceptor_pin', label: 'Bill Acceptor Pin' },
				{ key: 'coin_slot_set_pin', label: 'Coinslot Set Pin' },
				{ key: 'system_ready_led_pin', label: 'System Ready LED Pin' },
				{ key: 'insert_coin_led_pin', label: 'Insert Coin LED Pin' },
				{ key: 'insert_coin_button_pin', label: 'Insert Coin Button Pin' },
				{ key: 'night_light_pin', label: 'Night Light Pin' },
				{
					key: 'led_trigger_type',
					label: 'LED Trigger Output',
					decode: { '0': 'LOW', '1': 'HIGH' }
				},
				{ key: 'lcd_sda_pin', label: 'LCD SDA Pin' },
				{ key: 'lcd_scl_pin', label: 'LCD SCL Pin' },
				{ key: 'lan_cs_pin', label: 'LAN CS Pin' }
			]
		},
		{
			title: 'LCD Setting',
			fields: [
				{
					key: 'lcd_screen',
					label: 'LCD Screen',
					decode: { '0': 'None', '1': '16x2', '2': '20x4' }
				},
				{ key: 'welcome_lcd_marquee', label: 'Welcome LCD Marquee' }
			]
		},
		{
			title: 'Voucher Setting',
			fields: [
				{ key: 'voucher_prefix', label: 'Voucher Prefix' },
				{
					key: 'voucher_login_option',
					label: 'Voucher Login Option',
					decode: { '0': 'Username only', '1': 'Username + Password' }
				},
				{ key: 'voucher_profile', label: 'Voucher Profile' },
				{
					key: 'voucher_validity',
					label: 'Voucher Validity',
					decode: { '0': 'Stack up validity when extend time', '1': 'First Validity + extend time' }
				},
				{ key: 'voucher_length', label: 'Voucher Length' }
			]
		},
		{
			title: 'Internet Check',
			fields: [
				{
					key: 'check_internet_status',
					label: 'Check Internet Status when insert coin',
					decode: yesNo
				}
			]
		},
		{
			title: 'Thermal Printing',
			fields: [
				{ key: 'printer_pin', label: 'Printer Pin' },
				{ key: 'print_option', label: 'Print Option' }
			]
		},
		{
			title: 'Other Setting',
			fields: [
				{ key: 'api_key', label: 'API Key', secret: true },
				{ key: 'setup_done_flag', label: 'Setup Done', decode: yesNo }
			]
		}
	];

	function fieldValue(field: FieldDef): string {
		if (!config) return '';
		const raw = config[field.key];
		if (raw === undefined || raw === null || raw === '') return '—';
		const str = String(raw);
		if (field.decode && field.decode[str] !== undefined) return field.decode[str];
		return str;
	}

	function displayValue(field: FieldDef): string {
		const value = fieldValue(field);
		if (field.secret && value !== '—' && !revealed[field.key]) {
			return '•'.repeat(Math.min(value.length, 12));
		}
		return value;
	}

	function toggleReveal(key: string): void {
		revealed[key] = !revealed[key];
	}

	function loadData(): void {
		controller?.abort();
		controller = new AbortController();
		isLoading = true;
		loadError = '';

		fetch(`/x-api/vendo-machines/${vendoId}/config`, { signal: controller.signal })
			.then((r) => (r.ok ? r.json() : Promise.reject(r.statusText)))
			.then((data) => {
				config = data;
			})
			.catch((err) => {
				if (err?.name !== 'AbortError') {
					loadError = 'Unable to load configuration from the device.';
				}
			})
			.finally(() => {
				isLoading = false;
			});
	}

	onMount(loadData);
	onDestroy(() => controller?.abort());
</script>

<div class="card">
	<div class="card-header">
		<span class="header-icon" uk-icon="icon: cog; ratio: 1"></span>
		<span class="card-title">System Configuration</span>
		<button class="refresh-btn" onclick={loadData} disabled={isLoading} title="Refresh">
			<span uk-icon="icon: refresh; ratio: 0.8"></span>
		</button>
	</div>

	{#if isLoading}
		{#each sections as section}
			<div class="section">
				<span class="section-title">{section.title}</span>
				<div class="status-list">
					{#each { length: 3 } as _}
						<div class="status-row">
							<div class="skeleton skeleton-label"></div>
							<div class="skeleton skeleton-value"></div>
						</div>
					{/each}
				</div>
			</div>
		{/each}
	{:else if loadError}
		<p class="status-empty">{loadError}</p>
	{:else if config}
		{#each sections as section}
			<div class="section">
				<span class="section-title">{section.title}</span>
				<div class="status-list">
					{#each section.fields as field}
						<div class="status-row">
							<span class="status-key">{field.label}</span>
							<div class="status-val-wrap">
								<span class="status-val">{displayValue(field)}</span>
								{#if field.secret && fieldValue(field) !== '—'}
									<button type="button" class="reveal-btn" onclick={() => toggleReveal(field.key)}>
										{revealed[field.key] ? 'Hide' : 'Show'}
									</button>
								{/if}
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/each}
	{:else}
		<p class="status-empty">Configuration not available.</p>
	{/if}
</div>

<style>
	.card {
		background: #fff;
		border: 1px solid #e8e8e8;
		border-radius: 10px;
		overflow: hidden;
		margin-bottom: 12px;
	}

	.card-header {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 14px 18px;
		border-bottom: 1px solid #f0f0f0;
	}

	.section {
		border-bottom: 1px solid #f0f0f0;
	}

	.section:last-child {
		border-bottom: none;
	}

	.section-title {
		display: block;
		padding: 12px 18px 4px;
		font-size: 0.72rem;
		font-weight: 700;
		color: #aaa;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}

	.header-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 30px;
		height: 30px;
		background: #fff4f1;
		border-radius: 7px;
		color: var(--color-theme-1);
		flex-shrink: 0;
	}

	.card-title {
		font-size: 0.88rem;
		font-weight: 700;
		color: #1a1a1a;
		margin-right: auto;
	}

	.refresh-btn {
		background: none;
		border: 1px solid #e8e8e8;
		border-radius: 6px;
		padding: 5px 8px;
		cursor: pointer;
		color: #888;
		display: flex;
		align-items: center;
		transition:
			background 0.15s,
			color 0.15s;
	}

	.refresh-btn:hover:not(:disabled) {
		background: #f5f5f5;
		color: #333;
	}

	.refresh-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	.status-list {
		padding: 4px 0;
	}

	.status-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 10px 18px;
		border-bottom: 1px solid #f6f6f6;
		gap: 12px;
	}

	.status-row:last-child {
		border-bottom: none;
	}

	.status-key {
		font-size: 0.82rem;
		color: #666;
		font-weight: 500;
	}

	.status-val {
		font-size: 0.82rem;
		color: #1a1a1a;
		font-weight: 600;
		text-align: right;
		word-break: break-all;
	}

	.status-val-wrap {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.reveal-btn {
		background: none;
		border: none;
		padding: 0;
		cursor: pointer;
		color: var(--color-theme-1);
		font-size: 0.74rem;
		font-weight: 600;
		font-family: inherit;
		white-space: nowrap;
	}

	.reveal-btn:hover {
		color: #555;
	}

	.status-empty {
		padding: 20px 18px;
		font-size: 0.82rem;
		color: #bbb;
		font-style: italic;
		margin: 0;
	}

	.skeleton {
		background: linear-gradient(90deg, #f0f0f0 25%, #e8e8e8 50%, #f0f0f0 75%);
		background-size: 200% 100%;
		animation: shimmer 1.4s infinite;
		border-radius: 4px;
	}

	.skeleton-label {
		width: 130px;
		height: 13px;
	}
	.skeleton-value {
		width: 70px;
		height: 13px;
	}

	@keyframes shimmer {
		0% {
			background-position: 200% 0;
		}
		100% {
			background-position: -200% 0;
		}
	}
</style>
