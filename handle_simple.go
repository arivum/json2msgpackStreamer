package json2msgpackStreamer

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"unicode"
)

func (t *JSON2MsgPackStreamer) handleString() error {
	var (
		err        error
		subStr     []byte
		strLen     = 0
		subStrLen  int
		EOFString  = true
		curPos     *blockBufPos
		round      = 0
		moreSubstr []byte
		tmp        []byte
	)

	for {
		if subStr, err = t.r.ReadSlice('"'); err != nil && err != bufio.ErrBufferFull {
			return err
		} else if err == bufio.ErrBufferFull {
			tmp = make([]byte, len(subStr))
			copy(tmp, subStr)
			if moreSubstr, err = t.r.ReadBytes('"'); err != nil {
				return err
			}
			subStr = append(tmp, moreSubstr...)
		}

		subStrLen = len(subStr)
		if subStrLen <= 1 || subStr[subStrLen-2] != '\\' {
			subStr = subStr[:subStrLen-1]
			EOFString = true
		}

		strLen += len(subStr)

		if EOFString {
			if round == 0 {
				switch {
				case strLen < max5BitPlusOne:
					t.buf.appendByte(msgPackFlagFixStr | byte(strLen))
				case strLen < max8BitPlusOne:
					t.buf.appendByte(msgPackFlagStr8)
					t.buf.appendByte(byte(strLen))
				case strLen < max16BitPlusOne:
					t.buf.appendByte(msgPackFlagStr16)
					t.buf.append(lenTo2Bytes(strLen))
				case strLen < max32BitPlusOne:
					t.buf.appendByte(msgPackFlagStr32)
					t.buf.append(lenTo4Bytes(strLen))
				default:
					return errors.New("string sizes > 2^32 not allowed")
				}
				t.buf.append(subStr)
				return nil
			} else {
				t.buf.writeToOffset(lenTo4Bytes(strLen), curPos)
				return nil
			}
		} else {
			if round == 0 {
				t.buf.appendByte(msgPackFlagStr32)
				curPos = t.buf.getCurrentPos()
				t.buf.append(placeholder32Bit)
			}
			t.buf.append(subStr)
		}
		round = 1
	}
}

func (t *JSON2MsgPackStreamer) handleAtomic() error {
	var (
		err      error
		valueStr = make([]byte, 20)
		EOF      = false
		index    = 0
		tmp      []byte
		buf      = bytes.NewBuffer(nil)
	)

	for {
		if t.nextByte, err = t.r.ReadByte(); err != nil && err != io.EOF {
			return err
		} else if err == io.EOF {
			EOF = true
		} else {
			switch t.nextByte {
			case '\n', ' ', '\r':
				// do nothing
			case ',', ']', '}':
				t.r.UnreadByte()
				EOF = true
			default:
				if cap(valueStr) <= index+1 {
					tmp = make([]byte, cap(valueStr)<<1)
					copy(tmp, valueStr[:index])
					valueStr = tmp
				}
				valueStr[index] = t.nextByte
				index++
			}
		}

		if EOF {
			valueStr = valueStr[:index]
			break
		}
	}

	if len(valueStr) == 0 {
		return errors.New("couldn't parse json")
	}

	switch {
	case unicode.IsDigit(rune(valueStr[0])):
		if bytes.Contains(valueStr, valueDot) {
			var value float64
			if value, err = strconv.ParseFloat(string(valueStr), 64); err != nil {
				return err
			}
			if float64(float32(value)) == value {
				if err = binary.Write(buf, binary.BigEndian, float32(value)); err != nil {
					return err
				}
				t.buf.appendByte(msgPackFlagFloat32)
				t.buf.append(buf.Bytes())
				return nil
			} else {
				if err = binary.Write(buf, binary.BigEndian, value); err != nil {
					return err
				}
				t.buf.appendByte(msgPackFlagFloat64)
				t.buf.append(buf.Bytes())
				return nil
			}
		} else {
			var (
				value uint64
			)
			if value, err = strconv.ParseUint(string(valueStr), 10, 64); err != nil {
				return err
			}
			switch {
			case value < (1 << 7):
				t.buf.appendByte(msgPackFlagPosFixInt | byte(value))
				return nil
			case value <= math.MaxUint8:
				t.buf.appendByte(msgPackFlagUint8)
				t.buf.appendByte(byte(value))
				return nil
			case value <= math.MaxInt16:
				binary.BigEndian.PutUint16(buf2Bytes, uint16(value))
				t.buf.appendByte(msgPackFlagUint16)
				t.buf.append(buf2Bytes)
				return nil
			case value <= math.MaxInt32:
				binary.BigEndian.PutUint32(buf4Bytes, uint32(value))
				t.buf.appendByte(msgPackFlagUint32)
				t.buf.append(buf4Bytes)
				return nil
			default:
				binary.BigEndian.PutUint64(buf8Bytes, uint64(value))
				t.buf.appendByte(msgPackFlagUint64)
				t.buf.append(buf8Bytes)
				return nil
			}
		}
	case valueStr[0] == '-':
		if bytes.Contains(valueStr, valueDot) {
			var value float64
			if value, err = strconv.ParseFloat(string(valueStr), 64); err != nil {
				return err
			}
			if float64(float32(value)) == value {
				if err = binary.Write(buf, binary.BigEndian, float32(value)); err != nil {
					return err
				}
				t.buf.appendByte(msgPackFlagFloat32)
				t.buf.append(buf.Bytes())
				return nil
			} else {
				if err = binary.Write(buf, binary.BigEndian, value); err != nil {
					return err
				}
				t.buf.appendByte(msgPackFlagFloat64)
				t.buf.append(buf.Bytes())
				return nil
			}
		} else {
			var value int64
			if value, err = strconv.ParseInt(string(valueStr), 10, 64); err != nil {
				return err
			}
			switch {
			case value > -(1 << 5):
				t.buf.appendByte(msgPackFlagNegFixInt | byte(value))
				return nil
			case value >= math.MinInt8:
				if err = binary.Write(buf, binary.BigEndian, int8(value)); err != nil {
					return err
				}
				t.buf.appendByte(msgPackFlagInt8)
				t.buf.append(buf.Bytes())
				return nil
			case value >= math.MinInt16:
				if err = binary.Write(buf, binary.BigEndian, int16(value)); err != nil {
					return err
				}
				t.buf.appendByte(msgPackFlagInt16)
				t.buf.append(buf.Bytes())
				return nil
			case value >= math.MinInt32:
				if err = binary.Write(buf, binary.BigEndian, int32(value)); err != nil {
					return err
				}
				t.buf.appendByte(msgPackFlagInt32)
				t.buf.append(buf.Bytes())
				return nil
			default:
				if err = binary.Write(buf, binary.BigEndian, value); err != nil {
					return err
				}
				t.buf.appendByte(msgPackFlagInt64)
				t.buf.append(buf.Bytes())
				return nil
			}
		}
	case valueStr[0] == 'n':
		if bytes.Equal(valueStr, valueNil) {
			t.buf.appendByte(msgPackFlagNil)
			return nil
		}
		return fmt.Errorf("failed to parse json on token %s", string(valueStr))
	case valueStr[0] == 'f':
		if bytes.Equal(valueStr, valueFalse) {
			t.buf.appendByte(msgPackFlagBoolFalse)
			return nil
		}
		return fmt.Errorf("failed to parse json on token %s", string(valueStr))
	case valueStr[0] == 't':
		if bytes.Equal(valueStr, valueTrue) {
			t.buf.appendByte(msgPackFlagBoolTrue)
			return nil
		}
		return fmt.Errorf("failed to parse json on token %s", string(valueStr))
	default:
		return fmt.Errorf("failed to parse json on token %s", string(valueStr))
	}
}
