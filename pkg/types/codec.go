package types

type Protocols interface {
	// return 1. stream id if have one 2. headers bytes
	EncodeHeaders(headers interface{}) (uint32, IoBuffer)

	EncodeData(data IoBuffer) IoBuffer

	EncodeTrailers(trailers map[string]string) IoBuffer

	Decode(data IoBuffer, filter DecodeFilter)
}

type Encoder interface {
	// return 1. stream id if have one 2. headers bytes
	EncodeHeaders(headers interface{}) (uint32, IoBuffer)

	EncodeData(data IoBuffer) IoBuffer

	EncodeTrailers(trailers map[string]string) IoBuffer
}

type Decoder interface {
	// return 1. bytes decoded 2. decoded cmd
	Decode(data IoBuffer) (int, interface{})
}

type DecodeFilter interface {
	OnDecodeHeader(streamId uint32, headers map[string]string) FilterStatus

	OnDecodeData(streamId uint32, data IoBuffer) FilterStatus

	OnDecodeTrailer(streamId uint32, trailers map[string]string) FilterStatus
}

type DecoderCallbacks interface {
	OnResponseValue(value interface{})
}

type DecoderFactory interface {
	Create(cb DecoderCallbacks) Decoder
}
