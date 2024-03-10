from contextlib import asynccontextmanager

import uvicorn
import os
import sys
from typing import Union, Annotated

from dotenv import load_dotenv
from fastapi import FastAPI, Depends, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from fastapi.security import OAuth2PasswordBearer, OAuth2PasswordRequestForm

import sql_app
from commands.vendo_status_log import VendoStatusLog
from controllers.api.log_controller import LogController
from controllers.api.login_controller import LoginController
from controllers.api.sale_controller import SaleController
from controllers.api.vendo_controller import VendoController
from controllers.api.withdrawal_controller import WithdrawalController
from juanfi_api import JuanfiApi
from juanfi_logger import JuanfiLogger
from models.vendo import VendoMachine
from sql_app.database import SessionLocal
from sql_app.schemas import VendoLogResponse, VendoSaleResponse, User
from user_repository import UserRepository
from apscheduler.schedulers.background import BackgroundScheduler

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
load_dotenv(os.path.join(BASE_DIR, ".env"))
sys.path.append(BASE_DIR)


@asynccontextmanager
async def lifespan(app:FastAPI):
    scheduler = BackgroundScheduler()
    scheduler.add_job(VendoStatusLog().run, "interval", minutes=10)
    scheduler.start()
    yield


app = FastAPI(lifespan=lifespan)

oauth2_scheme = OAuth2PasswordBearer(tokenUrl="token")

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
    return JSONResponse(content={"message": "Ok"})


@app.post("/token")
async def login(
        form_data: Annotated[OAuth2PasswordRequestForm, Depends()],
        controller: LoginController = Depends(LoginController)
):
    return controller.login(form_data)


async def get_current_user(
        token: Annotated[str, Depends(oauth2_scheme)],
        controller: LoginController = Depends(LoginController)
):
    user = controller.get_user_by_token(token)
    return user


async def get_current_active_user(
        current_user: Annotated[User, Depends(get_current_user)]
):
    if not current_user.is_active:
        raise HTTPException(status_code=400, detail="Inactive user")
    return current_user


def fake_decode_token(token):
    return User(
        username=token + "fakedecoded", password="1234", is_active=True
    )


@app.get("/users/me")
async def read_users_me(current_user: Annotated[User, Depends(get_current_active_user)]):
    return current_user


@app.get("/users")
async def read_users(token: Annotated[str, Depends(oauth2_scheme)], q: Union[str, None] = None):
    user_repository = UserRepository()
    users = user_repository.search(q)
    return JSONResponse({
        "data": users,
        "token": token
    })


@app.get("/users/{user_id}")
def read_user(user_id: int, q: Union[str, None] = None):
    return JSONResponse({"item_id": user_id, "q": q})


@app.get("/logs", response_model=VendoLogResponse)
async def read_logs(
        controller: LogController = Depends(LogController),
        q: Union[str, None] = None,
        date: Union[str, None] = None,
        vendo_id: Union[int, None] = None
):
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
def read_sales(controller: SaleController = Depends(SaleController), q: Union[str, None] = None,
               date: Union[str, None] = None, vendo_id: Union[int, None] = None):
    return controller.search(q, date, vendo_id)
    return sale_controller.search(q, date, vendo_id)


@app.get("/daily-sales")
def read_daily_sales(controller: SaleController = Depends(SaleController)):
    return controller.daily_sales()


@app.get("/vendo_status")
async def read_logs():
    status = JuanfiApi().get_system_status()
    return JSONResponse(status)


@app.get("/vendo-machines")
async def read_logs(controller: VendoController = Depends(VendoController), q: Union[str, None] = None):
    return controller.all(q)


@app.get("/vendo-machines/{vendo_id}/status")
async def read_vendo_status(vendo_id: int, controller: VendoController = Depends(VendoController)):
    return controller.vendo_status(vendo_id)


@app.post("/vendo-machines")
async def read_logs(vendo: VendoMachine, controller: VendoController = Depends(VendoController)):
    return controller.store(vendo)


@app.delete("/vendo-machines/{vendo_id}")
async def read_logs(vendo_id: int, controller: VendoController = Depends(VendoController)):
    return controller.delete(vendo_id)


@app.post("/vendo-machines/{vendo_id}/withdraw-current-sales")
async def read_logs(vendo_id: int, controller: VendoController = Depends(VendoController)):
    return controller.withdraw_current_sales(vendo_id)


@app.get("/withdrawals")
async def read_logs(controller: WithdrawalController = Depends(WithdrawalController)):
    return controller.search()


if __name__ == "__main__":
    uvicorn.run(app, host="127.0.0.1", port=8000)
