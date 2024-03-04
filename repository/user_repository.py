from typing import Union

from sqlalchemy.sql.functions import user

from sql_app import models
from sql_app.database import SessionLocal


class UserRepository:

    def check_user(self, username: str, password: str) -> Union[models.User, None]:
        db = SessionLocal()

        user_model = (db.query(models.User)
                      .where(models.User.username == username)
                      .where(models.User.password == password)
                      .first())

        return user_model

    def get_user_by_token(self, token: str) -> Union[models.User, None]:
        db = SessionLocal()
        # TODO add more secure way of authenticating token
        return (db.query(models.User)
                .where(models.User.username == token)
                .first())
        pass
