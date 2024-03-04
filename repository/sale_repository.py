from typing import Union

from sqlalchemy import func
from sqlalchemy.orm import joinedload

from dependencies import get_db
from sql_app.models import VendoSale, Vendo
from fastapi import Depends
from sqlalchemy.orm import Session


class SaleRepository:
    def __init__(self, db: Session = Depends(get_db)):
        self._db = db

    def search(self,
               q: Union[str, None] = None,
               date: Union[str, None] = None,
               vendo_id: Union[int, None] = None) -> list:
        db = self._db

        query = (db.query(VendoSale).options(joinedload(VendoSale.vendo))
                 .order_by(VendoSale.sale_time.desc()))

        if q is not None:
            query = (query.where(VendoSale.mac_address.like("%{}%".format(q)))
                     .where(VendoSale.voucher.like("%{}%".format(q))))

        if date is not None:
            query = (query.where(func.date(VendoSale.sale_time) == date))

        if vendo_id is not None:
            query = (query.where(VendoSale.vendo_id == vendo_id))

        return query.all()

    def get_daily_sales(self) -> list[dict]:
        db = self._db
        query = (db.query(func.date(VendoSale.sale_time).label("date"), func.sum(VendoSale.amount).label("total"),
                          VendoSale.vendo_id, Vendo.name.label("vendo_name"))
                 .join(VendoSale.vendo)
                 .where(VendoSale.sale_time > func.date("now", "-3 months"))
                 .group_by(func.date(VendoSale.sale_time), VendoSale.vendo_id)
                 .order_by(VendoSale.sale_time))
        return query.all()
