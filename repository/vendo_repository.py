from typing import Union

import sql_app.models
from database import Database
from dependencies import get_db
from models.vendo import VendoMachine
from sql_app import models
from sql_app.database import SessionLocal, engine
from sqlalchemy.orm import Session, joinedload
from fastapi import Depends

models.Vendo.metadata.create_all(bind=engine)


class VendoRepository:
    _db_session: Session
    _db: Database

    def __init__(self, db: Session = Depends(get_db)):
        self._db = Database()
        self._db_session = db

    def search(self, q: Union[str, None] = None, is_active: bool | None = None) -> list:
        db = self._db_session

        query = (db.query(sql_app.models.Vendo)
                 .options(joinedload(sql_app.models.Vendo.recent_status))
                 .order_by(models.Vendo.name))

        if q is not None:
            query = query.where(models.Vendo.name.like("%{0}%".format(q)))

        if is_active is not None:
            query = query.where(models.Vendo.is_active == is_active)

        return query.all()

    def add(self, name: str, url: str, api_key: str) -> None:
        self._db.execute("INSERT INTO vendos (name, api_url, api_key)"
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

    def allActive(self, db: Session = Depends(get_db)) -> list:
        return db.query(models.Vendo).where(models.Vendo.is_active == True).all()

    def get(self, id: int) -> models.Vendo:
        db = self._db_session
        return db.query(models.Vendo).filter(models.Vendo.id == id).first()

    def set_status(self, id: int, status: bool) -> None:
        db = self._db_session
        db.query(models.Vendo).filter(models.Vendo.id == id).update({models.Vendo.is_active: status})
        db.commit()
