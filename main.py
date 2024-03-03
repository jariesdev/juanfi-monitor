from typing import Union

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse

from controllers.api.sale_controller import SaleController
from controllers.api.vendo_controller import VendoController
from juanfi_api import JuanfiApi
from juanfi_logger import JuanfiLogger
from log_repository import LogRepository
from models.vendo import VendoMachine
from user_repository import UserRepository

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
    return JSONResponse({"Hello": "World"})


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


@app.get("/logs")
async def read_logs(q: Union[str, None] = None, date: Union[str, None] = None):
    log_repository = LogRepository()
    logs = log_repository.search(q, date)

    def addition(d: list) -> dict:
        return {
            "id": d[0],
            "log_time": d[1],
            "description": d[2],
            "created_at": d[3],
        }

    return JSONResponse({
        "data": list(map(addition, logs))
    })


@app.post("/log/refresh")
async def read_logs():
    juanfi_logger = JuanfiLogger()
    juanfi_logger.run()
    return JSONResponse({
        "data": None,
        "status": "ok"
    })


@app.get("/sales")
def read_sales(q: Union[str, None] = None, date: Union[str, None] = None):
    sale_controller = SaleController()
    return sale_controller.all(q, date)


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
