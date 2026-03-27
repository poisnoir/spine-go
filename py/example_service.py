import asyncio
import logging
import sys
from spine import Namespace, new_service

async def main():
    logging.basicConfig(level=logging.INFO)
    logger = logging.getLogger("example")
    
    ns = Namespace("example", "meow", logger=logger, use_encryption=False)
    await ns.join()

    async def len_func(input_str):
        return len(input_str)

    async def print_func(input_str):
        print(f"Service received: {input_str}")
        return f"printed {input_str}"

    service1 = new_service(ns, "string_length", len_func)
    await service1.start()
    
    service2 = new_service(ns, "print", print_func)
    await service2.start()

    logger.info("Services started. Press Ctrl+C to stop.")
    try:
        while True:
            await asyncio.sleep(1)
    except KeyboardInterrupt:
        pass
    finally:
        await ns.disconnect()

if __name__ == "__main__":
    asyncio.run(main())
