from dataclasses import dataclass
import os

import boto3
import psycopg2
from tenacity import retry, retry_if_exception_type, stop_after_delay, wait_fixed
from rds_manager.logging_util import get_logger


# One instance only, don't have the extra $$
PGHOST = "postgres-db.carrdl6ran45.us-east-1.rds.amazonaws.com"
PGINSTANCE = "postgres-db"
PGDATABASE = "postgres"
PGPORT = "5432"
# IDEALLY, these would live in a secret manager
# but I don't want to pay for it, so meh
PGUSER = "willwang"
PGPASSWORD = os.getenv("PGPASSWORD")

# use the base logger s.t. it will log properly in the notbooks
logger = get_logger()

rds_client = boto3.client("rds", region_name="us-east-1")


class DBStateException(Exception):
    pass


@dataclass
class DBState:
    AVAILABLE = "available"
    STOPPED = "stopped"
    STOPPING = "stopping"
    STARTING = "starting"


def __update_instance(fn):
    res = fn(DBInstanceIdentifier=PGINSTANCE)
    logger.debug(res)
    return res


@retry(
    stop=stop_after_delay(120),
    reraise=True,
    wait=wait_fixed(45),
    retry=retry_if_exception_type(DBStateException),
    before_sleep=lambda state: logger.info(
        "Waiting for instance update to compelete (%d) ...",
        int(state.attempt_number) * 45,
    ),
)
def __instance_ready(state):
    """Keep checking the instance status until it is ready"""
    response = __update_instance(rds_client.describe_db_instances)
    status = response["DBInstances"][0]["DBInstanceStatus"]
    if status == state:
        return True
    raise DBStateException


def start_instance():
    try:
        __update_instance(rds_client.start_db_instance)
        logger.info("RDS Instance %s is being started", PGINSTANCE)
        __instance_ready(DBState.AVAILABLE)
    except rds_client.exceptions.InvalidDBInstanceStateFault:
        logger.info("RDS Instance %s is already started", PGINSTANCE)


def stop_instance():
    try:
        __update_instance(rds_client.stop_db_instance)
        logger.info("RDS Instance %s is being stopped", PGINSTANCE)
        __instance_ready(DBState.STOPPED)
    except rds_client.exceptions.InvalidDBInstanceStateFault:
        logger.info("RDS Instance %s is already stopped", PGINSTANCE)
        raise DBStateException


def __connect_db(table_name):
    conn = psycopg2.connect(
        host=PGHOST, port=PGPORT, dbname=table_name, user=PGUSER, password=PGPASSWORD
    )
    logger.info("Connection established to %s at %s", PGDATABASE, PGHOST)
    cursor = conn.cursor()
    return conn, cursor


def connect(table_name=PGDATABASE):
    try:
        logger.info("Trying to establish connection to %s", PGDATABASE)
        return __connect_db(table_name)
    except Exception:
        logger.info("Cannot connect, try restarting db instance")
        try:
            stop_instance()  # anything other than state fault will break this
        except DBStateException:
            try:
                start_instance()
                return __connect_db(table_name)
            except Exception as e:
                logger.error("Cannot start db %s", PGDATABASE)
                logger.exception(e)
    return None, None
