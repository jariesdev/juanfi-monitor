from sqlalchemy import Boolean, Column, ForeignKey, Integer, String, Float, DateTime, UniqueConstraint, TIMESTAMP, func
from sqlalchemy.orm import relationship
from sqlalchemy.ext.declarative import declarative_base
import datetime

Base = declarative_base()


class User(Base):
    __tablename__ = "users"

    id = Column(Integer, primary_key=True)
    username = Column(String, unique=True, nullable=False)
    password = Column(String, nullable=False)
    is_active = Column(Boolean, default=True)
    created_at = Column(TIMESTAMP, server_default=func.now(), nullable=False)
    updated_at = Column(TIMESTAMP)


class Vendo(Base):
    __tablename__ = "vendos"

    id = Column(Integer, primary_key=True)
    name = Column(String, unique=True, nullable=False)
    mac_address = Column(String)
    api_url = Column(String)
    api_key = Column(String)
    is_online = Column(Boolean, default=False)
    total_sales = Column(Float)
    current_sales = Column(Float)
    created_at = Column(TIMESTAMP, server_default=func.now(), nullable=False)
    updated_at = Column(TIMESTAMP)


class VendoLog(Base):
    __tablename__ = "vendo_logs"

    id = Column(Integer, primary_key=True)
    vendo_id = Column(Integer, ForeignKey("vendos.id"))
    log_time = Column(DateTime)
    description = Column(String)
    vendo = relationship(Vendo, back_populates="vendos")
    created_at = Column(TIMESTAMP, server_default=func.now(), nullable=False)
    updated_at = Column(TIMESTAMP)
    UniqueConstraint("log_time")


class VendoLog(Base):
    __tablename__ = "vendo_sales"

    id = Column(Integer, primary_key=True)
    vendo_id = Column(Integer, ForeignKey("vendos.id"))
    vendo = relationship(Vendo, back_populates="vendos")
    sale_time = Column(DateTime)
    mac_address = Column(String)
    voucher = Column(String)
    amount = Column(Float)
    created_at = Column(TIMESTAMP, server_default=func.now(), nullable=False)
    updated_at = Column(TIMESTAMP)
    UniqueConstraint("sale_time")


class VendoStatus(Base):
    __tablename__ = "vendo_status"

    id = Column(Integer, primary_key=True)
    vendo_id = Column(Integer, ForeignKey("vendos.id"))
    vendo = relationship(Vendo, back_populates="vendos")
    is_online = Column(Boolean, default=False)
    active_users = Column(Integer)
    created_at = Column(TIMESTAMP, server_default=func.now(), nullable=False)
    updated_at = Column(TIMESTAMP)
