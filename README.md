# json2msgpackStreamer


![](https://img.shields.io/github/v/tag/arivum/json2msgpackStreamer?label=latest&color=%234488BB)
![](https://img.shields.io/github/go-mod/go-version/arivum/json2msgpackStreamer?color=%234488BB)
![](https://img.shields.io/github/workflow/status/arivum/json2msgpackStreamer/Go)
![](https://img.shields.io/github/license/arivum/json2msgpackStreamer?color=%234488BB)


This module converts a (ND)JSON stream to a messagepack stream on-the-fly.

[See package documentation](https://pkg.go.dev/github.com/arivum/json2msgpackStreamer)

## How to use

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/arivum/json2msgpackStreamer"
	"github.com/vmihailenco/msgpack/v5"
)

func main() {
	var rawJSON = map[string]interface{}{
		"a": 0,
		"b": true,
		"c": []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
		"d": map[string]interface{}{
			"d_a": 1,
			"d_b": "teststring",
		},
		"e": float32(1.2),
		"f": float64(2),
		"g": 0x5a,
		"h": nil,
		"i": -(1 << 33),
		"j": true,
		"k": []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p"},
		"l": map[string]interface{}{
			"d_a": 1,
			"d_b": "teststring",
		},
		"m": float32(1.2),
		"n": float64(2),
		"o": 0x5a,
		"p": nil,
	}
	var rawJSON2 = map[string]interface{}{
		"a": 0,
		"b": true,
		"c": []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p"},
		"d": map[string]interface{}{
			"d_a": 1,
			"d_b": "teststring",
		},
	}

	// Create the JSON buffer that needs to be converted to a msgpack stream
	var jsonObj, _ = json.Marshal(rawJSON)
	var jsonObj2, _ = json.Marshal(rawJSON2)
	jsonObj = append(jsonObj, jsonObj...)
	jsonObj = append(jsonObj, jsonObj2...)
	jsonObj = append(jsonObj, jsonObj...)
	jsonObj = append(jsonObj, jsonObj2...)
	fmt.Println(string(jsonObj))

	// Create the converter:
	// JSON2MsgPackStreamer is of type io.Reader and can be passed to e.g. "github.com/vmihailenco/msgpack/v5".NewDecoder(r io.Reader)
	var conv = json2msgpackStreamer.NewJSON2MsgPackStreamer(bytes.NewBuffer(jsonObj))

	// Create the msgpack decoder that reads from the converter
	msgpackDec := msgpack.NewDecoder(conv)

	i := 0
	for {
		var entry interface{}

		if err := msgpackDec.Decode(&entry); err != nil {
			break
		}
		fmt.Printf("Entry #%d: %+v\n", i, entry)
		i++
	}
}
```
