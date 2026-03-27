import asyncio
import logging
from spine import Namespace, ServiceCaller

async def main():
    logging.basicConfig(level=logging.INFO)
    logger = logging.getLogger("example_caller")
    
    ns = Namespace("example", "meow", logger=logger, use_encryption=False)
    await ns.join()

    # Create caller for string_length service
    caller1 = ServiceCaller(ns, "string_length")
    await caller1.start()

    # Create caller for print service
    caller2 = ServiceCaller(ns, "print")
    await caller2.start()

    # Wait for connections (simplified)
    await asyncio.sleep(1)

    # Call service 1
    result1, err = await caller1.call("hello world", timeout=5.0)
    if err:
        logger.error(f"Error calling string_length: {err}")
    else:
        logger.info(f"string_length result: {result1}")

    # Call service 2
    result2, err = await caller2.call("hello from python", timeout=5.0)
    if err:
        logger.error(f"Error calling print: {err}")
    else:
        logger.info(f"print result: {result2}")

    await ns.disconnect()

if __name__ == "__main__":
    asyncio.run(main())
