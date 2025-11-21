from fastapi import Depends
from sqlalchemy.orm import Session
from starlette.responses import JSONResponse

from repository.reward_repository import RewardRepository
from sql_app.database import SessionLocal


class RewardController:
    def __init__(self, repository: RewardRepository = Depends(RewardRepository)):
        self._repository = repository

    def customer_reward(self, mac_address: str):
        points = self._repository.get_reward_by_mac(mac_address)
        return JSONResponse(status_code=200, content={
            "data": {
                "points": points
            }
        })
        # return points
