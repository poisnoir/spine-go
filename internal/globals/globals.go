package globals

// Header
const HEADER_LENGTH int = 5
const STATUS_CODE_INDEX int = 0
const MESSAGE_LENGTH_INDEX int = 1

// Hash
const MAX_PACKET_SIZE int = 4096

// Status Codes
const OK_STATUS uint8 = 0
const PING_CODE uint8 = 1
const ERROR_INCORRECT_HASH uint8 = 255
const ERROR_CORRUPT_PAYLOAD uint8 = 254
const ERROR_HANDLER_INTERNAL_ERROR uint8 = 253

// Zero conf
const ZERO_CONF_PUBLISHER_PREFIX = "publisher_"
const ZERO_CONF_SERVICE_PREFIX = "service_"
const ZERO_CONF_TYPE = "_botzilla._tcp"
const ZERO_CONF_DOMAIN = "local."
