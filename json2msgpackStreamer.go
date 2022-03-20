/*
 * Copyright (c) 2022, arivum.
 * All rights reserved.
 * SPDX-License-Identifier: MIT
 * For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/MIT
 */

package json2msgpackStreamer

import (
	"bufio"
	"io"
)

func NewJSON2MsgPackStreamer(r io.Reader) *JSON2MsgPackStreamer {
	var (
		pipeR, pipeW = io.Pipe()
		t            = &JSON2MsgPackStreamer{
			r:                 bufio.NewReader(r),
			underlayingReader: r,
			pipeW:             pipeW,
			pipeR:             pipeR,
			buf:               newBlockBuf(),
		}
	)

	go t.convert()

	return t
}

func (t *JSON2MsgPackStreamer) Read(p []byte) (int, error) {
	return t.pipeR.Read(p)
}

func (t *JSON2MsgPackStreamer) GetLastError() error {
	return t.lastError
}

func (t *JSON2MsgPackStreamer) convert() {
	defer t.pipeW.Close()
	for {
		if t.nextByte, t.lastError = t.r.ReadByte(); t.lastError != nil {
			break
		}

		switch t.nextByte {
		case ' ', '\t', '\r':
			// do nothing
		case '{':
			if t.lastError = t.handleStruct(); t.lastError != nil {
				return
			}
		case '[':
			if t.lastError = t.handleArray(); t.lastError != nil {
				return
			}
		case '"':
			if t.lastError = t.handleString(); t.lastError != nil {
				return
			}
		case '\n':
			t.buf.flushToWriter(t.pipeW)
			t.buf = newBlockBuf()
			t.buf.reset()
		default:
			t.r.UnreadByte()
			if t.lastError = t.handleAtomic(); t.lastError != nil {
				return
			}
		}
	}

	if t.nextByte != '\n' {
		t.buf.flushToWriter(t.pipeW)
	}
	t.buf.reset()
}
