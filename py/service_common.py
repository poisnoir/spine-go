import asyncio
import logging
import socket
from zeroconf import ServiceInfo, Zeroconf
from . import globals
from .mad import new_mad

async def generate_service(namespace, name, port=0):
    logger = namespace.logger.getChild(f"service.{name}")

    key_enc = new_mad()
    value_enc = new_mad()

    # KCP Listener setup
    # This part depends on the KCP library used.
    # We'll use a placeholder for now as we don't have the exact library.
    # In a real scenario, we'd use 'kcp.KCPServer' or similar.
    
    # Discovery registration
    zc = namespace.reg.zc
    service_type = f"{namespace.name}{globals.ZERO_CONF_TYPE}."
    full_name = f"{globals.ZERO_CONF_SERVICE_PREFIX}{name}.{service_type}"
    
    # For now, let's assume we get the local IP
    local_ip = socket.gethostbyname(socket.gethostname())
    
    info = ServiceInfo(
        type_=service_type,
        name=full_name,
        addresses=[socket.inet_aton(local_ip)],
        port=port,
        properties={'id': f'spine_service_{name}'},
        server=f"{socket.gethostname()}.local."
    )
    
    zc.register_service(info)
    
    return key_enc, value_enc, info

async def establish_connection(reader, writer, key_code, value_code, logger):
    try:
        data = await reader.read(globals.MAX_PACKET_SIZE)
        if data != key_code:
            logger.error("Failed to establish connection: invalid key code")
            return False
            
        writer.write(bytes([globals.OK_STATUS_CODE]))
        
        data = await reader.read(globals.MAX_PACKET_SIZE)
        if data != value_code:
            logger.error("Failed to establish connection: invalid value code")
            return False
            
        writer.write(bytes([globals.OK_STATUS_CODE]))
        return True
    except Exception as e:
        logger.error(f"Error establishing connection: {e}")
        return False

async def handle_caller_request(reader, writer, key_encoder, value_encoder, process_request, logger):
    # Establish connection with handshake
    if not await establish_connection(reader, writer, key_encoder.code(), value_encoder.code(), logger):
        writer.close()
        return

    try:
        while True:
            data = await reader.read(globals.MAX_PACKET_SIZE)
            if not data:
                break

            if data[0] == globals.PING_CODE:
                writer.write(bytes([globals.PONG_CODE]))
                continue

            if data[0] != globals.SERVICE_REQUEST:
                logger.error("Received invalid operation code")
                writer.write(bytes([globals.ERROR_INVALID_OPERATION_CODE]))
                continue

            # Decode key
            try:
                key = key_encoder.decode(data[1:])
            except Exception as e:
                logger.error(f"Unable to decode key: {e}")
                continue

            # Process request
            try:
                response, err = await process_request(key)
                if err:
                    logger.error(f"Handler failed: {err}")
                    writer.write(bytes([globals.ERROR_SERVICE_ERROR_CODE]))
                    continue
                
                # Encode response
                res_data = value_encoder.encode(response)
                writer.write(res_data)
            except Exception as e:
                logger.error(f"Process request failed: {e}")
                writer.write(bytes([globals.ERROR_HANDLER_INTERNAL_ERROR_CODE]))
                
    except asyncio.CancelledError:
        pass
    except Exception as e:
        logger.error(f"Connection error: {e}")
    finally:
        writer.close()
