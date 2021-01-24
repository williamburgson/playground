import logging


def get_logger(name=None):
    if name is not None:
        logger = logging.getLogger(name)
    else:
        logger = logging.getLogger()
    logger.setLevel(logging.INFO)
    logging.basicConfig(
        format=(
            "%(asctime)s: %(module)s:%(funcName)s:%(lineno)d"
            " - %(levelname)s - %(message)s"
        )
    )
    return logger
