import argparse
import os
from typing import Tuple, Optional

import aiohttp
from loguru import logger

from pipecat.transports.services.helpers.daily_rest import DailyRESTHelper

# Constants
DEFAULT_API_URL = "https://api.daily.co/v1"
TOKEN_EXPIRY_TIME = 60 * 60  # 1 hour in seconds

async def configure(aiohttp_session: aiohttp.ClientSession) -> Tuple[str, str]:
    """
    Configure the Daily room and authentication settings.
    
    Args:
        aiohttp_session: Active aiohttp client session
        
    Returns:
        Tuple containing (room_url, token)
        
    Raises:
        Exception: If required configuration is missing
    """

    logger.info("🔧 Starting configuration")
    
    # Parse command line arguments
    parser = argparse.ArgumentParser(description="Daily AI SDK Bot Sample")
    parser.add_argument(
        "-u", "--url",
        type=str,
        required=False,
        help="URL of the Daily room to join"
    )
    parser.add_argument(
        "-k", "--apikey",
        type=str,
        required=False,
        help="Daily API Key (needed to create an owner token for the room)"
    )

    args, unknown = parser.parse_known_args()
    
    # Get configuration from args or environment
    url = args.url or os.getenv("DAILY_SAMPLE_ROOM_URL")
    key = args.apikey or os.getenv("DAILY_API_KEY")
    api_url = os.getenv("DAILY_API_URL", DEFAULT_API_URL)

    # Validate URL
    if not url:
        logger.error("❌ No Daily room URL specified")
        raise Exception("No Daily room specified.")
    logger.info(f"✅ Using room URL: {url}")

    # Validate API key
    if not key:
        logger.error("❌ No Daily API key specified")
        raise Exception("No Daily API key specified.")
    logger.debug("✅ API key validated")

    try:
        # Initialize REST helper
        logger.info("🔄 Initializing Daily REST helper")
        daily_rest_helper = DailyRESTHelper(
            daily_api_key=key,
            daily_api_url=api_url,
            aiohttp_session=aiohttp_session,
        )
        
        # Get room token
        logger.info(f"🔑 Requesting room token (expires in {TOKEN_EXPIRY_TIME/60:.0f} minutes)")
        token = await daily_rest_helper.get_token(url, TOKEN_EXPIRY_TIME)
        
        if not token:
            logger.error("❌ Failed to obtain room token")
            raise Exception("Failed to obtain room token")
            
        logger.info("✅ Room token acquired successfully")
        logger.info("✅ Configuration complete")
        
        return (url, token)
        
    except Exception as e:
        logger.error(f"❌ Configuration failed: {str(e)}")
        raise

if __name__ == "__main__":
    logger.info("⚠️ This module should be imported, not run directly")