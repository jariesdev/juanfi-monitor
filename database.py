import sqlite3
from sqlite3 import Connection


class Database:
    def get_connection(self) -> Connection:
        return sqlite3.connect("app.db", check_same_thread=True)
