"""Fixtures and configurations."""

from .config import logger


@pytest.fixture(autouse=True)
def mocklogger() -> .Cursor:
    """Mock logger to pring to stdout."""
    logger.remove()
    logger.add(sys.stdout)
    return logger

