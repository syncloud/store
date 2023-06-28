package crypto

import (
	"bytes"
	"encoding/base64"
)

func EncodeV1(data []byte) []byte {
	buf := new(bytes.Buffer)
	buf.Grow(base64.StdEncoding.EncodedLen(len(data) + 1))
	enc := base64.NewEncoder(base64.StdEncoding, buf)
	enc.Write(v1Header)
	enc.Write(data)
	enc.Close()
	flat := buf.Bytes()
	flatSize := len(flat)

	buf = new(bytes.Buffer)
	buf.Grow(flatSize + flatSize/maxEncodeLineLength + 1)
	off := 0
	for {
		endOff := off + maxEncodeLineLength
		if endOff > flatSize {
			endOff = flatSize
		}
		buf.Write(flat[off:endOff])
		off = endOff
		if off >= flatSize {
			break
		}
		buf.WriteByte('\n')
	}

	return buf.Bytes()
}
