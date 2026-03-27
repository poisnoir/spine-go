import asyncio
import logging
import socket
import time
from zeroconf import ServiceBrowser, ServiceInfo, Zeroconf
from . import globals

class Registry:
    def __init__(self, name, logger=None):
        self.name = name
        self.components = {}
        self.logger = logger or logging.getLogger("registry")
        self.zc = Zeroconf()
        self.browser = None

    async def start(self):
        # Browsing for services of a certain type
        service_type = f"{self.name}{globals.ZERO_CONF_TYPE}."
        self.browser = ServiceBrowser(self.zc, service_type, self)
        
    def add_service(self, zc, type, name):
        info = zc.get_service_info(type, name)
        if info:
            addr = socket.inet_ntoa(info.addresses[0])
            port = info.port
            self.components[name] = {
                "addr": f"{addr}:{port}",
                "expiry": time.time() + (info.properties.get(b'ttl', 120))
            }
            self.logger.info(f"Added service: {name} at {addr}:{port}")

    def update_service(self, zc, type, name):
        # We can treat update same as add for simplicity
        self.add_service(zc, type, name)

    def remove_service(self, zc, type, name):
        if name in self.components:
            del self.components[name]
            self.logger.info(f"Removed service: {name}")

    async def lookup(self, name, timeout=5.0):
        # First check in-memory cache
        # ZeroConf full name might be different from simple name
        full_name = f"{name}.{self.name}{globals.ZERO_CONF_TYPE}.{globals.ZERO_CONF_DOMAIN}"
        
        if full_name in self.components:
            comp = self.components[full_name]
            if time.time() < comp["expiry"]:
                return comp["addr"]

        # If not in cache or expired, try a direct lookup
        service_type = f"{self.name}{globals.ZERO_CONF_TYPE}."
        info = self.zc.get_service_info(service_type, f"{name}.{service_type}", timeout=int(timeout * 1000))
        if info and info.addresses:
            addr = socket.inet_ntoa(info.addresses[0])
            port = info.port
            return f"{addr}:{port}"
        
        raise Exception(f"Service not found: {name}")

    def stop(self):
        if self.browser:
            self.browser.cancel()
        self.zc.close()

    async def janitor(self, interval=30):
        while True:
            await asyncio.sleep(interval)
            now = time.time()
            expired = [name for name, comp in self.components.items() if now > comp["expiry"]]
            for name in expired:
                self.logger.debug(f"Cleaning up expired component: {name}")
                del self.components[name]

    def remove_on_failure(self, failed_comp):
        if failed_comp in self.components:
            del self.components[failed_comp]
