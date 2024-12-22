import asyncio
import os
from typing import Optional
import certifi
import ssl

import aiohttp
from dotenv import load_dotenv
from loguru import logger
from runner import configure
from prompt import generate_shelter_prompt, CommunityServices
# from analysis import store_transcript_async

from pipecat.audio.vad.silero import SileroVADAnalyzer
from pipecat.frames.frames import EndFrame, LLMMessagesFrame
from pipecat.pipeline.pipeline import Pipeline
from pipecat.pipeline.runner import PipelineRunner
from pipecat.pipeline.task import PipelineParams, PipelineTask
from pipecat.processors.audio.audio_buffer_processor import AudioBufferProcessor
# from pipecat.services.canonical import CanonicalMetricsService
from pipecat.services.elevenlabs import ElevenLabsTTSService
from pipecat.services.anthropic import AnthropicLLMService, AnthropicLLMContext
from pipecat.transports.services.daily import DailyParams, DailyTransport

# Constants
INITIAL_SILENCE_TIMEOUT = 5  # seconds to wait before first greeting
MAX_SILENCE_DURATION = 30    # maximum seconds of silence before check-in
RECONNECTION_TIMEOUT = 60    # seconds to wait before considering call dropped

class CallState:
    def __init__(self):
        self.is_speaking: bool = False
        self.has_greeted: bool = False
        self.has_ended: bool = False
        self.current_transcription: str = ""
        self.initial_silence_timer: Optional[asyncio.Task] = None
        self.extended_silence_timer: Optional[asyncio.Task] = None
        self.participant_id: Optional[str] = None

def get_transcript_from_context(context) -> str:
    """Extract the conversation transcript from the AnthropicLLMContext"""
    transcript_lines = []
    
    # Access the messages directly from the context
    for message in context.messages:
        role = message["role"]
        content = message["content"]
        
        # Skip system messages
        if role == "system":
            continue
            
        # Format the speaker label
        speaker = "Bot" if role == "assistant" else "User"
        
        # Handle different content formats
        if isinstance(content, str):
            transcript_lines.append(f"{speaker}: {content}")
        elif isinstance(content, list):
            # Process each content block
            for block in content:
                if block["type"] == "text":
                    transcript_lines.append(f"{speaker}: {block['text']}")
                # Skip tool_use, tool_result, and image blocks from transcript
    
    return "\n".join(transcript_lines)

async def send_final_transcript(transcript: str):
    """Send final transcript to stdout for server to capture"""
    print(f"TRANSCRIPT_UPDATE:{transcript}")

async def main():
    ssl_context = ssl.create_default_context(cafile=certifi.where())
    connector = aiohttp.TCPConnector(ssl=ssl_context)
    async with aiohttp.ClientSession(connector=connector) as session:
        logger.info("🚀 Initializing bot...")
        
        try:
            (room_url, token) = await configure(session)
            logger.info("✅ Room configuration successful")
        except Exception as e:
            logger.error(f"❌ Failed to configure room: {e}")
            return

        # Initialize call state
        state = CallState()

        # Configure transport
        transport = DailyTransport(
            room_url,
            token,
            "Chatbot",
            DailyParams(
                audio_out_enabled=True,
                audio_in_enabled=True,
                camera_out_enabled=False,
                vad_enabled=True,
                vad_audio_passthrough=True,
                vad_analyzer=SileroVADAnalyzer(),
                transcription_enabled=True,
            ),
        )

        # Initialize services
        try:
            tts = ElevenLabsTTSService(
                api_key=os.getenv("ELEVENLABS_API_KEY"),
                voice_id=os.getenv("ELEVENLABS_VOICE_ID"),
                aiohttp_session=session,
            )
            llm = AnthropicLLMService(api_key=os.getenv("ANTHROPIC_API_KEY"))
            logger.info("✅ Services initialized successfully")
        except Exception as e:
            logger.error(f"❌ Failed to initialize services: {e}")
            return

        prompt = generate_shelter_prompt(
            'Bread of Life', 
            'Bethany Presbyterian', 
            [CommunityServices.SHELTER, CommunityServices.UNEMPLOYMENT_RESOURCES]
        )
        
        # Initialize context and pipeline
        messages = [{
            "role": "system",
            "content": prompt
        }]
        context = AnthropicLLMContext(messages)
        context_aggregator = llm.create_context_aggregator(context)

        audio_buffer_processor = AudioBufferProcessor()
        # canonical = CanonicalMetricsService(
        #     audio_buffer_processor=audio_buffer_processor,
        #     aiohttp_session=session,
        #     api_key=os.getenv("CANONICAL_API_KEY"),
        #     call_id=str(uuid.uuid4()),
        #     assistant="pipecat-chatbot",
        #     assistant_speaks_first=True,
        # )

        pipeline = Pipeline([
            transport.input(),
            context_aggregator.user(),
            # stt??
            llm,
            tts,
            transport.output(),
            audio_buffer_processor,
            # canonical,
            context_aggregator.assistant(),
        ])

        task = PipelineTask(pipeline, PipelineParams(allow_interruptions=True))

        async def check_initial_silence():
            # Handler for initial silence after connection
            if not state.has_greeted and not state.is_speaking:
                logger.info("🤔 No initial response detected, initiating greeting")
                await task.queue_frames([LLMMessagesFrame([{
                    "role": "assistant",
                    "content": "Hello? Is anyone there?"
                }])])
                state.has_greeted = True
                # Start monitoring for extended silence
                start_extended_silence_monitoring()

        async def check_extended_silence():
            # Handler for extended silence during conversation
            if not state.is_speaking and not state.has_ended:
                logger.info("⏰ Extended silence detected, checking connection")
                await task.queue_frames([LLMMessagesFrame([{
                    "role": "assistant",
                    "content": "I haven't heard anything for a while. Are you still there?"
                }])])

        def start_extended_silence_monitoring():
            # Starts the extended silence monitoring timer
            if state.extended_silence_timer:
                state.extended_silence_timer.cancel()
            state.extended_silence_timer = asyncio.create_task(asyncio.sleep(MAX_SILENCE_DURATION))
            state.extended_silence_timer.add_done_callback(
                lambda _: asyncio.create_task(check_extended_silence())
            )

        @transport.event_handler("on_transcription_message")
        async def on_transcription_message(transport, message):
            # Handles transcription messages from the participant
            # Cancel silence timer if participant speaks
            if state.initial_silence_timer and not state.initial_silence_timer.done():
                state.initial_silence_timer.cancel()
                logger.info("✅ Initial speech detected, canceling greeting")

            if message["rawResponse"]["is_final"]:
                text = message["text"]
                if not state.is_speaking:
                    return

                state.has_greeted = True
                state.current_transcription = text
                logger.info(f"🗣️ Transcribed: {text}")
                
                try:
                    await task.queue_frames([LLMMessagesFrame([{
                        "role": "user",
                        "content": state.current_transcription
                    }])])
                    state.current_transcription = ""
                    # Reset extended silence timer
                    start_extended_silence_monitoring()
                except Exception as e:
                    logger.error(f"❌ Failed to process transcription: {e}")

        @transport.event_handler("on_participant_updated")
        async def on_participant_updated(transport, participant):
            # Check if this update includes VAD information
            if "vad" in participant:
                vad_status = participant["vad"]
                if vad_status:
                    state.is_speaking = True
                    logger.debug("🎤 Voice activity started")
                else:
                    state.is_speaking = False
                    logger.debug("🛑 Voice activity stopped")

        @transport.event_handler("on_first_participant_joined")
        async def on_first_participant_joined(transport, participant):
            # Handles first participant joining the call
            state.participant_id = participant["id"]
            logger.info(f"👋 First participant joined: {state.participant_id}")
            
            try:
                await transport.capture_participant_transcription(state.participant_id)
                logger.info("✅ Transcription capture started")
                
                # Start initial silence timer
                state.initial_silence_timer = asyncio.create_task(asyncio.sleep(INITIAL_SILENCE_TIMEOUT))
                state.initial_silence_timer.add_done_callback(
                    lambda _: asyncio.create_task(check_initial_silence())
                )
            except Exception as e:
                logger.error(f"❌ Failed to setup participant: {e}")

        @transport.event_handler("on_participant_left")
        async def on_participant_left(transport, participant, reason):
            # Handles participant leaving the call
            logger.info(f"👋 Participant left: {participant['id']}")
            state.has_ended = True
            if state.extended_silence_timer:
                state.extended_silence_timer.cancel()

            transcript = get_transcript_from_context(context)
            send_final_transcript(f"TRANSCRIPT_UPDATE:{transcript}")

            await task.queue_frame(EndFrame())

        @transport.event_handler("on_call_state_updated")
        async def on_call_state_updated(transport, state_update):
            # Handles call state updates
            if state_update == "left":
                logger.info("📞 Call ended")
                state.has_ended = True
                if state.extended_silence_timer:
                    state.extended_silence_timer.cancel()

                transcript = get_transcript_from_context(context)
                send_final_transcript(f"TRANSCRIPT_UPDATE:{transcript}")

                await task.queue_frame(EndFrame())

        # Start the pipeline
        runner = PipelineRunner()
        logger.info("✅ Bot initialization complete, starting pipeline")
        await runner.run(task)

if __name__ == "__main__":
    asyncio.run(main())