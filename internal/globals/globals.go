package globals

// Header
const HEADER_LENGTH int = 5
const STATUS_CODE_INDEX int = 0
const MESSAGE_LENGTH_INDEX int = 1

// Hash
const MAX_PACKET_SIZE int = 4096

// Status Codes
const OK_STATUS_CODE uint8 = 0
const PING_CODE uint8 = 1
const PONG_CODE uint8 = 2
const SERVICE_REQUEST uint8 = 3

const ERROR_SERVICE_ERROR_CODE uint8 = 252
const ERROR_INVALID_OPERATION_CODE uint8 = 253
const ERROR_CORRUPT_PAYLOAD_CODE uint8 = 254
const ERROR_HANDLER_INTERNAL_ERROR_CODE uint8 = 255

// Zero conf
const ZERO_CONF_PUBLISHER_PREFIX = "publisher_"
const ZERO_CONF_SERVICE_PREFIX = "service_"
const ZERO_CONF_TYPE = "_botzilla._tcp"
const ZERO_CONF_DOMAIN = "local."

const ERROR_CORRUPT_PAYLOAD = "CORRUPT_PAYLOAD"
const ERROR_PING = "service didn't respond to ping"
const ERROR_PAYLOAD_SIZE = "failed to encode key. key is too big. max key size is 4kb"
