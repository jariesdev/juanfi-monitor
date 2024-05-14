from repository.log_repository import LogRepository
from fastapi import Depends
from fastapi_pagination import Page
from sql_app.schemas import LogsSearchRequest


class LogController:
    _repository: LogRepository

    def __init__(self, repository: LogRepository = Depends(LogRepository)):
        self._repository = repository

    def search(self, request: LogsSearchRequest) -> Page:
        date = None
        if request.date is not None:
            date = request.date.strftime('%Y-%m-%d')

        result = self._repository.search(
            search=request.q,
            date=date,
            vendo_id=request.vendo_id)
        return result