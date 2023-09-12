"""Telegram Management Service."""

from collections.abc import AsyncGenerator

from loguru import logger
from pyrogram import Client
from pyrogram.enums import ChatType
from pyrogram.errors import RPCError

from .helper_types import MsgStatus, Result


async def remove_messages(app: Client, chat_id: int, message_ids: list[int]) -> list[Result]:
    """Remove messages by list of ids."""
    try:
        removed = await app.delete_messages(chat_id, message_ids)
    except RPCError:
        logger.exception(
            "Exception during removing messages",
            chat_id=chat_id,
            msg_ids=message_ids,
        )
        removed = 0

    logger.info("Removed messages for chat.", chat_id=chat_id, msg_count=removed)
    status = MsgStatus.error.value if removed != len(message_ids) else MsgStatus.removed.value
    return [Result(msg_id, chat_id, None, None, status) for msg_id in message_ids]


async def get_public_chats(app: Client) -> list[int]:
    """Get public chat ids."""
    return [
        dial.chat.id async for dial in app.get_dialogs() if dial.chat.type == ChatType.SUPERGROUP
    ]


def extract_ids(msgs: list[Result]) -> list[int]:
    """Return only ids of fetched messages."""
    return [res.id for res in msgs]


async def fetch_messages(
    app: Client,
    chat_id: int,
    batch_size: int = 100,
) -> AsyncGenerator[list[Result], None]:
    """Fetch messages from chat batched."""
    results = []
    try:
        async for message in app.get_chat_history(chat_id):
            if len(results) == batch_size:
                yield results
                results = []

            results.append(
                Result(
                    message.id,
                    chat_id,
                    message.text,
                    message.date,
                    MsgStatus.fetched.value,
                ),
            )
        yield results
    except RPCError:
        logger.exception("Failed to fetch messages", chat_id=chat_id)
        return
