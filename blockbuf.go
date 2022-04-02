/*
 * Copyright (c) 2022, arivum.
 * All rights reserved.
 * SPDX-License-Identifier: MIT
 * For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/MIT
 */

package json2msgpackStreamer

import (
	"io"
	"sort"
)

func newBlockBuf() *blockbuf {
	var first = newBlock()

	return &blockbuf{
		first:   first,
		current: first,
	}
}

func (b *blockbuf) reset() {
	b.first.index = 0
	b.first.next = nil
}

func (b *blockbuf) getCurrentPos() *blockBufPos {
	return &blockBufPos{
		block: b.current,
		index: b.current.index,
	}
}

func (b *blockbuf) append(data []byte) {
	var (
		dLen     = len(data)
		lower    = 0
		copylen  = 0
		newblock *block
	)

	for dLen > 0 {
		if b.current.index+dLen < bufsize {
			copylen = dLen
		} else {
			copylen = bufsize - b.current.index
		}
		copy(b.current.buf[b.current.index:], data[lower:lower+copylen])
		b.current.index += copylen
		lower += copylen
		dLen -= copylen
		if b.current.index == bufsize {
			newblock = newBlock()
			b.current.next = newblock
			b.current = newblock
		}
	}
}

func (b *blockbuf) appendByte(data byte) {
	b.current.buf[b.current.index] = data
	b.current.index++
	if b.current.index == bufsize {
		newblock := newBlock()
		b.current.next = newblock
		b.current = newblock
	}
}

func (b *blockbuf) writeToOffset(data []byte, pos *blockBufPos) {
	var (
		dLen    = len(data)
		lower   = 0
		copylen = 0
	)

	for dLen > 0 {
		if pos.index+dLen < bufsize {
			copylen = dLen
		} else {
			copylen = bufsize - pos.index
		}
		copy(pos.block.buf[pos.index:], data[lower:lower+copylen])
		pos.index += copylen
		lower += copylen
		dLen -= copylen
		if pos.index == bufsize {
			pos.block = pos.block.next
			pos.index = 0
		}
	}
}

func (b *blockbuf) writeByteToOffset(data byte, pos *blockBufPos) {
	pos.block.buf[pos.index] = data
}

func (b *blockbuf) flushToWriter(w io.Writer) error {
	var (
		iter   = b.first
		err    error
		insert *blockInsertion
		index  int
	)

	for iter != nil {
		index = 0
		for _, insert = range iter.insertions {
			if _, err = w.Write(iter.buf[index:insert.pos]); err != nil {
				return err
			}
			if _, err = w.Write(insert.buf); err != nil {
				return err
			}
			index = insert.pos
		}

		if index < iter.index {
			if _, err = w.Write(iter.buf[index:iter.index]); err != nil {
				return err
			}
		}

		iter = iter.next
	}
	return nil
}

func (b *blockbuf) insertOnOffset(data []byte, pos *blockBufPos) {
	var index = sort.Search(len(pos.block.insertions), func(i int) bool { return pos.block.insertions[i].pos >= pos.index })

	if index < len(pos.block.insertions) {
		pos.block.insertions = append(pos.block.insertions[:index+1], pos.block.insertions[index:]...)
		pos.block.insertions[index] = newInsertion(data, pos.index)
	} else {
		pos.block.insertions = append(pos.block.insertions, newInsertion(data, pos.index))
	}
}

// func (b *blockbuf) getBytes() []byte {
// 	var (
// 		iter = b.first
// 		out  = make([]byte, 0)
// 	)

// 	for {
// 		out = append(out, iter.buf[:iter.index]...)
// 		if iter.next == nil {
// 			break
// 		} else {
// 			iter = iter.next
// 		}
// 	}
// 	return out
// }

// func (b *blockbuf) Print() {
// 	var iter = b.first
// 	for {
// 		fmt.Printf("%+v\n", iter.buf)
// 		if iter.next == nil {
// 			break
// 		} else {
// 			iter = iter.next
// 		}
// 	}
// }

func newBlock() *block {
	return &block{
		buf:        [bufsize]byte{},
		next:       nil,
		index:      0,
		insertions: make([]*blockInsertion, 0),
	}
}

func newInsertion(insertion []byte, pos int) *blockInsertion {
	var buf = make([]byte, len(insertion))
	copy(buf, insertion)
	return &blockInsertion{
		buf: insertion,
		pos: pos,
	}
}
