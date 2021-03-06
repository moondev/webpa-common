package wrp

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var capacities = []int{-1, 0, 2, 10, 50}

func testEncoderPoolFormat(t *testing.T, ep *EncoderPool) {
	assert := assert.New(t)

	assert.True(ep.Format() >= 0)
	assert.True(ep.Format() < lastFormat)
}

func testEncoderPoolPutGet(t *testing.T, ep *EncoderPool) {
	var (
		assert  = assert.New(t)
		require = require.New(t)
	)

	require.Zero(ep.Len())
	require.True(ep.Cap() > 0)

	assert.NotNil(ep.Get())
	assert.Zero(ep.Len())
	assert.True(ep.Cap() > 0)

	for ep.Len() < ep.Cap() {
		assert.True(ep.Put(ep.New()))
	}

	assert.False(ep.Put(ep.New()))

	for ep.Len() > 0 {
		assert.NotNil(ep.Get())
	}

	assert.True(ep.Put(ep.New()))
}

func testEncoderPoolEncode(t *testing.T, ep *EncoderPool, dp *DecoderPool) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		input  = &Message{Payload: []byte("hi!"), Source: "test"}
		output = new(bytes.Buffer)

		decoded = new(Message)
	)

	require.NoError(ep.Encode(output, input))
	assert.NoError(dp.Decode(decoded, output))

	assert.Equal(*input, *decoded)
}

func testEncoderPoolEncodeBytes(t *testing.T, ep *EncoderPool, dp *DecoderPool) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		input  = &Message{Payload: []byte("hi!"), Source: "test"}
		output []byte

		decoded = new(Message)
	)

	require.NoError(ep.EncodeBytes(&output, input))
	assert.NoError(dp.DecodeBytes(decoded, output))

	assert.Equal(*input, *decoded)
}

func TestEncoderPool(t *testing.T) {
	for f := Format(0); f < lastFormat; f++ {
		t.Run(f.String(), func(t *testing.T) {
			for _, c := range capacities {
				t.Run(fmt.Sprintf("Capacity:%d", c), func(t *testing.T) {
					t.Run("Format", func(t *testing.T) {
						testEncoderPoolFormat(t, NewEncoderPool(c, f))
					})

					t.Run("PutGet", func(t *testing.T) {
						testEncoderPoolPutGet(t, NewEncoderPool(c, f))
					})

					t.Run("Encode", func(t *testing.T) {
						testEncoderPoolEncode(t, NewEncoderPool(c, f), NewDecoderPool(c, f))
					})

					t.Run("EncodeBytes", func(t *testing.T) {
						testEncoderPoolEncodeBytes(t, NewEncoderPool(c, f), NewDecoderPool(c, f))
					})
				})
			}
		})
	}
}

func testDecoderPoolFormat(t *testing.T, dp *DecoderPool) {
	assert := assert.New(t)

	assert.True(dp.Format() >= 0)
	assert.True(dp.Format() < lastFormat)
}

func testDecoderPoolPutGet(t *testing.T, dp *DecoderPool) {
	var (
		assert  = assert.New(t)
		require = require.New(t)
	)

	require.Zero(dp.Len())
	require.True(dp.Cap() > 0)

	assert.NotNil(dp.Get())
	assert.Zero(dp.Len())
	assert.True(dp.Cap() > 0)

	for dp.Len() < dp.Cap() {
		assert.True(dp.Put(dp.New()))
	}

	assert.False(dp.Put(dp.New()))

	for dp.Len() > 0 {
		assert.NotNil(dp.Get())
	}

	assert.True(dp.Put(dp.New()))
}

func TestDecoderPool(t *testing.T) {
	for f := Format(0); f < lastFormat; f++ {
		t.Run(f.String(), func(t *testing.T) {
			for _, c := range capacities {
				t.Run(fmt.Sprintf("Capacity:%d", c), func(t *testing.T) {
					t.Run("Format", func(t *testing.T) {
						testDecoderPoolFormat(t, NewDecoderPool(c, f))
					})

					t.Run("PutGet", func(t *testing.T) {
						testDecoderPoolPutGet(t, NewDecoderPool(c, f))
					})
				})
			}
		})
	}
}
func BenchmarkWRP(b *testing.B) {
	var (
		require = require.New(b)
		payload = make([]byte, 1024)
	)

	rand.Read(payload)

	var (
		status                  int64 = 200
		requestDeliveryResponse int64 = 1

		message = &Message{
			Type:                    SimpleRequestResponseMessageType,
			Source:                  "test",
			Destination:             "mac:123412341234",
			Status:                  &status,
			RequestDeliveryResponse: &requestDeliveryResponse,
			Accept:                  "application/json",
			ContentType:             "application/octet-stream",
			Payload:                 payload,
		}

		encoded      = make(map[Format][]byte)
		decoderPools = make(map[Format]*DecoderPool)
		encoderPools = make(map[Format]*EncoderPool)
	)

	for _, f := range AllFormats() {
		var (
			output  []byte
			encoder = NewEncoderBytes(&output, f)
		)

		require.NoError(encoder.Encode(message))
		encoded[f] = output
		decoderPools[f] = NewDecoderPool(10, f)
		encoderPools[f] = NewEncoderPool(10, f)
	}

	b.ResetTimer()
	for _, f := range AllFormats() {
		b.Run(f.String(), func(b *testing.B) {
			b.Run("EncoderPool", func(b *testing.B) {
				benchmarkEncoderPool(b, encoderPools[f], message)
			})

			b.Run("DecoderPool", func(b *testing.B) {
				benchmarkDecoderPool(b, decoderPools[f], encoded[f])
			})

			b.Run("Encoder", func(b *testing.B) {
				benchmarkEncoder(b, f, message)
			})

			b.Run("Decoder", func(b *testing.B) {
				benchmarkDecoder(b, f, encoded[f])
			})
		})
	}
}

func benchmarkEncoderPool(b *testing.B, pool *EncoderPool, message *Message) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var output []byte
			if err := pool.EncodeBytes(&output, message); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func benchmarkDecoderPool(b *testing.B, pool *DecoderPool, data []byte) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var message Message
			if err := pool.DecodeBytes(&message, data); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func benchmarkEncoder(b *testing.B, format Format, message *Message) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var (
				output  []byte
				encoder = NewEncoderBytes(&output, format)
			)

			if err := encoder.Encode(message); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func benchmarkDecoder(b *testing.B, format Format, data []byte) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var (
				message Message
				decoder = NewDecoderBytes(data, format)
			)

			if err := decoder.Decode(&message); err != nil {
				b.Fatal(err)
			}
		}
	})
}
