from typing import Union

from database import Database


class UserRepository:
    _db: Database

    def __init__(self):
        self._db = Database()

    def search(self, q: Union[str, None] = None) -> list:
        conn = self._db.get_connection()
        cur = conn.cursor()
        if q is not None:
            cur.execute(
                "SELECT * FROM users WHERE username LIKE ? or password LIKE ?",
                [
                    "%{}%".format(q),
                    "%{}%".format(q),
                ]
            )
        else:
            cur.execute("SELECT * FROM users")

        rows = cur.fetchall()
        conn.close()
        return rows
