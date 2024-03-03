from typing import Union

from sqlalchemy import func
from sqlalchemy.orm import joinedload

from database import Database
from sql_app.database import SessionLocal
from sql_app.models import VendoSale


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
                 .order_by(VendoSale.sale_time))

        if q is not None:
            query = (query.where(VendoSale.mac_address.like("%{}%".format(q)))
                     .where(VendoSale.voucher.like("%{}%".format(q))))

        if date is not None:
            query = (query.where(func.date(VendoSale.sale_time) == date))

        if vendo_id is not None:
            query = (query.where(VendoSale.vendo_id == vendo_id))

        return query.all()

    def get_daily_sales(self) -> list[dict]:
        conn = self._db.get_connection()
        cur = conn.cursor()
        cur.execute("SELECT DATE(sale_time) AS date, SUM(amount) AS total "
                    "FROM vendo_sales "
                    "WHERE sale_time > DATE('now', '-3 months')"
                    "GROUP BY DATE(sale_time) "
                    "ORDER BY sale_time ASC")
        rows = cur.fetchall()
        conn.close()

        def mapper(d: list) -> dict:
            return {
                "date": d[0],
                "total": d[1],
            }

        return list(map(mapper, rows))
