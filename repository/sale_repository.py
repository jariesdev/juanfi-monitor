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
                "SELECT * FROM juanfi_sales "
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
