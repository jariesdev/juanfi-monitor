from typing import Union

from sqlalchemy.sql.functions import current_user

from dependencies import get_db
from fastapi import Depends
from sqlalchemy.orm import Session

from sql_app.models import Withdrawal


class WithdrawalRepository:
    _db: Session

    def __init__(self, db: Session = Depends(get_db)):
        self._db = db

    def add(self, vendo_id: int, amount: float, user_id: Union[int, None] = None) -> Withdrawal:
        withdrawal = Withdrawal(
            vendo_id=vendo_id,
            amount=amount,
            user_id=user_id
        )
        self._db.add(withdrawal)
        self._db.commit()

        return withdrawal
        