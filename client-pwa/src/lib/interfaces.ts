export interface iVendo {
    id: number
    name: string
    mac_address: string
    api_url: string
    api_key: string
    is_online: number
    total_sales: number
    current_sales: number
    created_at: string
}


export interface iVendoLog {
    id: number
    log_time: string
    description: string
    created_at: string
    vendo: iVendo
}