package wrp

import (
	"github.com/Comcast/webpa-common/tracing"
)

// Entity describes a single WRP message decoded from some external source, such as an HTTP request.
// This type implements Routable and can optionally be associated with one or more spans.
type Entity struct {
	format   Format
	contents []byte
	message  Message
	spans    []tracing.Span
}

func (e *Entity) MessageType() MessageType {
	return e.message.Type
}

func (e *Entity) To() string {
	return e.message.Destination
}

func (e *Entity) From() string {
	return e.message.Source
}

func (e *Entity) TransactionKey() string {
	return e.message.TransactionUUID
}

func (e *Entity) Spans() []tracing.Span {
	return e.spans
}

func (e *Entity) WithSpans(spans ...tracing.Span) interface{} {
	copy := *e
	copy.spans = spans
	return &copy
}

// DecodeEntityBytes decodes an entity from the given byte slice.  The original
// formatted contents are preserved and associted with the returned Entity.
func DecodeEntityBytes(f Format, c []byte) (*Entity, error) {
	e := &Entity{
		format:   f,
		contents: c,
	}

	if err := NewDecoderBytes(c, f).Decode(&e.message); err != nil {
		return nil, err
	}

	return e, nil
}

// EncodeEntityBytes encodes the given entity into a byte slice.  If the entity's contents
// are already formatted appropriately, this function is essentially a no-op and will set the
// given bytes slice to the internal contents buffer.
func EncodeEntityBytes(e *Entity, f Format, o *[]byte) error {
	if e.format == f && len(e.contents) > 0 {
		*o = e.contents
		return nil
	}

	return NewEncoderBytes(o, f).Encode(&e.message)
}
