from datetime import datetime

from pydantic import BaseModel


class VendoMachine(BaseModel):
    name: str
    mac_address: str | None = None
    api_url: str | None = None
    api_key: str | None = None
    is_online: bool = False
    created_at: datetime | None = None
    total_sales: float = 0
    current_sales: float = 0
