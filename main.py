from enum import Enum

from fastapi import FastAPI

app = FastAPI()


class ModelName(str, Enum):
    alexnet = "alexnet"
    resnet = "resnet"
    lenet = "lenet"


@app.get("/")
async def root():
    return {"message": "Hello World"}


@app.get("/items/{item_id}")
async def read_item(item_id: int):
    return {"item_id": item_id}


@app.get("/modles/{model_name}")
async def get_model(model_name: ModelName):
    # requires python >= 3.10
    match model_name:
        case ModelName.alexnet:
            return {"model_name": model_name, "message": "Deep Learning FTW!"}
        case ModelName.lenet:
            return {"model_name": model_name, "message": "LeCNN all the images"}
        case _:
            return {"model_name": model_name, "message": "Have some residuals"}
