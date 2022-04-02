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

const (
	bufsize = 4 << 10
)

type blockInsertion struct {
	buf []byte
	pos int
}

type blockbuf struct {
	first   *block
	current *block
}

type block struct {
	next       *block
	buf        [bufsize]byte
	index      int
	insertions []*blockInsertion
}

type blockBufPos struct {
	block *block
	index int
}

type JSON2MsgPackStreamer struct {
	r                 *bufio.Reader
	pipeR             *io.PipeReader
	pipeW             *io.PipeWriter
	underlayingReader io.Reader
	buf               *blockbuf
	nextByte          byte
	lastError         error
}
