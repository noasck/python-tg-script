"""Config parser. Set up logger."""
import os
import threading

from loguru import logger
from pydantic import ValidationError
from pydantic.dataclasses import ConfigDict, dataclass

APP_NAME = "tg_manage"
_APP_ENV_PREFIX = APP_NAME.upper() + "_"


@dataclass(
    config=ConfigDict(
        extra="ignore",
        alias_generator=lambda x: x.upper(),
    ),
)
class ConfigParser:

    """Pass lowercased envs names below."""

    bot_token: str
    api_id: int
    api_hash: str


class _Config:

    """Thread safe singleton."""

    _instance = None
    _lock = threading.Lock()

    def __new__(cls) -> "_Config":
        """Aquire lock and create instance per thread."""
        if cls._instance is None:
            with cls._lock:
                if not cls._instance:
                    cls._instance = super().__new__(cls)
                    cls._instance.__init()
        return cls._instance

    def __init(self) -> None:
        """Parse configuration."""
        try:
            # Reading environmental variables
            # with APP_NAME + _ prefix.
            # Passing to serializer without prefix.
            self.config = ConfigParser(
                **{
                    env.replace(_APP_ENV_PREFIX, ""): value
                    for env, value in os.environ.items()
                    if env.startswith(APP_NAME.upper())
                },
            )
        except ValidationError as exc:
            for error in exc.errors():
                logger.error(
                    "Invalid config: {}. Reason: {}.",
                    error.get("loc", ["NA"])[0],
                    error.get("msg", "NA"),
                )
            raise
        except BaseException:
            logger.exception("Unknown exception during config parsing.")
            raise

        logger.add(
            f"logs/{APP_NAME}_{{time}}.log",
            compression="zip",
            rotation="12:00",
            level="INFO",
            serialize=True,
            backtrace=False,
            diagnose=False,
            format="",
        )

        logger.info("Logger set up successfully")
        logger.info("Configuration read successfully.")


config = _Config().config
