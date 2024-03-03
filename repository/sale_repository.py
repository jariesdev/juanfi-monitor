from typing import Union
from database import Database


class SaleRepository:
    _db: Database

    def __init__(self):
        self._db = Database()

    def search(self, q: Union[str, None] = None, date: Union[str, None] = None) -> list:
        conn = self._db.get_connection()
        conn.set_trace_callback(print)
        cur = conn.cursor()

        query = ("SELECT id, sale_time, mac_address, voucher, amount, DATETIME(created_at, 'localtime') "
                 "FROM vendo_sales WHERE 1=1 ")

        if q is not None:
            query = query + ("AND (vendo_sales.mac_address LIKE '%{0}%' "
                             "OR vendo_sales.voucher LIKE '%{0}%') ").format(q)

        if date is not None:
            query = query + "AND DATE(vendo_logs.sale_time) = '{}' ".format(date)

        query = query + "ORDER BY sale_time DESC "
        cur.execute(query)

        rows = cur.fetchall()
        conn.close()
        return rows

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
