"""Storage Options. Default to save to a file."""
import csv
import os
import aiofiles
from loguru import logger

from .helper_types import Result


async def handle_results(row_data: list[Result], filename: str) -> None:
    """
    Write data to a CSV file asynchronously.

    If the file exists, it appends the data as a row.
    If the file doesn't exist, it creates the file and writes
    the header followed by the data as a row.

    Args:
        row_data (tuple): The data to write as a row.
        filename (str): The name of the CSV file.

    """
    if not filename:
        logger.warning("No persistance set up.")

    # Check if the CSV file exists
    file_exists = os.path.exists(filename)

    async with aiofiles.open(filename, mode="a+", newline="") as file:
        writer = csv.writer(file)

        # If the file doesn't exist, write the header first
        if not file_exists:
            await writer.writerow(list(Result._fields))

        # Write the row data
        await writer.writerow(row_data)
