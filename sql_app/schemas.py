from datetime import datetime

from pydantic import BaseModel, ConfigDict


class UserBase(BaseModel):
    username: str
    password: str

    model_config = ConfigDict(from_attributes=True)


class UserCreate(UserBase):
    pass


class User(UserBase):
    pass
    is_active: bool
    created_at: datetime
    updated_at: datetime | None = None


class VendoBase(BaseModel):
    name: str
    api_url: str | None = None
    api_key: str | None = None

    model_config = ConfigDict(from_attributes=True)


class VendoCreate(VendoBase):
    pass


class Vendo(BaseModel):
    mac_address: str | None = None
    is_online: bool = False
    total_sales: float = 0
    current_sales: float = 0
    created_at: datetime
    updated_at: datetime | None = None
