package de.arcan.arclib.protocol

/** Protocol operation codes. Format: 8-bit category + 8-bit operation. */
object OpCode {
    // Core Session Operations
    const val NOP               = 0x0000
    const val KEEPALIVE         = 0x0001
    const val KEEPALIVE_ACK     = 0x0002
    const val SESSION_INIT      = 0x0003
    const val SESSION_ACK       = 0x0004
    const val SESSION_CLOSE     = 0x0005
    const val SESSION_CLOSE_ACK = 0x0006
    const val SESSION_RESUME    = 0x0007
    const val SESSION_RESUMED   = 0x0008

    // Key Management Operations
    const val KEY_EXCHANGE_INIT     = 0x0010
    const val KEY_EXCHANGE_RESPONSE = 0x0011
    const val KEY_EXCHANGE_COMPLETE = 0x0012
    const val SESSION_ROTATE        = 0x0016
    const val SESSION_REVOKE        = 0x0017

    // Error
    const val ERROR = 0x00FF

    // ArcHub Admin (0x20)
    const val ADMIN_LIST_SERVICES   = 0x2001
    const val ADMIN_START_SERVICE   = 0x2002
    const val ADMIN_STOP_SERVICE    = 0x2003
    const val ADMIN_ENABLE_SERVICE  = 0x2004
    const val ADMIN_DISABLE_SERVICE = 0x2005
    const val ADMIN_BACKUP          = 0x2006
    const val ADMIN_RESTORE         = 0x2007
    const val ADMIN_SERVICE_STATUS  = 0x2010
    const val ADMIN_CREATE_PROFILE  = 0x2011
    const val ADMIN_DELETE_PROFILE  = 0x2012
    const val ADMIN_GET_PROFILE     = 0x2013
    const val ADMIN_CHECK_PORT      = 0x2014
    const val ADMIN_SERVICE_DETAILS = 0x2015
    const val ADMIN_SERVICE_LOGS    = 0x2016
    const val ADMIN_OPEN_FIREWALL   = 0x2017

    // ArcHub Schema Sync (0x21)
    const val SCHEMA_LIST     = 0x2100
    const val SCHEMA_GET      = 0x2101
    const val SCHEMA_UPLOAD   = 0x2102
    const val SCHEMA_DELETE   = 0x2103
    const val SCHEMA_SEARCH   = 0x2104
    const val SCHEMA_VERSIONS = 0x2105

    // ArcHub Wizard Sync (0x21)
    const val WIZARD_LIST   = 0x2110
    const val WIZARD_GET    = 0x2111
    const val WIZARD_UPLOAD = 0x2112
    const val WIZARD_DELETE = 0x2113
    const val WIZARD_SEARCH = 0x2114

    // ArcHub Profile Sync (0x21)
    const val PROFILE_LIST        = 0x2120
    const val PROFILE_GET         = 0x2121
    const val PROFILE_UPLOAD      = 0x2122
    const val PROFILE_DELETE      = 0x2123
    const val PROFILE_SEARCH      = 0x2124
    const val PROFILE_RATE        = 0x2125
    const val PROFILE_GET_RATINGS = 0x2126
    const val PROFILE_DOWNLOAD    = 0x2127

    // ArcHub Community (0x21)
    const val TAG_ADD          = 0x2140
    const val TAG_REMOVE       = 0x2141
    const val TAG_LIST         = 0x2142
    const val TAG_SEARCH       = 0x2143
    const val COMMUNITY_SEARCH = 0x2150

    // Market Data — SilverWidget (0x30)
    const val MARKET_SUBSCRIBE        = 0x3000
    const val MARKET_PRICE_TICK       = 0x3001
    const val MARKET_UNSUBSCRIBE      = 0x3002
    const val MARKET_GET_CANDLES      = 0x3010
    const val MARKET_CANDLES_DATA     = 0x3011
    const val MARKET_SYNC_SETTINGS    = 0x3020
    const val MARKET_SETTINGS_DATA    = 0x3021
    const val MARKET_UPDATE_SETTING   = 0x3022
    const val MARKET_GET_SOURCES      = 0x3030
    const val MARKET_SOURCES_LIST     = 0x3031
    const val MARKET_GET_INSTRUMENTS  = 0x3040
    const val MARKET_INSTRUMENTS_LIST = 0x3041
    const val MARKET_GET_CANDLES_BATCH  = 0x3050
    const val MARKET_CANDLES_BATCH_DATA = 0x3051
    const val MARKET_SEARCH_INSTRUMENTS = 0x3060
    const val MARKET_SEARCH_RESULTS     = 0x3061
    const val MARKET_ADD_INSTRUMENT     = 0x3062
    const val MARKET_DELETE_INSTRUMENT  = 0x3063

    private val names = mapOf(
        NOP to "NOP", KEEPALIVE to "KEEPALIVE", KEEPALIVE_ACK to "KEEPALIVE_ACK",
        SESSION_INIT to "SESSION_INIT", SESSION_ACK to "SESSION_ACK",
        SESSION_CLOSE to "SESSION_CLOSE", SESSION_CLOSE_ACK to "SESSION_CLOSE_ACK",
        SESSION_RESUME to "SESSION_RESUME", SESSION_RESUMED to "SESSION_RESUMED",
        KEY_EXCHANGE_INIT to "KEY_EXCHANGE_INIT", KEY_EXCHANGE_RESPONSE to "KEY_EXCHANGE_RESPONSE",
        KEY_EXCHANGE_COMPLETE to "KEY_EXCHANGE_COMPLETE",
        SESSION_ROTATE to "SESSION_ROTATE", SESSION_REVOKE to "SESSION_REVOKE",
        ERROR to "ERROR",
        MARKET_SUBSCRIBE to "MARKET_SUBSCRIBE", MARKET_PRICE_TICK to "MARKET_PRICE_TICK",
        MARKET_UNSUBSCRIBE to "MARKET_UNSUBSCRIBE",
        MARKET_GET_CANDLES to "MARKET_GET_CANDLES", MARKET_CANDLES_DATA to "MARKET_CANDLES_DATA",
        MARKET_SYNC_SETTINGS to "MARKET_SYNC_SETTINGS", MARKET_SETTINGS_DATA to "MARKET_SETTINGS_DATA",
        MARKET_UPDATE_SETTING to "MARKET_UPDATE_SETTING",
        MARKET_GET_SOURCES to "MARKET_GET_SOURCES", MARKET_SOURCES_LIST to "MARKET_SOURCES_LIST",
        MARKET_GET_INSTRUMENTS to "MARKET_GET_INSTRUMENTS", MARKET_INSTRUMENTS_LIST to "MARKET_INSTRUMENTS_LIST",
        MARKET_GET_CANDLES_BATCH to "MARKET_GET_CANDLES_BATCH", MARKET_CANDLES_BATCH_DATA to "MARKET_CANDLES_BATCH_DATA",
        MARKET_SEARCH_INSTRUMENTS to "MARKET_SEARCH_INSTRUMENTS", MARKET_SEARCH_RESULTS to "MARKET_SEARCH_RESULTS",
        MARKET_ADD_INSTRUMENT to "MARKET_ADD_INSTRUMENT", MARKET_DELETE_INSTRUMENT to "MARKET_DELETE_INSTRUMENT",
    )

    /** Look up the name for an opcode. Returns empty string if unknown. */
    fun lookup(code: Int): String = names[code] ?: ""

    /** Extract category byte (high byte). */
    fun category(code: Int): Int = (code shr 8) and 0xFF

    /** Extract operation byte (low byte). */
    fun operation(code: Int): Int = code and 0xFF
}
