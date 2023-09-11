"""Log into Telegram API."""
import datetime

from config import config
from loguru import logger
from pyrogram import Client
from pyrogram.errors import RPCError

app = Client(
    f"userbot_{datetime.datetime.now(tz=datetime.UTC).date()}",
    workdir="sessions/",
    api_id=config.api_id,
    api_hash=config.api_hash,
)


async def login() -> None:
    """Login into Telegram to create session."""
    try:
        async with app:
            logger.info("Please, follow the instructions to authorize and create session.")
            await app.send_message("me", "New login into the Account.")
            logger.info("You could find a session in ./sessions folder.")
    except RPCError:
        logger.exception("Couldn't log in.")


if __name__ == "__main__":
    app.run(login())
