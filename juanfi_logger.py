from database import Database
from juanfi_api import JuanfiApi
from sqlite3 import Connection


class JuanfiLogger():
    _juanfi: JuanfiApi
    _db_conn: Connection

    def __init__(self):
        self._juanfi = JuanfiApi()
        self._db_conn = Database().get_connection()

    def run(self):
        logs = self._juanfi.get_system_logs()
        for log in logs:
            self._db_conn.execute(
                "INSERT OR IGNORE INTO juanfi_logs(`log_time`, `description`) VALUES ('{0}','{1}')".format(*log)
            )
            self._db_conn.commit()


if __name__ == "__main__":
    juanfi_logger = JuanfiLogger()
    juanfi_logger.run()
