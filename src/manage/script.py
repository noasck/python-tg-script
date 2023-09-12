"""Main processing logic."""
from loguru import logger
from pyrogram import Client

from .config import APP_NAME, config
from .storage import handle_results
from .tg_service import extract_ids, fetch_messages, get_public_chats, remove_messages

app = Client(
    APP_NAME,
    api_id=config.api_id,
    api_hash=config.api_hash,
    session_string=config.bot_token,
)


async def process() -> None:
    """Process Telegram Messages for userbot."""
    logger.info("Starting processing messages")
    async with app:
        if config.remove_message_ids and config.remove_chat_id:
            await handle_results(
                await remove_messages(app, config.remove_chat_id, config.remove_message_ids),
                config.persist_path,
            )
            logger.info(
                "Removed messages for chat with ids.",
                msg_ids=config.remove_message_ids,
                chat_id=config.remove_chat_id,
            )
            return

        for chat_id in await get_public_chats(app):
            logger.info("Fetching messages for chat", chat_id=chat_id)
            async for msg_batch in fetch_messages(app, chat_id, config.batch_size):
                if config.remove_all:
                    results = await remove_messages(app, chat_id, extract_ids(msg_batch))
                else:
                    results = msg_batch

                await handle_results(results, config.persist_path)

                logger.info("Fetched a batch of messages", batch_size=len(msg_batch))

    logger.info("Finished processing messages.")


def run() -> None:
    """Run processing messages."""
    app.run(process())
