from typing import Union

from starlette.responses import JSONResponse

from repository.sale_repository import SaleRepository


class SaleController():
    _repository: SaleRepository

    def __init__(self):
        self._repository = SaleRepository()

    def all(self, q: Union[str, None] = None) -> JSONResponse:
        sales = self._repository.search(q)
        print(sales)
        def addition(d: list) -> dict:
            return {
                "id": d[0],
                "sale_time": d[1],
                "mac_address": d[2],
                "voucher": d[3],
                "amount": d[4],
                "created_at": d[5],
            }

        return JSONResponse({
            "data": list(map(addition, sales))
        })
