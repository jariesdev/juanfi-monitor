from typing import Union, List
from fastapi.encoders import jsonable_encoder
from fastapi.responses import JSONResponse
from fastapi import Depends

from repository.sale_repository import SaleRepository


class SaleController():
    _repository: SaleRepository

    def __init__(self, repository: SaleRepository = Depends(SaleRepository)):
        self._repository = repository

    def search(self, q: Union[str, None] = None, date: Union[str, None] = None,
               vendo_id: Union[int, None] = None) -> JSONResponse:
        result = self._repository.search(q, date, vendo_id)

        return JSONResponse({
            "data": jsonable_encoder(result)
        })

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
