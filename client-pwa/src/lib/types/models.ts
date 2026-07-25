export interface iVendoStatus {
	total_sales: number;
	vendo_id: number;
	customer_count: number;
	wireless_strength: number;
	created_at: string;
	id: number;
	current_sales: number;
	free_heap: number;
	active_users: number;
}

export interface iVendo {
	id: number;
	name: string;
	mac_address: string;
	api_url: string;
	api_key: string;
	is_online: number;
	total_sales: number;
	current_sales: number;
	active_users: number;
	created_at: string;
	is_active: boolean;
	commission: number;
	recent_status?: iVendoStatus;
}

export interface iVendoLog {
	id: number;
	log_time: string;
	description: string;
	created_at: string;
	vendo: iVendo;
}

export interface iSale {
	id: number;
	sale_time: string;
	mac_address: string;
	voucher: string;
	amount: number;
	created_at: string;
	vendo: iVendo;
}

export interface iRole {
	id: number;
	name: string;
	permissions: string[];
	created_at: string;
	updated_at: string | null;
}

export interface iUser {
	id: number;
	username: string;
	is_active: boolean;
	roles: iRole[];
	vendos?: iVendo[];
	created_at: string;
	updated_at: string | null;
}

export const ALL_PERMISSIONS = [
	'dashboard',
	'account',
	'vendos',
	'sales',
	'logs',
	'withdrawals',
	'rates',
	'vouchers',
	'settings',
	'users',
	'vendoconfig',
	'profit'
] as const;

export type Permission = (typeof ALL_PERMISSIONS)[number];

export interface iNotification {
	id: number;
	message: string;
	user_id?: number | null;
	created_at?: string;
	type?: 'info' | 'alert';
}

export interface iVendoRate {
	id: number;
	vendo_id: number | null;
	name: string;
	price: number;
	minutes: number;
	validity_minutes: number;
	data_limit_mb: number | null;
	user_profile: string | null;
	sort_order: number;
	created_at: string;
	updated_at: string | null;
}

export interface iVendoVoucher {
	id: number;
	vendo_id: number;
	code: string;
	prefix: string;
	amount: number;
	duration_minutes: number;
	added_to_sales: boolean;
	printed_thermal: boolean;
	created_at: string;
}
