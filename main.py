from rds_manager import rds_manager, data_loader, table_schemas

DATA_DIR = "data"

conn, cursor = rds_manager.connect()
dl = data_loader.DataLoader(conn)

for table in [table_schemas.Employee, table_schemas.Salary, table_schemas.Activity]:
    dl.create_table(table)
    dl.load_data(data_dir=DATA_DIR, table=table)
