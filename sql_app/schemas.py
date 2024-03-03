from datetime import datetime

from pydantic import BaseModel


class VendoBase(BaseModel):
    name: str
    api_url: str | None = None
    api_key: str | None = None


class VendoCreate(VendoBase):
    pass


class Vendo(BaseModel):
    mac_address: str | None = None
    is_online: bool = False
    created_at: datetime | None = None
    total_sales: float = 0
    current_sales: float = 0
