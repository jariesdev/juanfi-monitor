from typing import Union

from fastapi import Depends
from fastapi.encoders import jsonable_encoder
from starlette.responses import JSONResponse

from juanfi_api import JuanfiApi
from models.vendo import VendoMachine
from repository.vendo_repository import VendoRepository
from repository.withdrawal_repository import WithdrawalRepository
from repository.user_repository import UserRepository
from sql_app.schemas import SetVendoStatusRequest


def get_repo():
    return VendoRepository()


class VendoController():
    _repository: VendoRepository

    def __init__(
            self,
            repository: VendoRepository = Depends(VendoRepository),
            withdrawal_repository: WithdrawalRepository = Depends(WithdrawalRepository)
    ):
        self._repository = repository
        self._withdrawal_repository = withdrawal_repository

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

    def get(self, vendo_id: int):
        vendo = self._repository.get(vendo_id)
        return {
            "data": vendo
        }

    def delete(self, vendo_id: int):
        self._repository.delete_vendo(vendo_id)
        return {
            "data": None
        }

    def vendo_status(self, vendo_id: int):
        vendo = self._repository.get(vendo_id)
        status = JuanfiApi(vendo).get_system_status()
        return JSONResponse(status)

    def withdraw_current_sales(self, vendo_id: int):
        vendo = self._repository.get(vendo_id)
        status = JuanfiApi(vendo).get_system_status()
        JuanfiApi(vendo).reset_current_sales()

        withdrawal = self._withdrawal_repository.add(
            vendo_id=vendo.id,
            amount=float(status.get('current_coin_count'))
        )

        return JSONResponse({
            "message": jsonable_encoder(withdrawal)
        })

    def set_status(self, vendo_id: int, request: SetVendoStatusRequest):
        self._repository.set_status(vendo_id, request.status)
        return JSONResponse({
            "message": "success",
        })
