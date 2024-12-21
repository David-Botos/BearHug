# import asyncio
# import logging
# import uuid
# from typing import Optional
# from datetime import datetime
# from concurrent.futures import ThreadPoolExecutor
# from supabase import supabase

# logger = logging.getLogger(__name__)

# class TranscriptProcessor:
#     _instance: Optional['TranscriptProcessor'] = None
#     _executor: Optional[ThreadPoolExecutor] = None
    
#     def __new__(cls):
#         if cls._instance is None:
#             cls._instance = super(TranscriptProcessor, cls).__new__(cls)
#             cls._executor = ThreadPoolExecutor(max_workers=2)
#         return cls._instance

#     @classmethod
#     def get_instance(cls):
#         if cls._instance is None:
#             cls._instance = TranscriptProcessor()
#         return cls._instance

#     def __init__(self):
#         if not hasattr(self, 'initialized'):
#             self.initialized = True
#             self.processing_tasks = set()
#             self.loop = asyncio.get_event_loop()

#     async def _store_transcript(self, transcript: str):
#         """
#         Stores the transcript in Supabase
#         """
#         try:
#             # TODO: Implement Supabase storage logic here
#             logger.info("Storing transcript in Supabase")
#             store_res = (supabase.table("transcripts").insert({"id": uuid.uuid4(), "full_transcript": transcript}).execute())
#             transcript_uuid = store_res.data[0].id
#             # TODO: Call analysis
            
#         except Exception as e:
#             logger.error(f"Error storing transcript: {e}")
#             raise

#     def _handle_task_completion(self, task, task_type: str, task_set: set):
#         """
#         Handles the completion of a task and removes it from the tracking set
#         """
#         task_set.discard(task)
#         try:
#             task.result()  # Get result to handle any exceptions
#             logger.info(f"Successfully completed {task_type} transcript task")
#         except Exception as e:
#             logger.error(f"Error in {task_type} transcript task: {e}")

# # Create singleton instance
# processor = TranscriptProcessor.get_instance()

# async def store_transcript_async(transcript: str):
#     """
#     Store transcript asynchronously without blocking
#     """
#     if not transcript.strip():
#         logger.warning("Empty transcript received, skipping processing")
#         return

#     try:
#         # Create detached task
#         store_task = processor.loop.create_task(processor._store_transcript(transcript))
        
#         # Add to processing set but don't await
#         processor.processing_tasks.add(store_task)
        
#         # Add callback to remove task from set when done
#         store_task.add_done_callback(
#             lambda t: processor._handle_task_completion(t, "store", processor.processing_tasks)
#         )
        
#     except Exception as e:
#         logger.error(f"Failed to create transcript storage task: {e}")