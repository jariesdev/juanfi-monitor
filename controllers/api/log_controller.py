from typing import Union

from fastapi.encoders import jsonable_encoder
from fastapi.responses import JSONResponse

from repository.log_repository import LogRepository
from fastapi import Depends

class LogController:
    _repository: LogRepository

    def __init__(self, repository: Depends(LogRepository)):
        self._repository = repository

    def search(self, q: Union[str, None] = None, date: Union[str, None] = None, vendo_id: Union[int, None] = None):
        result = self._repository.search(q, date, vendo_id)
        return JSONResponse(content={
            "data": jsonable_encoder(result)
        })
