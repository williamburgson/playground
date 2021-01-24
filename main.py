from rds_manager import rds_manager, data_loader, table_schemas

DATA_DIR = "data"

conn, cursor = rds_manager.connect()
dl = data_loader.DataLoader(conn)

users: table_schemas.Table = table_schemas.Users
dl.create_table(users)
dl.load_data(data_dir=DATA_DIR, table=users)

trips: table_schemas.Table = table_schemas.Trips
dl.create_table(trips)
dl.load_data(data_dir=DATA_DIR, table=trips)
