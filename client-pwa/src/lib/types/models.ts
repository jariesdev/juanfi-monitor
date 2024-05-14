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
	created_at: string;
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

export interface iUser {
	username: string;
	is_active: boolean;
}
