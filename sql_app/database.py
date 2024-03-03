from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
import logging

SQLALCHEMY_DATABASE_URL = "sqlite:///./app.db"

logging.basicConfig()
logging.getLogger("sqlalchemy.engine").setLevel(logging.INFO)

engine = create_engine(
    SQLALCHEMY_DATABASE_URL, connect_args={"check_same_thread": False}
)
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

