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
        query = (db.query(func.ceil(func.max(VendoStatus.active_users)).label('average_active_users'),
                                func.avg(VendoStatus.free_heap).label('average_free_heap'),
                          Vendo.name)
                 .join(Vendo, VendoStatus.vendo_id == Vendo.id)
                 .order_by(VendoStatus.created_at.desc())
                 .group_by(func.strftime('%Y%m%d%H', VendoStatus.created_at)))

        if vendo_id is not None:
            query = query.filter(VendoStatus.vendo_id == vendo_id)

        if from_date is not None:
            query = query.filter(VendoStatus.created_at >= from_date)

        if from_date is not None:
            query = query.filter(VendoStatus.created_at <= to_date)

        rows = query.all()
        db.close()

        def mapper(row) -> dict:
            return {
                "average_active_users": row[0],
                "average_free_heap": row[1],
                "vendo_name": row[2],
            }

        return list(map(mapper, rows))
