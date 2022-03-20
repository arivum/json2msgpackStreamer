package json2msgpackStreamer

func (t *JSON2MsgPackStreamer) handleArray() error {
	var (
		err              error
		numValues        = 0
		typePos, sizePos *blockBufPos
	)

	typePos = t.buf.getCurrentPos()
	t.buf.appendByte(msgPackFlagArray16)
	sizePos = t.buf.getCurrentPos()

	for {
		if t.nextByte, err = t.r.ReadByte(); err != nil {
			return err
		}

		switch t.nextByte {
		case ' ', '\n', '\r', ',', '\t':
			// Do nothing
		case '"':
			if err = t.handleString(); err != nil {
				return err
			}
			numValues++
		case '{':
			if err = t.handleStruct(); err != nil {
				return err
			}
			numValues++
		case '[':
			if err = t.handleArray(); err != nil {
				return err
			}
			numValues++
		case ']':
			switch {
			case numValues < max4BitPlusOne:
				t.buf.writeByteToOffset(msgPackFlagFixArray|byte(numValues), typePos)
			case numValues < max16BitPlusOne:
				t.buf.insertOnOffset(lenTo2Bytes(numValues), sizePos)
			case numValues < max32BitPlusOne:
				t.buf.writeByteToOffset(msgPackFlagArray32, typePos)
				t.buf.insertOnOffset(lenTo4Bytes(numValues), sizePos)
			}
			return nil
		default:
			t.r.UnreadByte()
			if err = t.handleAtomic(); err != nil {
				return err
			}
			numValues++
		}
	}
}

func (t *JSON2MsgPackStreamer) handleStruct() error {
	var (
		err              error
		numKeys          = 0
		typePos, sizePos *blockBufPos
	)

	typePos = t.buf.getCurrentPos()
	t.buf.appendByte(msgPackFlagMap16)
	sizePos = t.buf.getCurrentPos()

	for {
		if t.nextByte, err = t.r.ReadByte(); err != nil {
			return err
		}

		switch t.nextByte {
		case ' ', '\n', '\r', ',', '\t':
			// Do nothing
		case ':':
			// value start
			numKeys++
		case '"':
			if err = t.handleString(); err != nil {
				return err
			}
		case '{':
			if err = t.handleStruct(); err != nil {
				return err
			}
		case '[':
			if err = t.handleArray(); err != nil {
				return err
			}
		case '}':
			switch {
			case numKeys < max4BitPlusOne:
				t.buf.writeByteToOffset(msgPackFlagFixMap|byte(numKeys), typePos)
			case numKeys < max16BitPlusOne:
				t.buf.insertOnOffset(lenTo2Bytes(numKeys), sizePos)
			case numKeys < max32BitPlusOne:
				t.buf.writeByteToOffset(msgPackFlagMap32, typePos)
				t.buf.insertOnOffset(lenTo4Bytes(numKeys), sizePos)
			}
			return nil
		default:
			t.r.UnreadByte()
			if err = t.handleAtomic(); err != nil {
				return err
			}
		}
	}
}
