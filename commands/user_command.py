import typer

from sql_app.database import SessionLocal
from sql_app.models import User


class UserCommand:

    def add_user(self):
        username = input("Enter your username: ")
        password = input("Enter your password: ")

        db = SessionLocal()
        user = User(username=username, password=password, is_active=True)
        db.add(user)
        db.commit()
        db.close()

        typer.echo(f"User {username} was added")