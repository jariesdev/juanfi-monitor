from typing import Union
from database import Database


class SaleRepository:
    _db: Database

    def __init__(self):
        self._db = Database()

    def search(self, q: Union[str, None] = None) -> list:
        conn = self._db.get_connection()
        cur = conn.cursor()
        if q is not None:
            cur.execute(
                "SELECT id, sale_time, mac_address, voucher, amount, DATETIME(created_at, 'localtime') "
                "FROM juanfi_sales "
                "WHERE juanfi_sales.mac_address LIKE ? "
                "OR juanfi_sales.voucher LIKE ? "
                "ORDER BY sale_time DESC",
                [
                    "%{}%".format(q),
                    "%{}%".format(q),
                ]
            )
        else:
            cur.execute("SELECT * FROM juanfi_sales ORDER BY sale_time DESC")

        rows = cur.fetchall()
        conn.close()
        return rows

    def get_daily_sales(self) -> list[dict]:
        conn = self._db.get_connection()
        cur = conn.cursor()
        cur.execute("SELECT DATE(sale_time) AS date, SUM(amount) AS total "
                    "FROM juanfi_sales "
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
