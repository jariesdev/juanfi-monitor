from typing import Union

from database import Database


class LogRepository:
    _db: Database

    def __init__(self):
        self._db = Database()

    def search(self, q: Union[str, None] = None) -> list:
        conn = self._db.get_connection()
        cur = conn.cursor()
        if q is not None:
            cur.execute(
                "SELECT * FROM juanfi_logs WHERE juanfi_logs.description LIKE ?",
                [
                    "%{}%".format(q),
                ]
            )
        else:
            cur.execute("SELECT * FROM juanfi_logs")

        rows = cur.fetchall()
        conn.close()
        return rows
