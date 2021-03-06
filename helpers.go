/*
 * Copyright (c) 2022, arivum.
 * All rights reserved.
 * SPDX-License-Identifier: MIT
 * For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/MIT
 */

package json2msgpackStreamer

func lenTo2Bytes(len int) []byte {
	lenBuf16[0] = byte(len >> 8)
	lenBuf16[1] = byte(len)
	return lenBuf16
}

func lenTo4Bytes(len int) []byte {
	lenBuf32[0] = byte(len >> 24)
	lenBuf32[1] = byte(len >> 16)
	lenBuf32[2] = byte(len >> 8)
	lenBuf32[3] = byte(len)
	return lenBuf32
}
