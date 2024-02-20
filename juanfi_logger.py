import datetime
import time

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
        raw_logs = self._juanfi.load_system_logs()
        self.store_sales(raw_logs)
        self.store_logs()

    def store_logs(self):
        logs = self._juanfi.get_formatted_system_logs()
        for log in logs:
            self._db_conn.execute(
                "INSERT OR IGNORE INTO juanfi_logs(`log_time`, `description`, `created_at`) VALUES ('{0}', '{1}', DATETIME('now', 'localtime'))".format(*log)
            )
            self._db_conn.commit()
        print("%s: logs inserted" % datetime.datetime.now())

    def store_sales(self, logs: list):
        sale_logs = filter(lambda log: log.get("log_type_index") == 14, logs)
        for _log in sale_logs:
            rows = self._db_conn.cursor().execute(
                "SELECT * FROM juanfi_sales "
                "WHERE sale_time BETWEEN DATETIME(?, '-10 seconds') AND DATETIME(?, '+10 seconds')"
                "AND mac_address = ?",
                [
                    self._juanfi.compute_log_time(_log.get("time")),
                    self._juanfi.compute_log_time(_log.get("time")),
                    _log.get("log_params")[1],
                ]
            ).fetchall()

            if len(rows) == 0:
                self._db_conn.execute(
                    "INSERT OR IGNORE INTO juanfi_sales(`sale_time`, `mac_address`, `voucher`, `amount`, `created_at`) VALUES ('{0}', '{1}', '{2}', '{3}', DATETIME('now', 'localtime'))".format(
                        self._juanfi.compute_log_time(_log.get("time")),
                        _log.get("log_params")[0],
                        _log.get("log_params")[1],
                        _log.get("log_params")[2]
                    )
                )

            self._db_conn.commit()
        print("%s: logs inserted" % datetime.datetime.now())


if __name__ == "__main__":
    juanfi_logger = JuanfiLogger()
    juanfi_logger.run()
