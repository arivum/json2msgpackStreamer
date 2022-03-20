package json2msgpackStreamer

import (
	"bufio"
	"io"
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
	buf        []byte
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