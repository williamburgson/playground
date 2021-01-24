from rds_manager import rds_manager, data_loader, table_schemas

DATA_DIR = "data"

conn, cursor = rds_manager.connect()
dl = data_loader.DataLoader(cursor)
users: table_schemas.Table = table_schemas.Users
dl.create_table(users)
# dl.load_data(data_dir=DATA_DIR, table=users)
