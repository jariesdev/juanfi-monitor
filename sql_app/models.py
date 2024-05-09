from sqlalchemy import *
from sqlalchemy import orm
from sqlalchemy.orm import *
from sqlalchemy.ext.declarative import declarative_base

Base = declarative_base()


class User(Base):
    __tablename__ = "users"

    id = Column(Integer, primary_key=True)
    username = Column(String, unique=True, nullable=False)
    password = deferred(Column(String, nullable=False))
    is_active = Column(Boolean, default=True)
    created_at = Column(TIMESTAMP, server_default=func.now(), nullable=False)
    updated_at = Column(TIMESTAMP)


class VendoStatus(Base):
    __tablename__ = "vendo_status"

    id = Column(Integer, primary_key=True)
    vendo_id = Column(Integer, ForeignKey("vendos.id"))
    vendo = relationship("Vendo", back_populates="vendo_status")
    total_sales = Column(Float)
    current_sales = Column(Float)
    customer_count = Column(Integer)
    free_heap = Column(Integer)
    wireless_strength = Column(Float)
    active_users = Column(Integer)
    created_at = Column(TIMESTAMP, server_default=func.now(), nullable=False)
    updated_at = Column(TIMESTAMP)


class Vendo(Base):
    __tablename__ = "vendos"

    id = Column(Integer, primary_key=True)
    name = Column(String, unique=True, nullable=False)
    mac_address = Column(String)
    api_url = Column(String)
    api_key = deferred(Column(String))
    is_online = Column(Boolean, server_default="1")
    total_sales = Column(Float)
    current_sales = Column(Float)
    is_active = Column(Integer, default=1)
    created_at = Column(TIMESTAMP, server_default=func.now(), nullable=False)
    updated_at = Column(TIMESTAMP)
    vendo_logs = relationship("VendoLog", back_populates='vendo')
    vendo_sales = relationship("VendoSale", back_populates='vendo')
    vendo_status = relationship("VendoStatus", back_populates='vendo')
    withdrawals = relationship("Withdrawal", back_populates='vendo')

    @orm.declared_attr
    def __mapper_args__(vendo):
        children = VendoStatus
        most_recent_child = (
            select(children.id)
            .where(children.vendo_id == vendo.id)
            .order_by(VendoStatus.id.desc())
            .limit(1)
            .correlate(vendo.__table__)
            .scalar_subquery()
        )

        rel = orm.relationship(
            VendoStatus,
            primaryjoin=and_(
                VendoStatus.id == most_recent_child, VendoStatus.vendo_id == vendo.id
            ),
            uselist=False,
            viewonly=True,
        )
        return {'properties': {'recent_status': rel}}


class VendoLog(Base):
    __tablename__ = "vendo_logs"

    id = Column(Integer, primary_key=True)
    vendo_id = Column(Integer, ForeignKey("vendos.id"))
    log_time = Column(DateTime)
    description = Column(String)
    vendo = relationship("Vendo", back_populates='vendo_logs')
    created_at = Column(TIMESTAMP, server_default=func.now(), nullable=False)
    updated_at = Column(TIMESTAMP)
    UniqueConstraint("vendo_id", "log_time")


class VendoSale(Base):
    __tablename__ = "vendo_sales"

    id = Column(Integer, primary_key=True)
    vendo_id = Column(Integer, ForeignKey("vendos.id"))
    vendo = relationship("Vendo", back_populates="vendo_sales")
    sale_time = Column(DateTime)
    mac_address = Column(String)
    voucher = Column(String)
    amount = Column(Float)
    created_at = Column(TIMESTAMP, server_default=func.now(), nullable=False)
    updated_at = Column(TIMESTAMP)
    UniqueConstraint("vendo_id")


class Withdrawal(Base):
    __tablename__ = "withdrawals"

    id = Column(Integer, primary_key=True)
    vendo_id = Column(Integer, ForeignKey("vendos.id"))
    vendo = relationship("Vendo", back_populates="withdrawals")
    user_id = Column(Integer, ForeignKey("users.id"))
    amount = Column(Float)
    created_at = Column(TIMESTAMP, server_default=func.now(), nullable=False)
    updated_at = Column(TIMESTAMP)


