from datetime import date
from typing import Union

from sqlalchemy import func, or_, select
from sqlalchemy.orm import joinedload, Session

import sql_app.schemas
from dependencies import get_db
from sql_app.models import VendoSale, Vendo
from fastapi import Depends
from sqlalchemy.orm import Session
from fastapi_pagination import Page
from fastapi_pagination.ext.sqlalchemy import paginate

class SaleRepository:
    def __init__(self, db: Session = Depends(get_db)):
        self._db = db

    def search(self,
               search: Union[str, None] = None,
               date: Union[str, None] = None,
               vendo_id: Union[int, None] = None) -> Page[sql_app.schemas.VendoSale]:
        db: Session = self._db

        query = (select(VendoSale).options(joinedload(VendoSale.vendo))
                 .order_by(VendoSale.sale_time.desc()))

        if search is not None and search != "":
            query = (query.filter(
                or_(VendoSale.mac_address.like("%{}%".format(search)), VendoSale.voucher.like("%{}%".format(search)))))

        if date is not None:
            query = (query.where(func.date(VendoSale.sale_time) == date))

        if vendo_id is not None:
            query = (query.where(VendoSale.vendo_id == vendo_id))

        return paginate(db, query)

    def get_daily_sales(self, from_date: date, to_date: date) -> list[dict]:
        db = self._db
        query = (db.query(func.date(VendoSale.sale_time).label("date"), func.sum(VendoSale.amount).label("total"),
                          VendoSale.vendo_id, Vendo.name.label("vendo_name"))
                 .join(VendoSale.vendo)
                 .filter(VendoSale.sale_time >= from_date.strftime('%Y-%m-%d'))
                 .filter(VendoSale.sale_time <= to_date.strftime('%Y-%m-%d'))
                 .filter(Vendo.is_active == True)
                 .group_by(func.date(VendoSale.sale_time), VendoSale.vendo_id)
                 .order_by(VendoSale.sale_time))
        return query.all()
