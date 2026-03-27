import asyncio
import logging
from .service_common import generate_service, handle_caller_request
from .mad import new_mad
from . import globals

class ThreadedService:
    def __init__(self, namespace, name, handler):
        self.name = name
        self.namespace = namespace
        self.handler = handler
        self.logger = namespace.logger.getChild(f"service.{name}")
        self.ctx = asyncio.Event()
        self.tasks = []
        
        # Encoders
        self.key_encoder = None
        self.value_encoder = None
        self.error_encoder = new_mad("string")
        self.info = None

    async def start(self):
        self.key_encoder, self.value_encoder, self.info = await generate_service(self.namespace, self.name)
        
        # In actual KCP library, we'd start a KCP server here
        # self.tasks.append(asyncio.create_task(self.run_listener()))

    async def process_request(self, key):
        try:
            # For threaded service, we just run the handler directly.
            # In Python asyncio, we might want to run it in a separate thread if it's blocking
            # but we'll assume it's an async handler or the user handles it.
            # If we want to truly mimic Go's threaded behavior:
            # return await asyncio.to_thread(self.handler, key)
            
            # Since this is an async framework, let's try awaiting it
            if asyncio.iscoroutinefunction(self.handler):
                result = await self.handler(key)
            else:
                result = await asyncio.to_thread(self.handler, key)
            return result, None
        except Exception as e:
            self.logger.error(f"Handler error: {e}")
            return None, e

    async def client_handler(self, reader, writer):
        await handle_caller_request(reader, writer, self.key_encoder, self.value_encoder, self.process_request, self.logger)

    async def close(self):
        self.ctx.set()
        for task in self.tasks:
            task.cancel()
        if self.info:
            self.namespace.reg.zc.unregister_service(self.info)

def new_threaded_service(namespace, name, handler):
    service = ThreadedService(namespace, name, handler)
    # Note: need to await service.start() in asyncio context
    return service
