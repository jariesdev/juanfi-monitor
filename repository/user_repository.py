from typing import Union

from fastapi import Depends

from dependencies import get_db
from sql_app import models
from sqlalchemy.orm import Session


class UserRepository:
    _db: Session

    def __init__(self, db: Session = Depends(get_db)) -> None:
        self._db = db

    def check_user(self, username: str, password: str) -> Union[models.User, None]:
        db = self._db

        user_model = (db.query(models.User)
                      .where(models.User.username == username)
                      .where(models.User.password == password)
                      .first())

        return user_model

    def get_user_by_token(self, token: str) -> Union[models.User, None]:
        db = self._db
        # TODO add more secure way of authenticating token
        return (db.query(models.User)
                .where(models.User.username == token)
                .first())
        pass
