from typing import Union
from fastapi import FastAPI
from fastapi.responses import JSONResponse

from juanfi_api import JuanfiApi
from log_repository import LogRepository
from user_repository import UserRepository

app = FastAPI()


@app.get("/")
def read_root():
    return JSONResponse({"Hello": "World"})


@app.get("/users")
async def read_users(q: Union[str, None] = None):
    user_repository = UserRepository()
    users = user_repository.search(q)
    return JSONResponse({
        "users": users
    })


@app.get("/users/{user_id}")
def read_user(user_id: int, q: Union[str, None] = None):
    return JSONResponse({"item_id": user_id, "q": q})


@app.get("/logs")
async def read_logs(q: Union[str, None] = None):
    log_repository = LogRepository()
    logs = log_repository.search(q)
    return JSONResponse({
        "logs": logs
    })

@app.get("/vendo_status")
async def read_logs():
    status = JuanfiApi().get_system_status()
    return JSONResponse(status)
