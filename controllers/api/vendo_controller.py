from typing import Union

from fastapi import Depends
from fastapi.encoders import jsonable_encoder
from starlette.responses import JSONResponse

from juanfi_api import JuanfiApi
from models.vendo import VendoMachine
from repository.vendo_repository import VendoRepository


def get_repo():
    return VendoRepository()


class VendoController():
    _repository: VendoRepository

    def __init__(self, repository: VendoRepository = Depends(VendoRepository)):
        self._repository = repository

    def all(self, q: Union[str, None] = None) -> JSONResponse:
        vendo_machines = self._repository.search(q)
        return JSONResponse({
            "data": jsonable_encoder(vendo_machines)
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

    def store(self, vendo: VendoMachine):
        vendo = self._repository.add_vendo(vendo)
        return {
            "data": vendo
        }

    def delete(self, id: int):
        self._repository.delete_vendo(id)
        return {
            "data": None
        }

    def vendo_status(self, id:int):
        vendo = self._repository.get(id)
        status = JuanfiApi(vendo).get_system_status()
        return JSONResponse(status)
