package frame

import (
	"encoding/binary"
	"errors"
	"io"
)

type Payload []byte

type StreamFrameCodec interface {
	Encode(io.Writer, Payload) error
	Decode(io.Reader) (Payload, error)
}

var ErrShortWrite = errors.New("short write")
var ErrShortRead = errors.New("short read")

type myFrameCodec struct {
}

func NewMyFrameCodec() StreamFrameCodec {
	return &myFrameCodec{}
}

func (p *myFrameCodec) Encode(writer io.Writer, payload Payload) error {
	var f = payload
	var totalLen int32 = int32(len(payload)) + 4

	err := binary.Write(writer, binary.BigEndian, &totalLen)
	if err != nil {
		return err
	}
	n, err := writer.Write(f)
	if err != nil {
		return err
	}
	if n != len(payload) {
		return ErrShortWrite
	}
	return nil
}

func (p *myFrameCodec) Decode(reader io.Reader) (Payload, error) {
	var totalLen int32
	err := binary.Read(reader, binary.BigEndian, &totalLen)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, totalLen-4)
	n, err := io.ReadFull(reader, buf)
	if err != nil {
		return nil, err
	}
	if n != int(totalLen-4) {
		return nil, ErrShortRead
	}
	return Payload(buf), nil
}
