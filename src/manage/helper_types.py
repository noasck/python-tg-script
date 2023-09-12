"""Helper types."""

from collections import namedtuple
from enum import Enum


class MsgStatus(Enum):
    """Message frocessing status."""

    fetched = "FETCHED"
    error = "ERROR"
    removed = "REMOVED"


Result = namedtuple("Result", ["id", "chat_id", "text", "date", "status"])
