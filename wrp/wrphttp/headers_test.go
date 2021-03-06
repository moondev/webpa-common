package wrphttp

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/Comcast/webpa-common/wrp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func testNewMessageFromHeadersSuccess(t *testing.T) {
	var (
		assert  = assert.New(t)
		require = require.New(t)

		expectedStatus                  int64 = 928
		expectedRequestDeliveryResponse int64 = 1
		expectedIncludeSpans            bool  = true

		testData = []struct {
			header   http.Header
			payload  io.Reader
			expected wrp.Message
		}{
			{
				header: http.Header{
					MessageTypeHeader: []string{"SimpleRequestResponse"},
				},
				payload: nil,
				expected: wrp.Message{
					Type: wrp.SimpleRequestResponseMessageType,
				},
			},
			{
				header: http.Header{
					MessageTypeHeader: []string{"SimpleRequestResponse"},
				},
				payload: strings.NewReader(""),
				expected: wrp.Message{
					Type: wrp.SimpleRequestResponseMessageType,
				},
			},
			{
				header: http.Header{
					MessageTypeHeader:             []string{"SimpleRequestResponse"},
					TransactionUuidHeader:         []string{"1234"},
					SourceHeader:                  []string{"test"},
					DestinationHeader:             []string{"mac:111122223333"},
					StatusHeader:                  []string{strconv.FormatInt(expectedStatus, 10)},
					RequestDeliveryResponseHeader: []string{strconv.FormatInt(expectedRequestDeliveryResponse, 10)},
					IncludeSpansHeader:            []string{strconv.FormatBool(expectedIncludeSpans)},
					SpanHeader: []string{
						"foo, bar, moo",
						"goo, gar, hoo",
					},
					AcceptHeader: []string{"application/json"},
					PathHeader:   []string{"/foo/bar"},
				},
				payload: nil,
				expected: wrp.Message{
					Type:                    wrp.SimpleRequestResponseMessageType,
					TransactionUUID:         "1234",
					Source:                  "test",
					Destination:             "mac:111122223333",
					Status:                  &expectedStatus,
					RequestDeliveryResponse: &expectedRequestDeliveryResponse,
					IncludeSpans:            &expectedIncludeSpans,
					Spans: [][]string{
						{"foo", "bar", "moo"},
						{"goo", "gar", "hoo"},
					},
					Accept: "application/json",
					Path:   "/foo/bar",
				},
			},
			{
				header: http.Header{
					MessageTypeHeader: []string{"SimpleEvent"},
					SourceHeader:      []string{"test"},
					DestinationHeader: []string{"mac:111122223333"},
					"Content-Type":    []string{"text/plain"},
				},
				payload: strings.NewReader("payload"),
				expected: wrp.Message{
					Type:        wrp.SimpleEventMessageType,
					Source:      "test",
					Destination: "mac:111122223333",
					ContentType: "text/plain",
					Payload:     []byte("payload"),
				},
			},
			{
				header: http.Header{
					MessageTypeHeader: []string{"SimpleEvent"},
					SourceHeader:      []string{"test"},
					DestinationHeader: []string{"mac:111122223333"},
				},
				payload: strings.NewReader("payload"),
				expected: wrp.Message{
					Type:        wrp.SimpleEventMessageType,
					Source:      "test",
					Destination: "mac:111122223333",
					ContentType: "application/octet-stream",
					Payload:     []byte("payload"),
				},
			},
		}
	)

	for _, record := range testData {
		t.Logf("%#v", record)
		actual, err := NewMessageFromHeaders(record.header, record.payload)
		require.NotNil(actual)
		assert.Equal(record.expected, *actual)
		assert.NoError(err)
	}
}

func testNewMessageFromHeadersBadMessageType(t *testing.T) {
	assert := assert.New(t)

	message, err := NewMessageFromHeaders(http.Header{}, nil)
	assert.Nil(message)
	assert.Error(err)

	message, err = NewMessageFromHeaders(http.Header{MessageTypeHeader: []string{"this could not possibly be a valid message type"}}, nil)
	assert.Nil(message)
	assert.Error(err)
}

func testNewMessageFromHeadersBadIntHeader(t *testing.T, headerName string) {
	assert := assert.New(t)

	message, err := NewMessageFromHeaders(
		http.Header{
			MessageTypeHeader: []string{wrp.SimpleEventMessageType.FriendlyName()},
			headerName:        []string{"this is not a valid integer"},
		},
		nil,
	)

	assert.Nil(message)
	assert.Error(err)
}

func testNewMessageFromHeadersBadBoolHeader(t *testing.T, headerName string) {
	assert := assert.New(t)

	message, err := NewMessageFromHeaders(
		http.Header{
			MessageTypeHeader: []string{wrp.SimpleEventMessageType.FriendlyName()},
			headerName:        []string{"this is not a valid boolean"},
		},
		nil,
	)

	assert.Nil(message)
	assert.Error(err)
}

func testNewMessageFromHeadersBadSpanHeader(t *testing.T) {
	assert := assert.New(t)

	message, err := NewMessageFromHeaders(
		http.Header{
			MessageTypeHeader: []string{wrp.SimpleEventMessageType.FriendlyName()},
			SpanHeader:        []string{"this is not a valid span"},
		},
		nil,
	)

	assert.Nil(message)
	assert.Error(err)
}

func testNewMessageFromHeadersBadPayload(t *testing.T) {
	var (
		assert = assert.New(t)
		reader = new(mockReadCloser)
	)

	reader.On("Read", mock.MatchedBy(func([]byte) bool { return true })).Return(0, errors.New("expected")).Once()

	message, err := NewMessageFromHeaders(
		http.Header{
			MessageTypeHeader: []string{wrp.SimpleEventMessageType.FriendlyName()},
		},
		reader,
	)

	assert.Nil(message)
	assert.Error(err)

	reader.AssertExpectations(t)
}

func TestNewMessageFromHeaders(t *testing.T) {
	t.Run("Success", testNewMessageFromHeadersSuccess)
	t.Run("BadMessageType", testNewMessageFromHeadersBadMessageType)

	t.Run("BadIntHeader", func(t *testing.T) {
		testNewMessageFromHeadersBadIntHeader(t, StatusHeader)
		testNewMessageFromHeadersBadIntHeader(t, RequestDeliveryResponseHeader)
	})

	t.Run("BadBoolHeader", func(t *testing.T) {
		testNewMessageFromHeadersBadBoolHeader(t, IncludeSpansHeader)
	})

	t.Run("BadSpanHeader", testNewMessageFromHeadersBadSpanHeader)
	t.Run("BadPayload", testNewMessageFromHeadersBadPayload)
}

func TestAddMessageHeaders(t *testing.T) {
	var (
		assert = assert.New(t)

		expectedStatus                  int64 = 123
		expectedRequestDeliveryResponse int64 = 2
		expectedIncludeSpans            bool  = true

		testData = []struct {
			message  wrp.Message
			expected http.Header
		}{
			{
				message: wrp.Message{
					Type: wrp.SimpleRequestResponseMessageType,
				},
				expected: http.Header{
					MessageTypeHeader: []string{wrp.SimpleRequestResponseMessageType.FriendlyName()},
				},
			},
			{
				message: wrp.Message{
					Type:                    wrp.SimpleRequestResponseMessageType,
					TransactionUUID:         "1-2-3-4",
					Source:                  "test",
					Destination:             "mac:112233445566",
					Status:                  &expectedStatus,
					RequestDeliveryResponse: &expectedRequestDeliveryResponse,
					IncludeSpans:            &expectedIncludeSpans,
					Spans:                   [][]string{{"foo", "bar", "graar"}},
					Accept:                  "application/json",
					Path:                    "/foo/bar",
				},
				expected: http.Header{
					MessageTypeHeader:             []string{wrp.SimpleRequestResponseMessageType.FriendlyName()},
					TransactionUuidHeader:         []string{"1-2-3-4"},
					SourceHeader:                  []string{"test"},
					DestinationHeader:             []string{"mac:112233445566"},
					StatusHeader:                  []string{strconv.FormatInt(expectedStatus, 10)},
					RequestDeliveryResponseHeader: []string{strconv.FormatInt(expectedRequestDeliveryResponse, 10)},
					IncludeSpansHeader:            []string{strconv.FormatBool(expectedIncludeSpans)},
					SpanHeader:                    []string{"foo,bar,graar"},
					AcceptHeader:                  []string{"application/json"},
					PathHeader:                    []string{"/foo/bar"},
				},
			},
		}
	)

	for _, record := range testData {
		t.Logf("%#v", record)
		actual := make(http.Header)
		AddMessageHeaders(actual, &record.message)
		assert.Equal(record.expected, actual)
	}
}

func testWriteMessagePayloadEmptyPayload(t *testing.T) {
	assert := assert.New(t)

	{
		var payload bytes.Buffer
		assert.NoError(WriteMessagePayload(nil, &payload, &wrp.Message{}))
		assert.Empty(payload.Bytes())
	}

	{
		var (
			payload bytes.Buffer
			header  http.Header
		)

		assert.NoError(WriteMessagePayload(header, &payload, &wrp.Message{}))
		assert.Empty(payload.Bytes())
		assert.Empty(header)
	}
}

func testWriteMessagePayloadNoHeader(t *testing.T) {
	var (
		assert  = assert.New(t)
		payload bytes.Buffer
	)

	assert.NoError(WriteMessagePayload(nil, &payload, &wrp.Message{Payload: []byte("payload")}))
	assert.Equal("payload", payload.String())
}

func testWriteMessagePayloadWithHeader(t *testing.T) {
	assert := assert.New(t)

	{
		var (
			header  = make(http.Header)
			payload bytes.Buffer
			message = wrp.Message{
				Payload:     []byte("this is json, no really"),
				ContentType: "application/json",
			}
		)

		assert.NoError(WriteMessagePayload(header, &payload, &message))
		assert.Equal("application/json", header.Get("Content-Type"))
		assert.Equal("this is json, no really", payload.String())
	}

	{
		var (
			header  = make(http.Header)
			payload bytes.Buffer
			message = wrp.Message{
				Payload: []byte("this is binary, honest"),
			}
		)

		assert.NoError(WriteMessagePayload(header, &payload, &message))
		assert.Equal("application/octet-stream", header.Get("Content-Type"))
		assert.Equal("this is binary, honest", payload.String())
	}
}

func TestWriteMessagePayload(t *testing.T) {
	t.Run("EmptyPayload", testWriteMessagePayloadEmptyPayload)
	t.Run("NoHeader", testWriteMessagePayloadNoHeader)
	t.Run("WithHeader", testWriteMessagePayloadWithHeader)
}
