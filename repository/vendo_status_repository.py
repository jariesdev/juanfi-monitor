from sql_app.database import SessionLocal
from fastapi import Depends
from sqlalchemy.orm import Session
from sqlalchemy import func

from sql_app.models import VendoStatus, Vendo
from typing import Union


class VendoStatusRepository:
    _db: Session

    def __init__(self):
        self._db = SessionLocal()

    def get_hourly_status(self, vendo_id: Union[int, None] = None, from_date: Union[str, None] = None,
                          to_date: Union[str, None] = None) -> list:
        db = SessionLocal()
        query = (db.query(func.strftime('%Y-%m-%d %H:00', VendoStatus.created_at),
                          func.max(VendoStatus.active_users).label('average_active_users'),
                          func.avg(VendoStatus.free_heap).label('average_free_heap'),
                          Vendo.name)
                 .join(Vendo, VendoStatus.vendo_id == Vendo.id)
                 .order_by(VendoStatus.created_at)
                 .group_by(func.strftime('%Y%m%d%H', VendoStatus.created_at))
                 .group_by(VendoStatus.vendo_id))


        if vendo_id is not None:
            query = query.filter(VendoStatus.vendo_id == vendo_id)

        if from_date is not None:
            query = query.filter(VendoStatus.created_at >= from_date)

        if to_date is not None:
            query = query.filter(VendoStatus.created_at <= to_date)

        rows = query.all()
        db.close()

        def mapper(row) -> dict:
            return {
                "time": row[0],
                "average_active_users": row[1],
                "average_free_heap": row[2],
                "vendo_name": row[3],
            }

        items = list(map(mapper, rows))
        return items
