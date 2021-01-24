from dataclasses import dataclass
import typing


@dataclass
class ForeignKey:
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
        ForeignKey("Client_Id", Users.name, "Users_Id"),
        ForeignKey("Driver_Id", Users.name, "Users_Id"),
    ]
    schema = [
        Column("Id", "INT"),
        Column("Client_Id", "INT"),
        Column("Driver_Id", "INT"),
        Column("City_ID", "INT"),
        Column("Status", "VARCHAR"),
        Column("Request_at", "DATE"),
    ]
