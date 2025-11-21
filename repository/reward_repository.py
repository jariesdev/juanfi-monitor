from fastapi import Depends
from sqlalchemy.orm import Session

from dependencies import get_db
from sql_app.models import Reward


class RewardRepository:
    def __init__(self, db: Session = Depends(get_db)):
        self._db = db

    def get_reward_by_mac(self, mac: str) -> float:
        reward = self._db.query(Reward).where(Reward.mac_address == mac).first()
        print(reward)
        if reward is None:
            return 0
        return reward.points
