import os
import sys
import time
from contextlib import asynccontextmanager
from typing import Union, Annotated

import uvicorn
from apscheduler.schedulers.background import BackgroundScheduler
from dotenv import load_dotenv
from fastapi import FastAPI, Depends, HTTPException, WebSocket, WebSocketDisconnect
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse, HTMLResponse
from fastapi.security import OAuth2PasswordBearer, OAuth2PasswordRequestForm

import sql_app
from commands.vendo_status_log import VendoStatusLog
from controllers.api.log_controller import LogController
from controllers.api.login_controller import LoginController
from controllers.api.sale_controller import SaleController
from controllers.api.vendo_controller import VendoController
from controllers.api.vendo_status_controller import VendoStatusController
from controllers.api.withdrawal_controller import WithdrawalController
from juanfi_logger import JuanfiLogger
from models.vendo import VendoMachine
from repository.notification_repository import NotificationRepository
from sql_app.database import SessionLocal
from sql_app.schemas import VendoLogResponse, VendoSaleResponse, User, SalesSearchRequest, LogsSearchRequest, \
    DailySaleRequest, SetVendoStatusRequest
from sql_app.models import VendoSale
from src.vendoreport.notification.notificatoin_manager import ConnectionManager
from user_repository import UserRepository
from fastapi_pagination import Page, add_pagination

BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
load_dotenv(os.path.join(BASE_DIR, ".env"))
sys.path.append(BASE_DIR)


@asynccontextmanager
async def lifespan(app: FastAPI):
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


@app.get("/logs", response_model=Page[sql_app.schemas.VendoLog])
async def read_logs(
        request: LogsSearchRequest = Depends(),
        controller: LogController = Depends(LogController)
):
    response = controller.search(request)
    return response


@app.post("/log/refresh")
async def refresh_logs():
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


@app.get("/sales", response_model=Page[sql_app.schemas.VendoSale])
def read_sales(request: SalesSearchRequest = Depends(), controller: SaleController = Depends(SaleController)):
    return controller.search(request)


@app.get("/daily-sales")
def read_daily_sales(request: DailySaleRequest = Depends(), controller: SaleController = Depends(SaleController)):
    return controller.daily_sales(request)


@app.get("/vendo-status-history")
async def read_vendo_status(
        controller: VendoStatusController = Depends(VendoStatusController),
        vendo_id: Union[int, None] = None,
        from_date: Union[str, None] = None,
        to_date: Union[str, None] = None,
        active_only: bool | None = False,
):
    return controller.search(vendo_id=vendo_id, from_date=from_date, to_date=to_date, active_only=active_only)


@app.get("/vendo-machines")
async def read_vendos(controller: VendoController = Depends(VendoController), q: Union[str, None] = None, is_active: bool | None = None):
    return controller.all(q, is_active)


@app.get("/vendo-machines/{vendo_id}/status")
async def read_vendo_status2(vendo_id: int, controller: VendoController = Depends(VendoController)):
    return controller.vendo_status(vendo_id)


@app.post("/vendo-machines")
async def store_vendo(vendo: VendoMachine, controller: VendoController = Depends(VendoController)):
    return controller.store(vendo)


@app.get("/vendo-machines/{vendo_id}")
async def delete_vendo(vendo_id: int, controller: VendoController = Depends(VendoController)):
    return controller.get(vendo_id)


@app.delete("/vendo-machines/{vendo_id}")
async def delete_vendo(vendo_id: int, controller: VendoController = Depends(VendoController)):
    return controller.delete(vendo_id)


@app.post("/vendo-machines/{vendo_id}/withdraw-current-sales")
async def read_vendo_sales(vendo_id: int, controller: VendoController = Depends(VendoController)):
    return controller.withdraw_current_sales(vendo_id)


@app.post("/vendo-machines/{vendo_id}/set-status")
async def set_vendo_status(vendo_id: int, request: SetVendoStatusRequest, controller: VendoController = Depends(VendoController)):
    return controller.set_status(vendo_id, request)


@app.get("/withdrawals")
async def read_withdrawals(controller: WithdrawalController = Depends(WithdrawalController)):
    return controller.search()


manager = ConnectionManager()

@app.websocket("/ws")
async def websocket_endpoint(websocket: WebSocket, repository: NotificationRepository = Depends(NotificationRepository)):
    await manager.connect(websocket)
    try:
        while True:
            unread = repository.pull_unread()
            # You can handle incoming messages from clients here if needed
            # data = await websocket.receive_text()
            # Example: Echoing back received message
            for notification in unread:
                await manager.broadcast(f"{notification.message}")

            time.sleep(10)
    except WebSocketDisconnect:
        manager.disconnect(websocket)
        await manager.broadcast(f"Client disconnected")

add_pagination(app)

if __name__ == "__main__":
    uvicorn.run(app, host="127.0.0.1", port=8000)
