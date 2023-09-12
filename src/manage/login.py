"""Log into Telegram API."""
import datetime

from config import config
from loguru import logger
from pyrogram import Client
from pyrogram.errors import RPCError

app = Client(
    "login_bot",
    in_memory=True,
    api_id=config.api_id,
    api_hash=config.api_hash,
)


async def login() -> None:
    """Login into Telegram to create session."""
    try:
        async with app:
            logger.info("Please, follow the instructions to authorize and create session.")
            await app.send_message("me", "New login into the Account.")

            try:
                with open(  # noqa: ASYNC101, PTH123
                    f"./sessions/userbot_{datetime.datetime.now(tz=datetime.UTC).date()}",
                    "w",
                ) as file:
                    file.write(await app.export_session_string())
            except Exception:  # noqa: BLE001
                logger.exception("New session could not be created.")
            logger.info("You could find a session in ./sessions folder.")

    except RPCError:
        logger.exception("Couldn't log in.")


if __name__ == "__main__":
    app.run(login())
