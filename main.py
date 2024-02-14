from typing import Union
from fastapi import FastAPI
import sqlite3

from log_repository import LogRepository
from user_repository import UserRepository

app = FastAPI()


@app.get("/")
def read_root():
    return {"Hello": "World"}


@app.get("/users")
async def read_users(q: Union[str, None] = None):
    user_repository = UserRepository()
    users = user_repository.search(q)
    return {
        "users": users
    }


@app.get("/users/{user_id}")
def read_user(user_id: int, q: Union[str, None] = None):
    return {"item_id": user_id, "q": q}


@app.get("/logs")
async def read_logs(q: Union[str, None] = None):
    log_repository = LogRepository()
    logs = log_repository.search(q)
    return {
        "logs": logs
    }
