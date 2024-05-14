from typing import Union, List
from fastapi.encoders import jsonable_encoder
from fastapi.responses import JSONResponse
from fastapi import Depends

from repository.sale_repository import SaleRepository
from sql_app.schemas import SalesSearchRequest
from fastapi_pagination import Page


class SaleController():
    _repository: SaleRepository

    def __init__(self, repository: SaleRepository = Depends(SaleRepository)):
        self._repository = repository

    def search(self, request: SalesSearchRequest) -> Page:
        date = None
        if request.date is not None:
            date = request.date.strftime("%Y-%m-%d")

        result = self._repository.search(
            search=request.q,
            date=date,
            vendo_id=request.vendo_id
        )
        return result

    def _map_list_to_dict(self, sales: list) -> list[dict]:
        def addition(d: list) -> dict:
            return {
                "id": d[0],
                "sale_time": d[1],
                "mac_address": d[2],
                "voucher": d[3],
                "amount": d[4],
                "created_at": d[5],
            }

        return list(map(addition, sales))

    def daily_sales(self):
        result = self._repository.get_daily_sales()
        return JSONResponse({
            "data": [dict(r._mapping) for r in result]
        })
