from typing import Union
from database import Database
from models.vendo import VendoMachine
from sql_app import models, schemas
from sql_app.database import SessionLocal, engine
from sqlalchemy.orm import Session
from fastapi import Depends

models.Vendo.metadata.create_all(bind=engine)


# Dependency
def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


class VendoRepository:
    _db: Database

    def __init__(self):
        self._db = Database()

    def search(self, q: Union[str, None] = None) -> list:
        conn = self._db.get_connection()
        conn.set_trace_callback(print)
        cur = conn.cursor()

        query = (
            "SELECT id, name, mac_address, api_url, is_online, current_sales, total_sales, DATETIME(created_at, 'localtime') "
            "FROM vendos WHERE 1=1 ")

        if q is not None:
            query = query + ("AND (name LIKE '%{0}%' "
                             "OR mac_address LIKE '%{0}%') ").format(q)

        query = query + "ORDER BY name "
        cur.execute(query)

        rows = cur.fetchall()
        conn.close()

        keys = ["id", "name", "mac_address", "api_url", "is_online", "current_sales", "total_sales", "created_at"]

        def mapper(d: list) -> dict:
            return dict(zip(keys, d))

        return list(map(mapper, rows))

    def add(self, name: str, url: str, api_key: str) -> None:
        self._db.execute("INSERT INTO vendos (name, url, api_key)"
                         "VALUES (?,?,?)", [name, url, api_key])

    def delete(self, id: int) -> None:
        self._db.execute("DELETE vendos WHERE id = ?", [id])

    def add_vendo(self, vendo: VendoMachine):
        conn = self._db.get_connection()
        cur = conn.cursor()
        cur.execute("INSERT INTO vendos (name, api_url, api_key) VALUES (?,?,?)",
                    [vendo.name, vendo.api_url, vendo.api_key])
        last_id = cur.lastrowid
        columns = [
            "id", "name", "mac_address", "api_url", "api_key", "total_sales", "current_sales", "is_online", "created_at"
        ]
        cur.execute(
            "SELECT {} FROM vendos WHERE id = ?".format(",".join(columns)),
            [last_id]);
        item = cur.fetchone()
        conn.commit()
        conn.close()
        output = zip(columns, list(item))

        return dict(output)

    def delete_vendo(self, id: int) -> None:
        conn = self._db.get_connection()
        conn.execute("DELETE vendos WHERE id = ?", [id])
        conn.commit()

    def all(self, db: Session = Depends(get_db)) -> list:
        return db.query(models.Vendo).all()

    def get(self, id: int) -> models.Vendo:
        db = SessionLocal()
        return db.query(models.Vendo).filter(models.Vendo.id == id).first()
