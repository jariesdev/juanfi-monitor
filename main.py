import os
import sys
from typing import Union

from dotenv import load_dotenv
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse

import sql_app
from controllers.api.log_controller import LogController
from controllers.api.sale_controller import SaleController
from controllers.api.vendo_controller import VendoController
from juanfi_api import JuanfiApi
from juanfi_logger import JuanfiLogger
from models.vendo import VendoMachine
from sql_app.database import SessionLocal
from sql_app.schemas import VendoLogResponse, VendoSaleResponse
from user_repository import UserRepository

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
load_dotenv(os.path.join(BASE_DIR, ".env"))
sys.path.append(BASE_DIR)

app = FastAPI()

origins = [
    "http://localhost",
    "http://localhost:5173",
    "http://localhost:4173",
    "http://localhost:8001",
    "http://192.46.225.21:8001",
    "*.jaries.dev",
]

app.add_middleware(
    CORSMiddleware,
    allow_origins=origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/")
def read_root():
    return "Ok"


@app.get("/users")
async def read_users(q: Union[str, None] = None):
    user_repository = UserRepository()
    users = user_repository.search(q)
    return JSONResponse({
        "data": users
    })


@app.get("/users/{user_id}")
def read_user(user_id: int, q: Union[str, None] = None):
    return JSONResponse({"item_id": user_id, "q": q})


@app.get("/logs", response_model=VendoLogResponse)
async def read_logs(q: Union[str, None] = None, date: Union[str, None] = None, vendo_id: Union[int, None] = None):
    controller = LogController()
    response = controller.search(q, date, vendo_id)
    return response


@app.post("/log/refresh")
async def read_logs():
    db = SessionLocal()
    vendos = db.query(sql_app.models.Vendo).all()
    if len(vendos) == 0:
        return JSONResponse({
            "data": None,
            "message": "No registered vendo. Please add first."
        })

    for vendo in vendos:
        try:
            logger = JuanfiLogger(vendo)
            logger.run()
        finally:
            pass

    return JSONResponse({
        "data": None,
        "status": "ok"
    })


@app.get("/sales", response_model=VendoSaleResponse)
def read_sales(q: Union[str, None] = None, date: Union[str, None] = None, vendo_id: Union[int, None] = None):
    sale_controller = SaleController()
    return sale_controller.search(q, date, vendo_id)


@app.get("/daily-sales")
def read_daily_sales():
    sale_controller = SaleController()
    return sale_controller.daily_sales()


@app.get("/vendo_status")
async def read_logs():
    status = JuanfiApi().get_system_status()
    return JSONResponse(status)


@app.get("/vendo-machines")
async def read_logs(q: Union[str, None] = None):
    controller = VendoController()
    return controller.all(q)


@app.get("/vendo-machines/{id}/status")
async def read_vendo_status(id: int):
    controller = VendoController()
    return controller.vendo_status(id)


@app.post("/vendo-machines")
async def read_logs(vendo: VendoMachine):
    sale_controller = VendoController()
    return sale_controller.store(vendo)


@app.delete("/vendo-machines/{id}")
async def read_logs(id: int):
    sale_controller = VendoController()
    return sale_controller.delete(id)
