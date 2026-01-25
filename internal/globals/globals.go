package globals

// Header
const HEADER_LENGTH = 5
const STATUS_CODE_INDEX = 0
const MESSAGE_LENGTH_INDEX = 1

// Hash
const HASH_LENGTH = 32

// Status Codes
const OK_STATUS uint8 = 0
const ERROR_INCORRECT_HASH uint8 = 255
const ERROR_CORRUPT_PAYLOAD uint8 = 254
const ERROR_HANDLER_INTERNAL_ERROR uint8 = 253

// Zero conf
const ZERO_CONF_SERVICE = "_botzilla._tcp"
const ZERO_CONF_DOMAIN = "local."
const ZERO_CONF_SERVICE_PREFIX = "_service"
const ZERO_CONF_PUBLISHER_PREFIX = "_publisher"
