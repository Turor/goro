package db

import (
	"encoding/binary"
	"fmt"
	"math"
)

const bsonMaxDocument = 16 << 20

const (
	bsonTypeDouble    byte = 0x01
	bsonTypeString    byte = 0x02
	bsonTypeDocument  byte = 0x03
	bsonTypeArray     byte = 0x04
	bsonTypeBinary    byte = 0x05
	bsonTypeUndefined byte = 0x06
	bsonTypeObjectID  byte = 0x07
	bsonTypeBool      byte = 0x08
	bsonTypeUTC       byte = 0x09
	bsonTypeNull      byte = 0x0A
	bsonTypeRegexp    byte = 0x0B
	bsonTypeInt32     byte = 0x10
	bsonTypeTimestamp byte = 0x11
	bsonTypeInt64     byte = 0x12
)

// DecodeBSONDocuments reads concatenated BSON documents, matching
// roBrowserLegacy loadBSONFile.
func DecodeBSONDocuments(data []byte) ([]map[string]any, error) {
	var docs []map[string]any
	offset := 0
	for offset < len(data) {
		if offset+4 > len(data) {
			return nil, fmt.Errorf("bson: truncated size at %d", offset)
		}
		size := int(int32(binary.LittleEndian.Uint32(data[offset:])))
		if size < 5 || size > bsonMaxDocument || offset+size > len(data) {
			return nil, fmt.Errorf("bson: invalid document size %d at %d", size, offset)
		}
		doc, err := decodeBSONDocument(data[offset : offset+size])
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
		offset += size
	}
	return docs, nil
}

func decodeBSONDocument(data []byte) (map[string]any, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("bson: document too small")
	}
	size := int(int32(binary.LittleEndian.Uint32(data)))
	if size != len(data) || data[len(data)-1] != 0 {
		return nil, fmt.Errorf("bson: corrupt document")
	}
	out := map[string]any{}
	order := make([]string, 0, 8)
	offset := 4
	for offset < len(data)-1 {
		typ := data[offset]
		offset++
		name, n, err := bsonCString(data[offset:])
		if err != nil {
			return nil, err
		}
		offset += n
		value, rest, err := decodeBSONValue(typ, data[offset:])
		if err != nil {
			return nil, fmt.Errorf("bson field %q: %w", name, err)
		}
		offset = len(data) - len(rest)
		out[name] = value
		order = append(order, name)
	}
	out[""] = order
	return out, nil
}

func decodeBSONValue(typ byte, data []byte) (any, []byte, error) {
	switch typ {
	case bsonTypeDouble:
		if len(data) < 8 {
			return nil, data, fmt.Errorf("truncated double")
		}
		u := binary.LittleEndian.Uint64(data)
		return math.Float64frombits(u), data[8:], nil
	case bsonTypeString:
		if len(data) < 4 {
			return nil, data, fmt.Errorf("truncated string size")
		}
		size := int(int32(binary.LittleEndian.Uint32(data)))
		if size < 1 || 4+size > len(data) {
			return nil, data, fmt.Errorf("invalid string size %d", size)
		}
		if data[4+size-1] != 0 {
			return nil, data, fmt.Errorf("string missing NUL")
		}
		return string(data[4 : 4+size-1]), data[4+size:], nil
	case bsonTypeDocument, bsonTypeArray:
		if len(data) < 4 {
			return nil, data, fmt.Errorf("truncated subdocument")
		}
		size := int(int32(binary.LittleEndian.Uint32(data)))
		if size < 5 || size > len(data) {
			return nil, data, fmt.Errorf("invalid subdocument size %d", size)
		}
		doc, err := decodeBSONDocument(data[:size])
		if err != nil {
			return nil, data, err
		}
		if typ == bsonTypeArray {
			return bsonArray(doc), data[size:], nil
		}
		return doc, data[size:], nil
	case bsonTypeBool:
		if len(data) < 1 {
			return nil, data, fmt.Errorf("truncated bool")
		}
		return data[0] != 0, data[1:], nil
	case bsonTypeNull, bsonTypeUndefined:
		return nil, data, nil
	case bsonTypeInt32:
		if len(data) < 4 {
			return nil, data, fmt.Errorf("truncated int32")
		}
		return int32(binary.LittleEndian.Uint32(data)), data[4:], nil
	case bsonTypeInt64, bsonTypeUTC, bsonTypeTimestamp:
		if len(data) < 8 {
			return nil, data, fmt.Errorf("truncated int64")
		}
		return int64(binary.LittleEndian.Uint64(data)), data[8:], nil
	case bsonTypeObjectID:
		if len(data) < 12 {
			return nil, data, fmt.Errorf("truncated objectid")
		}
		return data[:12], data[12:], nil
	case bsonTypeBinary:
		if len(data) < 5 {
			return nil, data, fmt.Errorf("truncated binary")
		}
		size := int(int32(binary.LittleEndian.Uint32(data)))
		if size < 0 || 5+size > len(data) {
			return nil, data, fmt.Errorf("invalid binary size %d", size)
		}
		return data[5 : 5+size], data[5+size:], nil
	default:
		return nil, data, fmt.Errorf("unsupported type 0x%02x", typ)
	}
}

func bsonCString(data []byte) (string, int, error) {
	for i, b := range data {
		if b == 0 {
			return string(data[:i]), i + 1, nil
		}
	}
	return "", 0, fmt.Errorf("bson: unterminated cstring")
}

func bsonArray(doc map[string]any) []any {
	order, _ := doc[""].([]string)
	out := make([]any, 0, len(order))
	for _, key := range order {
		if key == "" {
			continue
		}
		out = append(out, doc[key])
	}
	return out
}

func bsonFieldOrder(doc map[string]any) []string {
	order, _ := doc[""].([]string)
	out := make([]string, 0, len(order))
	for _, key := range order {
		if key != "" {
			out = append(out, key)
		}
	}
	return out
}

// EncodeBSONDocument writes a single BSON document. Values may be string,
// bool, int, int32, int64, float64, or map[string]any. Used to build fixtures.
func EncodeBSONDocument(fields map[string]any, order []string) []byte {
	if order == nil {
		order = make([]string, 0, len(fields))
		for k := range fields {
			order = append(order, k)
		}
	}
	body := []byte{}
	for _, key := range order {
		value, ok := fields[key]
		if !ok {
			continue
		}
		body = append(body, encodeBSONField(key, value)...)
	}
	out := make([]byte, 4+len(body)+1)
	binary.LittleEndian.PutUint32(out, uint32(len(out)))
	copy(out[4:], body)
	return out
}

func encodeBSONField(name string, value any) []byte {
	switch v := value.(type) {
	case string:
		payload := make([]byte, 4+len(v)+1)
		binary.LittleEndian.PutUint32(payload, uint32(len(v)+1))
		copy(payload[4:], v)
		return bsonField(bsonTypeString, name, payload)
	case bool:
		b := byte(0)
		if v {
			b = 1
		}
		return bsonField(bsonTypeBool, name, []byte{b})
	case int:
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(int32(v)))
		return bsonField(bsonTypeInt32, name, buf)
	case int32:
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(v))
		return bsonField(bsonTypeInt32, name, buf)
	case int64:
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, uint64(v))
		return bsonField(bsonTypeInt64, name, buf)
	case float64:
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, math.Float64bits(v))
		return bsonField(bsonTypeDouble, name, buf)
	case map[string]any:
		order, _ := valueOrder(v)
		return bsonField(bsonTypeDocument, name, EncodeBSONDocument(v, order))
	default:
		return nil
	}
}

func valueOrder(v map[string]any) ([]string, bool) {
	if order, ok := v[""].([]string); ok {
		return order, true
	}
	return nil, false
}

func bsonField(typ byte, name string, payload []byte) []byte {
	out := make([]byte, 1+len(name)+1+len(payload))
	out[0] = typ
	copy(out[1:], name)
	copy(out[1+len(name)+1:], payload)
	return out
}
