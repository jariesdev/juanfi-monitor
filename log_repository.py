from typing import Union
from database import Database


class LogRepository:
    _db: Database

    def __init__(self):
        self._db = Database()

    def search(self, q: Union[str, None] = None, date: Union[str, None] = None) -> list:
        conn = self._db.get_connection()
        conn.set_trace_callback(print)
        cur = conn.cursor()

        query = "SELECT id, log_time, description, DATETIME(created_at, 'localtime') FROM vendo_logs WHERE 1=1 ";

        if q is not None:
            query = query + "AND vendo_logs.description LIKE '%{}%' ".format(q)

        if date is not None:
            query = query + "AND DATE(vendo_logs.log_time) = '{}'".format(date)

        query = query + "ORDER BY log_time DESC "
        cur.execute(query)

        rows = cur.fetchall()
        conn.close()
        return rows
