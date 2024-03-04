from typing import Union

from sqlalchemy import func
from sqlalchemy.orm import joinedload

from database import Database
from sql_app.database import SessionLocal
from sql_app.models import VendoSale, Vendo


class SaleRepository:
    _db: Database

    def __init__(self):
        self._db = Database()

    def search(self,
               q: Union[str, None] = None,
               date: Union[str, None] = None,
               vendo_id: Union[int, None] = None) -> list:
        db = SessionLocal()

        query = (db.query(VendoSale).options(joinedload(VendoSale.vendo))
                 .order_by(VendoSale.sale_time.desc()))

        if q is not None:
            query = (query.where(VendoSale.mac_address.like("%{}%".format(q)))
                     .where(VendoSale.voucher.like("%{}%".format(q))))

        if date is not None:
            query = (query.where(func.date(VendoSale.sale_time) == date))

        if vendo_id is not None:
            query = (query.where(VendoSale.vendo_id == vendo_id))

        rows = query.all()
        db.close()

        return rows

    def get_daily_sales(self) -> list[dict]:
        db = SessionLocal()
        query = (db.query(func.date(VendoSale.sale_time).label("date"), func.sum(VendoSale.amount).label("total"), VendoSale.vendo_id, Vendo.name.label("vendo_name"))
                 .join(VendoSale.vendo)
                 .where(VendoSale.sale_time > func.date("now", "-3 months"))
                 .group_by(func.date(VendoSale.sale_time), VendoSale.vendo_id)
                 .order_by(VendoSale.sale_time))
        rows = query.all()
        db.close()

        return rows
