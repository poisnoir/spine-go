import asyncio
import logging
import time
from .mad import new_mad
from . import globals
from .network import ping

class ServiceCaller:
    def __init__(self, namespace, service_name):
        self.namespace = namespace
        self.service_name = service_name
        self.logger = namespace.logger.getChild(f"service_caller.{service_name}")
        
        self.key_encoder = new_mad()
        self.value_encoder = new_mad()
        self.error_encoder = new_mad("string")
        
        self.reader = None
        self.writer = None
        self.is_connected = False
        self.requests = asyncio.Queue(maxsize=100)
        self.ctx = asyncio.Event()
        self.tasks = []

    async def start(self):
        self.tasks.append(asyncio.create_task(self.run()))

    async def run(self):
        while not self.ctx.is_set():
            if self.is_connected:
                try:
                    # Timer for heartbeat (simplified)
                    # In a real implementation, we'd use wait_for with a timeout
                    try:
                        request_data, output_queue = await asyncio.wait_for(self.requests.get(), timeout=10.0)
                        
                        # Send request
                        try:
                            result = await self.send(request_data)
                            await output_queue.put((result, None))
                        except Exception as e:
                            self.logger.error(f"Send failed: {e}")
                            self.is_connected = False
                            await output_queue.put((None, e))
                        finally:
                            self.requests.task_done()
                            
                    except asyncio.TimeoutError:
                        # Ping
                        try:
                            await ping(self.reader, self.writer)
                        except Exception as e:
                            self.logger.error(f"Ping failed: {e}")
                            self.is_connected = False
                except asyncio.CancelledError:
                    break
                except Exception as e:
                    self.logger.error(f"Run loop error: {e}")
                    self.is_connected = False
            else:
                # Try to connect with exponential backoff
                await self.connect_with_retry()

    async def connect_with_retry(self):
        delay = 1.0
        while not self.is_connected and not self.ctx.is_set():
            try:
                await self.connect()
                self.is_connected = True
                self.logger.info(f"Connected to {self.service_name}")
            except Exception as e:
                self.logger.error(f"Failed to connect to {self.service_name}: {e}")
                await asyncio.sleep(delay)
                delay = min(delay * 2, 60.0)

    async def connect(self):
        # Finding the service
        addr = await self.namespace.get_service_addr(self.service_name)
        host, port = addr.split(':')
        port = int(port)

        # Establishing KCP connection
        # Placeholder for actual KCP library connection
        # self.reader, self.writer = await kcp.open_connection(host, port)
        
        # Simulating handshake for now
        # writer.write(self.key_encoder.code())
        # ... validation ...
        # writer.write(self.value_encoder.code())
        # ... validation ...
        
        pass

    async def send(self, key):
        request_size = self.key_encoder.get_required_size(key) + 1
        if request_size > globals.MAX_PACKET_SIZE:
            raise Exception(globals.ERROR_PAYLOAD_SIZE)

        # Build packet: [OP_CODE][ENCODED_KEY]
        packet = bytes([globals.SERVICE_REQUEST]) + self.key_encoder.encode(key)
        
        self.writer.write(packet)
        # Wait for response
        response = await self.reader.read(globals.MAX_PACKET_SIZE)
        
        if not response:
            raise Exception("No response from service")
            
        if response[0] != globals.OK_STATUS_CODE:
            raise Exception(globals.ERROR_SERVICE_HANDLER)

        return self.value_encoder.decode(response[1:])

    async def call(self, key, timeout=None):
        output_queue = asyncio.Queue(maxsize=1)
        await self.requests.put((key, output_queue))
        
        try:
            if timeout:
                return await asyncio.wait_for(output_queue.get(), timeout=timeout)
            else:
                return await output_queue.get()
        except asyncio.TimeoutError:
            raise Exception("Call timeout")

    async def close(self):
        self.ctx.set()
        for task in self.tasks:
            task.cancel()
        if self.writer:
            self.writer.close()
            # await self.writer.wait_closed()
