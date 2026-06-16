// MyClerk opcode catalog (canonical reference in arclib; apps copy locally, see cmd/opcode-verify).
// Migrated MyClerk maya operation codes (category 0x10).

package protocol

// Maya operations (category 0x10).
const (
	OpMayaQuery                OpCode = 0x1000 // Send query
	OpMayaResponse             OpCode = 0x1001 // Query response
	OpMayaStream               OpCode = 0x1002 // Streaming response
	OpMayaCommand              OpCode = 0x1010 // Execute command
	OpMayaResult               OpCode = 0x1011 // Command result
	OpMayaContext              OpCode = 0x1020 // Set context
	OpMayaMemory               OpCode = 0x1021 // Memory operation
	OpMayaPersonality          OpCode = 0x1022 // Set personality
	OpMayaVoice                OpCode = 0x1030 // Voice input
	OpMayaSpeak                OpCode = 0x1031 // Voice output
	OpMayaTTS                  OpCode = 0x1032 // Text-to-speech
	OpMayaSTT                  OpCode = 0x1033 // Speech-to-text
	OpMayaVoiceCatalog         OpCode = 0x1034 // Voice catalog management
	OpMayaLearn                OpCode = 0x1040 // Learn interaction
	OpMayaSuggest              OpCode = 0x1041 // Get suggestions
	OpMayaConversationList     OpCode = 0x1050 // List conversations
	OpMayaConversationGet      OpCode = 0x1051 // Get conversation
	OpMayaConversationDelete   OpCode = 0x1052 // Delete conversation
	OpMayaMessagesGet          OpCode = 0x1053 // Get conversation messages
	OpMayaModelsGet            OpCode = 0x1060 // Get available models
	OpMayaHealthGet            OpCode = 0x1061 // Get provider health
	OpMayaUsageGet             OpCode = 0x1062 // Get usage statistics
	OpMayaBudgetGet            OpCode = 0x1063 // Get budget for family
	OpMayaBudgetSet            OpCode = 0x1064 // Set budget for family
	OpMayaRecordingStart       OpCode = 0x1070 // Start story recording
	OpMayaRecordingStop        OpCode = 0x1071 // Stop story recording
	OpMayaRecordingPause       OpCode = 0x1072 // Pause story recording
	OpMayaRecordingResume      OpCode = 0x1073 // Resume story recording
	OpMayaRecordingGet         OpCode = 0x1074 // Get story recording
	OpMayaRecordingList        OpCode = 0x1075 // List story recordings
	OpMayaRecordingTranscribe  OpCode = 0x1076 // Transcribe story recording
	OpMayaRecordingAppendAudio OpCode = 0x1077 // Append audio to recording
	OpMayaTranscriptList       OpCode = 0x1078 // List voice transcripts
	OpMayaTranscriptGet        OpCode = 0x1079 // Get a voice transcript
	OpMayaTranscriptDelete     OpCode = 0x107A // Delete a voice transcript
	OpMayaTranscriptExport     OpCode = 0x107B // Export user transcripts (DSGVO)
	OpVoiceEnrollBegin         OpCode = 0x1080 // Start or resume voice enrollment
	OpVoiceEnrollSample        OpCode = 0x1081 // Submit one enrollment audio sample
	OpVoiceEnrollComplete      OpCode = 0x1082 // Train embedding from collected samples
	OpVoiceVerifyBegin         OpCode = 0x1083 // Request verification challenge nonce
	OpVoiceVerify              OpCode = 0x1084 // Verify a live audio clip against stored embedding
	OpVoiceDeleteEnrollment    OpCode = 0x1085 // GDPR hard delete of enrollment + embedding
	OpVoiceExtractEmbedding    OpCode = 0x1086
)
