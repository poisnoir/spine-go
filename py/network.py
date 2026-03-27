import asyncio
import logging
from . import globals

async def write_with_response(writer, data, has_response=True):
    writer.write(data)
    # The KCP library should handle flushing.
    if not has_response:
        return 0, None
    
    # Reading response
    # This depends on the specific KCP library implementation.
    # We'll assume a standard asyncio-like reader.
    response = await writer.read(globals.MAX_PACKET_SIZE)
    return len(response), response

async def ping(reader, writer):
    buf = bytes([globals.PING_CODE])
    n, res = await write_with_response(writer, buf, True)
    if not res or res[0] != globals.PONG_CODE:
        raise Exception(globals.ERROR_PING)
    return True
