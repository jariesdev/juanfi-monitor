from datetime import datetime

from fastapi import Depends

import sql_app.schemas
from dependencies import get_db
from sql_app import models
from sqlalchemy.orm import Session


class NotificationRepository:
    _db: Session

    def __init__(self, db: Session = Depends(get_db)) -> None:
        self._db = db

    def get_unread(self) -> list[sql_app.schemas.Notification]:
        db = self._db

        query = (db.query(models.Notification)
                  .where(models.Notification.read_at == None)
                 .order_by(models.Notification.read_at.asc()))

        return query.all()

    def pull_unread(self) -> list[sql_app.schemas.Notification]:
        db = self._db

        query = (db.query(models.Notification)
                  .where(models.Notification.read_at == None)
                 .order_by(models.Notification.read_at.asc()))

        unread = query.all()

        for notification in unread:
            (db.query(models.Notification)
             .filter(models.Notification.id == notification.id)
             .update({"read_at": datetime.now()}))
        db.commit()

        return unread

    def add(self, notification: sql_app.schemas.Notification) -> sql_app.schemas.Notification:
        db = self._db
        db.add(notification)
        db.commit()
        db.refresh(notification)

        return notification
