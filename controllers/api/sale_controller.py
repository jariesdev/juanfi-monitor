from typing import Union, List

from starlette.responses import JSONResponse

from repository.sale_repository import SaleRepository


class SaleController():
    _repository: SaleRepository

    def __init__(self):
        self._repository = SaleRepository()

    def all(self, q: Union[str, None] = None) -> JSONResponse:
        sales = self._repository.search(q)

        return JSONResponse({
            "data": self._map_list_to_dict(sales)
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

    def daily_sales(self) -> JSONResponse:
        sales = self._repository.get_daily_sales()
        return JSONResponse({
            "data": list(sales)
        })
