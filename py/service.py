import asyncio
import logging
from .service_common import generate_service, handle_caller_request
from .mad import new_mad
from . import globals

class Service:
    def __init__(self, namespace, name, handler):
        self.name = name
        self.namespace = namespace
        self.handler = handler
        self.logger = namespace.logger.getChild(f"service.{name}")
        self.requests = asyncio.Queue(maxsize=100)
        self.ctx = asyncio.Event()
        self.tasks = []
        
        # Encoders
        self.key_encoder = None
        self.value_encoder = None
        self.error_encoder = new_mad("string")
        self.info = None

    async def start(self):
        self.key_encoder, self.value_encoder, self.info = await generate_service(self.namespace, self.name)
        
        # Start handler loop
        self.tasks.append(asyncio.create_task(self.run_handler()))
        
        # Start listener
        # In actual KCP library, we'd start a KCP server here
        # self.tasks.append(asyncio.create_task(self.run_listener()))
        
    async def run_handler(self):
        while not self.ctx.is_set():
            try:
                request_data, output_queue = await self.requests.get()
                try:
                    # Run handler
                    response = await self.handler(request_data)
                    await output_queue.put((response, None))
                except Exception as e:
                    self.logger.error(f"Handler error: {e}")
                    await output_queue.put((None, e))
                finally:
                    self.requests.task_done()
            except asyncio.CancelledError:
                break
            except Exception as e:
                self.logger.error(f"Handler loop error: {e}")

    async def process_request(self, key):
        output_queue = asyncio.Queue(maxsize=1)
        await self.requests.put((key, output_queue))
        return await output_queue.get()

    async def client_handler(self, reader, writer):
        await handle_caller_request(reader, writer, self.key_encoder, self.value_encoder, self.process_request, self.logger)

    async def close(self):
        self.ctx.set()
        for task in self.tasks:
            task.cancel()
        if self.info:
            self.namespace.reg.zc.unregister_service(self.info)

def new_service(namespace, name, handler):
    service = Service(namespace, name, handler)
    # Note: need to await service.start() in asyncio context
    return service
