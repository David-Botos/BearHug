import argparse
import os
import subprocess
import ssl
import certifi
from contextlib import asynccontextmanager
from typing import Dict, Tuple, Optional
import aiohttp
import asyncio

from dotenv import load_dotenv
from fastapi import FastAPI, HTTPException, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse, RedirectResponse, StreamingResponse
from loguru import logger
import sys

from pipecat.transports.services.helpers.daily_rest import DailyRESTHelper, DailyRoomParams, DailyRoomProperties, DailyRoomSipParams, DailyRoomObject

def configure_logging():
    if not logger._core.handlers:
        logger.add(
            sys.stderr,
            format="<green>{time:YYYY-MM-DD HH:mm:ss}</green> | <level>{level}</level> | {message}",
            level="DEBUG",
        )

# Constants
MAX_BOTS_PER_ROOM = 1
DEFAULT_HOST = os.getenv("HOST", "127.0.0.1")
DEFAULT_PORT = int(os.getenv("FAST_API_PORT", "7860"))
DAILY_API_URL = os.getenv("DAILY_API_URL", "https://api.daily.co/v1")
TOKEN_EXPIRY_TIME = 60 * 60  # 1 hour in seconds

# Type definitions
BotProcess = Tuple[subprocess.Popen, str]  # (process, room_url)
bot_procs: Dict[int, BotProcess] = {}
daily_helpers: Dict[str, DailyRESTHelper] = {}

# Load environment variables
load_dotenv(override=True)

# Store active processes and their communication queues, keyed by room_url
process_manager = {
    "processes": {},  # type: Dict[str, asyncio.subprocess.Process]
    "queues": {},    # type: Dict[str, asyncio.Queue]
    "variables": {},  # type: Dict[str, dict]
    "call_ids": {}   # type: Dict[str, str] # Maps room_url to callId
}

async def create_dialin_daily_room(callId: str, callDomain: Optional[str] = None):
    """Creates room and starts subprocess, but handles communication separately"""
    logger.info("🏗️ Creating new room...")
    room: DailyRoomObject = await daily_helpers["rest"].create_room(params=params)
    if not room:
        raise HTTPException(status_code=500, detail="❌ Failed to get room")
    
    token = await daily_helpers["rest"].get_token(room.url, TOKEN_EXPIRY_TIME)
    if not token:
        raise HTTPException(status_code=500, detail="❌ Failed to get room token")
    
    # Create queue for this process
    process_manager["queues"][room.url] = asyncio.Queue()
    
    # Start the subprocess
    try:
        process = await asyncio.create_subprocess_exec(
            "python3",
            "-m",
            "bot",
            "-u", room.url,
            "-t", token,
            "-i", callId,
            "-d", callDomain if callDomain else "",
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE,
            cwd=os.path.dirname(os.path.abspath(__file__))
        )
        
        # Store process reference and callId mapping
        process_manager["processes"][room.url] = process
        process_manager["variables"][room.url] = {}
        process_manager["call_ids"][room.url] = callId
        
        # Start monitoring in background without awaiting
        asyncio.create_task(monitor_process_output(room.url))
        
    except Exception as e:
        # Clean up if process creation fails
        if room.url in process_manager["queues"]:
            del process_manager["queues"][room.url]
        raise HTTPException(status_code=500, detail=f"Failed to start subprocess: {e}")
    
    return room

async def monitor_process_output(room_url: str):
    """Separate monitoring function that runs independently"""
    process = process_manager["processes"].get(room_url)
    if not process:
        return
    
    try:
        while True:
            line = await process.stdout.readline()
            if not line:
                break
                
            decoded_line = line.decode().strip()
            
            # Handle transcript updates
            if decoded_line.startswith("TRANSCRIPT_UPDATE:"):
                transcript = decoded_line.split(":", 1)[1]
                process_manager["variables"][room_url]["transcript"] = transcript
                await process_manager["queues"][room_url].put(f"Transcript updated: {transcript}")
            # Handle other variable updates
            elif decoded_line.startswith("VARIABLE_UPDATE:"):
                var_data = decoded_line.split(":", 2)
                if len(var_data) == 3:
                    var_name, var_value = var_data[1], var_data[2]
                    process_manager["variables"][room_url][var_name] = var_value
            
            await process_manager["queues"][room_url].put(decoded_line)
            
        # Handle stderr
        async for line in process.stderr:
            await process_manager["queues"][room_url].put(f"ERROR: {line.decode().strip()}")
            
    except Exception as e:
        logger.error(f"Error monitoring process for room {room_url}: {e}")
    finally:
        await cleanup_process(room_url)

async def cleanup_process(room_url: str):
    """Clean up process resources"""
    for key in ["processes", "queues", "variables", "call_ids"]:
        if room_url in process_manager[key]:
            del process_manager[key][room_url]

@app.post("/daily_start_bot")
async def daily_start_bot(request: Request) -> JSONResponse:
    """Main webhook endpoint"""
    try:
        data = await request.json()
        if "test" in data:
            return JSONResponse({"test": True})
            
        callId = data.get("callId")
        callDomain = data.get("callDomain")
        
        if not callId:
            raise HTTPException(status_code=400, detail="Missing required 'callId'")
            
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid request format")

    logger.info(f"📞 Received call Daily dialin with CallId: {callId}, CallDomain: {callDomain}")
    
    room = await create_dialin_daily_room(callId, callDomain)
    
    return JSONResponse({
        "room_url": room.url,
        "sipUri": room.config.sip_endpoint
    })

@app.get("/bot_output/{room_url:path}")
async def get_bot_output(room_url: str):
    """Stream bot output for monitoring"""
    if room_url not in process_manager["queues"]:
        raise HTTPException(status_code=404, detail="Bot process not found")
    
    queue = process_manager["queues"][room_url]
    
    async def generate():
        while True:
            message = await queue.get()
            if message is None:
                break
            yield f"{message}\n"
    
    return StreamingResponse(generate(), media_type="text/plain")

@app.get("/bot_variable/{room_url:path}/{variable_name}")
async def get_bot_variable(room_url: str, variable_name: str):
    """Get specific variable value from bot process"""
    if room_url not in process_manager["variables"]:
        raise HTTPException(status_code=404, detail="Bot process not found")
    
    variables = process_manager["variables"][room_url]
    if variable_name not in variables:
        raise HTTPException(status_code=404, detail=f"Variable {variable_name} not found")
        
    return JSONResponse({
        "variable_name": variable_name,
        "value": variables[variable_name]
    })

@app.get("/transcript/{room_url:path}")
async def get_transcript(room_url: str):
    """Get the current transcript for a specific room"""
    if room_url not in process_manager["variables"]:
        raise HTTPException(status_code=404, detail="Room not found")
    
    variables = process_manager["variables"][room_url]
    transcript = variables.get("transcript", "")
    
    return JSONResponse({
        "room_url": room_url,
        "transcript": transcript
    })

if __name__ == "__main__":
    configure_logging()
    
    parser = argparse.ArgumentParser(description="BearHug FastAPI Server")
    parser.add_argument("--host", type=str, default=DEFAULT_HOST, help="Host address")
    parser.add_argument("--port", type=int, default=DEFAULT_PORT, help="Port number")
    parser.add_argument("--reload", action="store_true", help="Reload code on change")

    config = parser.parse_args()

    logger.info(f"🚀 Starting server on {config.host}:{config.port}")
    if config.reload:
        logger.info("🔄 Hot reload enabled")

    try:    
        import uvicorn
        uvicorn.run(
            "server:app",
            host=config.host,
            port=config.port,
            reload=config.reload,
        )
    except KeyboardInterrupt:
        print("⬇️ Pipecat server is shutting down...")




#Graveyard

### Old web redirect way of starting a room with a bot
# @app.get("/")
# async def start_agent(request: Request):
#     # Endpoint to create a new room and start a bot agent
#     client_ip = request.client.host
#     logger.info(f"📞 New agent request from {client_ip}")
    
#     try:
#         # Create new room
#         logger.info("🏗️ Creating new room...")

#         # Web communication
#         room: DailyRoomObject = await daily_helpers["rest"].create_room(DailyRoomParams())

#         # Check bot limits
#         num_bots_in_room = sum(
#             1 for proc in bot_procs.values() 
#             if proc[1] == room.url and proc[0].poll() is None
#         )
#         if num_bots_in_room >= MAX_BOTS_PER_ROOM:
#             logger.warning(f"⚠️ Max bot limit reached for room: {room.url}")
#             raise HTTPException(
#                 status_code=500, 
#                 detail=f"Max bot limit reached for room: {room.url}"
#             )

#         # Get room token
#         logger.info("🔑 Requesting room token...")
#         token = await daily_helpers["rest"].get_token(room.url)
#         if not token:
#             logger.error(f"❌ Failed to get token for room: {room.url}")
#             raise HTTPException(
#                 status_code=500,
#                 detail=f"Failed to get token for room: {room.url}"
#             )
#         logger.info("✅ Token acquired successfully")

#         # Start bot process with SSL certificate path
#         logger.info("🤖 Starting bot process...")
        
#         proc = subprocess.Popen(
#             [f"python3 -m bot -u {room.url} -t {token}"],
#             shell=True,
#             bufsize=1,
#             cwd=os.path.dirname(os.path.abspath(__file__)),
#         )
#         bot_procs[proc.pid] = (proc, room.url)
#         logger.info(f"✅ Bot started successfully with PID: {proc.pid}")

#         # The bot is started so next whatever hits the "/" endpoint is redirected to the daily room url 
#         return RedirectResponse(room.url)

#     except HTTPException:
#         raise
#     except Exception as e:
#         logger.error(f"❌ Unexpected error starting agent: {e}")
#         raise HTTPException(status_code=500, detail=str(e))