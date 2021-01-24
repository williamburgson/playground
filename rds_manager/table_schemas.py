from dataclasses import dataclass, field
import typing


@dataclass
class ForeignKey:
    constrain: str
    name: str
    ref_table: str
    ref_key: str


@dataclass
class Column:
    name: str
    type: str


@dataclass
class Schema:
    columns: typing.List[Column]


@dataclass
class Table:
    schema: Schema
    name: str
    source: str
    pk: Column
    fks: typing.List[ForeignKey]


class Users(Table):
    name = "Users"
    source = "users.csv"
    pk = Column("Users_Id", "INT")
    fks = []
    schema = [
        Column("Users_Id", "INT"),
        Column("Banned", "VARCHAR"),
        Column("Role", "VARCHAR"),
    ]


class Trips(Table):
    name = "Trips"
    source = "trips.csv"
    pk = Column("Id", "INT")
    fks = [
        ForeignKey("fk_cid", "Client_Id", Users.name, "User_Id"),
        ForeignKey("fk_did", "Driver_Id", Users.name, "User_Id"),
    ]
    schema = [
        Column("Id", "INT"),
        Column("Client_Id", "INT"),
        Column("Driver_Id", "INT"),
        Column("City_ID", "INT"),
        Column("Status", "INT"),
        Column("Request_at", "DATE"),
    ]
