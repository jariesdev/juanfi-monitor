from typing import Union

from sqlalchemy.orm import joinedload

from database import Database
from dependencies import get_db
from sql_app import models
from sqlalchemy import func, select
from fastapi import Depends
from sqlalchemy.orm import Session
from fastapi_pagination.ext.sqlalchemy import paginate

class LogRepository:
    _db: Database

    def __init__(self, db: Session = Depends(get_db)) -> None:
        self._db = db

    def search(
            self,
            search: Union[str, None] = None,
            date: Union[str, None] = None,
            vendo_id: Union[int, None] = None
    ) -> list:
        db = self._db

        query = (select(models.VendoLog)
                 .join(models.VendoLog.vendo)
                 .options(joinedload(models.VendoLog.vendo))
                 .order_by(models.VendoLog.log_time.desc()))

        # query = "SELECT id, log_time, description, DATETIME(created_at, 'localtime') FROM vendo_logs WHERE 1=1 ";

        if search is not None or search != "":
            query = query.where(models.VendoLog.description.like("%{}%".format(search)))

        if date is not None:
            query = query.filter(func.date(models.VendoLog.log_time) == date)

        if vendo_id is not None:
            query = query.filter(models.VendoLog.vendo_id == vendo_id)

        return paginate(self._db, query)
