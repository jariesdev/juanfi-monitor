import time

from fastapi import HTTPException
from fastapi.encoders import jsonable_encoder
from fastapi.responses import JSONResponse
from fastapi.security import OAuth2PasswordRequestForm
from datetime import datetime, timedelta
from repository.user_repository import UserRepository
import base64
import json

class LoginController:
    _repository: UserRepository

    def __init__(self):
        self._repository = UserRepository()

    def login(self, form_data: OAuth2PasswordRequestForm):
        user = self._repository.check_user(form_data.username, form_data.password)
        if not user:
            raise HTTPException(status_code=400, detail="Incorrect username or password")

        # TODO change to actual token
        expiry = datetime.now() + timedelta(hours=1)
        token = base64.b64encode(str({"username": user.username, "expiry": expiry.timestamp()}).encode("ascii")).decode("ascii")
        # base64_bytes = token.encode("ascii")
        # sample_string_bytes = base64.b64decode(base64_bytes)
        # sample_string = sample_string_bytes.decode("ascii")
        # dd = json.loads(sample_string)

        return JSONResponse(status_code=200, content={
            "access_token": token,
            "token_type": "bearer",
            "user": jsonable_encoder(user)
        })

    def get_user_by_token(self, token: str):
        return self._repository.get_user_by_token(token)
