from typing import Union

from fastapi import Depends
from fastapi.encoders import jsonable_encoder
from fastapi.responses import JSONResponse

from repository.vendo_status_repository import VendoStatusRepository


class VendoStatusController:
    def __init__(
            self,
            repository: VendoStatusRepository = Depends(VendoStatusRepository)
    ):
        self._repository = repository

    def search(self, vendo_id: Union[int, None] = None, from_date: Union[str, None] = None,
               to_date: Union[str, None] = None,
               active_only: bool = False) -> Union[dict, JSONResponse]:
        result = self._repository.get_hourly_status(vendo_id=vendo_id,
                                                    from_date=from_date,
                                                    to_date=to_date,
                                                    active_only=active_only)

        return JSONResponse({
            "data": jsonable_encoder(result)
        })
