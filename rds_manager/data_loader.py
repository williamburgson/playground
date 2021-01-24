import os
from rds_manager import table_schemas


class DataLoader:
    def __init__(self, cursor):
        self.cursor = cursor

    def __execute(self, sql):
        print(sql)
        res = self.cursor.execute
        print(res)
        return res

    def drop_table(self, table: table_schemas.Table):
        sql = f"DROP TABLE IF EXISTS {table.name};"
        return self.__execute(sql)

    def create_table(self, table: table_schemas.Table):
        self.drop_table(table)
        column_defs = [f"\t\t{c.name} {c.type}," for c in table.schema]
        fk_str = []
        if len(table.fks) > 0:
            fk_str.append(",")
            for fk in table.fks:
                constrain = f"""
                CONSTRAIN {fk.constrain}
                    FOREIGN KEY({fk.name})
                    REFEREMCES {fk.ref_table}({fk.ref_key})
                    ON DELETE SET NULL\n
                """
                fk_str.append(constrain)
        column_defs = "\n".join(column_defs)
        fk_str = "\n".join(fk_str)
        sql = f"""
        CREATE TABLE {table.name} (
        {column_defs}
            PRIMARY KEY ({table.pk.name})
            {fk_str}
        );
        """
        return self.__execute(sql)

    def load_data(self, table: table_schemas.Table, data_dir, skip_header=1, sep=","):
        table_id = "public." + table.name
        source = os.path.join(data_dir, table.source)
        cols = [c.name for c in table.schema]
        with open(source, "r") as fr:
            [fr.readline() for i in range(skip_header)]
            res = self.cursor.copy_from(fr, table_id, sep=sep, columns=cols)
        return res
