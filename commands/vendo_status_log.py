import logging
from datetime import datetime

from juanfi_api import JuanfiApi
from repository.vendo_repository import VendoRepository
from sql_app.database import SessionLocal
from sqlalchemy.orm import Session
from fastapi import Depends
from sql_app.models import Vendo, VendoStatus

class VendoStatusLog():
    def __init__(
            self,
            db: Session = SessionLocal()
    ):
        self._db = db

    def run(self):
        vendos = self._db.query(Vendo).where(Vendo.is_active == True).all()
        for vendo in vendos:
            try:
                print(vendo.name)
                status = JuanfiApi(vendo).get_system_status()
                self.__save_status(vendo, status)
            except Exception as e:
                logging.warning(repr(e))

    def __save_status(self, vendo, status):
        model = VendoStatus(
            vendo_id=vendo.id,
            total_sales=status.get('total_coin_count'),
            current_sales=status.get('current_coin_count'),
            customer_count=status.get('customer_count'),
            free_heap=status.get('free_heap'),
            wireless_strength=status.get('wireless_signal_strength'),
            active_users=status.get('active_user_count'),
            created_at=datetime.now()
        )
        self._db.add(model)
        self._db.commit()
