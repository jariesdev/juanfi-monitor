from typing import Union
from fastapi import FastAPI
from fastapi.responses import JSONResponse
from fastapi.middleware.cors import CORSMiddleware
from juanfi_api import JuanfiApi
from juanfi_logger import JuanfiLogger
from log_repository import LogRepository
from user_repository import UserRepository

app = FastAPI()

origins = [
    "http://localhost",
    "http://localhost:5173",
    "http://192.46.225.21:8001",
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
async def read_logs(q: Union[str, None] = None):
    log_repository = LogRepository()
    logs = log_repository.search(q)

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


@app.get("/vendo_status")
async def read_logs():
    status = JuanfiApi().get_system_status()
    return JSONResponse(status)
