import os
from rds_manager import table_schemas
from rds_manager.logging_util import get_logger


logger = get_logger()


class DataLoader:
    def __init__(self, conn):
        self.conn = conn
        self.cursor = conn.cursor()

    def __execute(self, sql):
        logger.info("Executing: %s", sql)
        self.cursor.execute(sql)
        res = self.conn.commit()
        return res

    def drop_table(self, table: table_schemas.Table):
        logger.info("Dropping table %s", table.name)
        sql = f"DROP TABLE IF EXISTS {table.name} CASCADE;"
        return self.__execute(sql)

    def create_table(self, table: table_schemas.Table):
        self.drop_table(table)
        logger.info("Creating table %s", table.name)
        column_defs = [f"\t\t{c.name} {c.type}," for c in table.schema]
        fk_str = []
        if len(table.fks) > 0:
            for fk in table.fks:
                constrain = f"""
                    FOREIGN KEY({fk.name})
                    REFERENCES {fk.ref_table}({fk.ref_key})
                    ON DELETE SET NULL
                """
                fk_str.append(constrain)
        column_defs = "\n".join(column_defs)
        fk_str = ",\n".join(fk_str)
        sql = f"""
        CREATE TABLE IF NOT EXISTS {table.name} (
        {column_defs}
            PRIMARY KEY ({table.pk.name}){',' if len(fk_str) > 0 else ''}
            {fk_str}
        );
        """
        return self.__execute(sql)

    def load_data(self, table: table_schemas.Table, data_dir, skip_header=1, sep=","):
        table_id = "public." + table.name
        source = os.path.join(data_dir, table.source)
        logger.info("Loading %s to %s", source, table_id)
        cols = [c.name for c in table.schema]
        with open(source, "r") as fr:
            [fr.readline() for i in range(skip_header)]
            res = self.cursor.copy_from(fr, table_id, sep=sep, columns=cols)
            self.conn.commit()
        return res
