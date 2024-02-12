from celery import Celery
from celery.utils.log import get_task_logger

logger = get_task_logger(__name__)

celery_app = Celery(
    __name__,
    # broker=f"redis://{os.getenv('REDIS_HOST')}:{os.getenv('REDIS_PORT')}/{os.getenv('REDIS_DB')}",
    # backend=f"redis://{os.getenv('REDIS_HOST')}:{os.getenv('REDIS_PORT')}/{os.getenv('REDIS_DB')}"
)


@celery_app.task
def echo(msg: str):
    logger.info(f"executing echo task, message: {msg}")
    return


if __name__ == "__main__":
    args = ["worker", "--loglevel=INFO"]
    celery_app.worker_main(argv=args)
