"""Fixtures and configurations."""
import sys

import pytest

from .config import logger


@pytest.fixture(autouse=True)
def mocklogger():  # noqa: ANN201
    """Mock logger to pring to stdout."""
    logger.remove()
    logger.add(sys.stdout)
    return logger
