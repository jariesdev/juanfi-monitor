from fastapi.encoders import jsonable_encoder
from fastapi.responses import JSONResponse

from repository.withdrawal_repository import WithdrawalRepository
from fastapi import Depends

from sql_app.models import Withdrawal


class WithdrawalController():
    def __init__(self, repository: WithdrawalRepository = Depends(WithdrawalRepository)):
        self._repository = repository

    def search(self):
        result = self._repository.search()
        return JSONResponse({
            "data": jsonable_encoder(result)
        }, status_code=200)
