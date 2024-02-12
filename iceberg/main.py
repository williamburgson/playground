import os
import logging
from pyiceberg.catalog import load_catalog

from table import schema, partition_spec, sort_order

catalog = load_catalog(
    "docs",
    **{
        "uri": "http://127.0.0.1:8181",
        "s3.endpoint": "http://192.168.1.238:9000",
        "py-io-imp": "pyiceberg.io.pyarrow.PyArrowFileIO",
        "s3.access-key-id": os.getenv("ACCESS_KEY"),
        "s3.secret-access-key": os.getenv("ACCESS_SECRET"),
    }
)


def main():
    ns = catalog.list_namespaces()
    if "example" not in ns:
        logging.info("initializing new example namespace")
        catalog.create_namespace("example")
    tb = catalog.list_tables("example")
    if "firearm_crime" not in [t[0] for t in tb]:
        logging.info("initializing new iceberg table with example fire arms data csv")
        catalog.create_table(
            identifier="example.firearm_crime",
            schema=schema,
            location="s3://iceberg",
            partition_spec=partition_spec,
            sort_order=sort_order,
        )


if __name__ == "main":
    main()
