import asyncio
import hashlib
import logging
from .registry import Registry
from . import globals

class Namespace:
    def __init__(self, name, secret_key, logger=None, use_encryption=False):
        self.name = name
        self.secret_key = secret_key
        self.logger = logger or logging.getLogger(f"spine.{name}")
        self.use_encryption = use_encryption
        self.encryption_key = None
        
        if use_encryption:
            # Hash secret key to get 32-byte key for AES-256
            self.encryption_key = hashlib.sha256(secret_key.encode()).digest()
            
        self.reg = Registry(name, self.logger)
        self.tasks = []

    async def join(self):
        await self.reg.start()
        # Start the janitor as a background task
        self.tasks.append(asyncio.create_task(self.reg.janitor()))
        return self

    async def disconnect(self):
        for task in self.tasks:
            task.cancel()
        self.reg.stop()

    async def get_service_addr(self, name):
        return await self.reg.lookup(globals.ZERO_CONF_SERVICE_PREFIX + name)

    async def get_publisher_addr(self, name):
        return await self.reg.lookup(globals.ZERO_CONF_PUBLISHER_PREFIX + name)
