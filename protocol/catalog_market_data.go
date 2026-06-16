// MyClerk opcode catalog (canonical reference in arclib; apps copy locally, see cmd/opcode-verify).
// Migrated MyClerk market data operation codes (category 0x30).

package protocol

// Market Data operations (category 0x30).
const (
	OpMarketSubscribe         OpCode = 0x3000 // Subscribe to price feeds
	OpMarketPriceTick         OpCode = 0x3001 // Real-time price push
	OpMarketUnsubscribe       OpCode = 0x3002 // Unsubscribe from feeds
	OpMarketGetCandles        OpCode = 0x3010 // Request OHLC candles
	OpMarketCandlesData       OpCode = 0x3011 // Candle data response
	OpMarketSyncSettings      OpCode = 0x3020 // Sync client settings
	OpMarketSettingsData      OpCode = 0x3021 // Settings data response
	OpMarketUpdateSetting     OpCode = 0x3022 // Update single setting
	OpMarketGetSources        OpCode = 0x3030 // List data sources
	OpMarketSourcesList       OpCode = 0x3031 // Sources list response
	OpMarketGetInstruments    OpCode = 0x3040 // List all instruments
	OpMarketInstrumentsList   OpCode = 0x3041 // Instruments list response
	OpMarketGetCandlesBatch   OpCode = 0x3050 // Batch candles request
	OpMarketCandlesBatchData  OpCode = 0x3051 // Batch candles response
	OpMarketSearchInstruments OpCode = 0x3060 // Search instruments
	OpMarketSearchResults     OpCode = 0x3061 // Search results response
	OpMarketAddInstrument     OpCode = 0x3062 // Add new instrument
	OpMarketDeleteInstrument  OpCode = 0x3063 // Delete instrument
)
