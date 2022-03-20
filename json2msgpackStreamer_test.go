package json2msgpackStreamer

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
)

func TestJSON2MsgPackStreamer(t *testing.T) {
	tests := []struct {
		name string
		arg  interface{}
		want interface{}
	}{
		{
			name: "atomic empty string",
			arg:  "",
			want: "",
		},
		{
			name: "atomic string; len less than 16 byte",
			arg:  "teststring",
			want: "teststring",
		},
		{
			name: "atomic string, len less than 2^16 byte",
			arg:  strings.Repeat("0", 2<<10),
			want: strings.Repeat("0", 2<<10),
		},
		{
			name: "atomic string, len less than 2^32 byte",
			arg:  strings.Repeat("0", 2<<17),
			want: strings.Repeat("0", 2<<17),
		},
		{
			name: "atomic zero uint8",
			arg:  uint8(0),
			want: int8(0),
		},
		{
			name: "atomic positive fixing",
			arg:  uint8(10),
			want: int8(10),
		},
		{
			name: "atomic uint8",
			arg:  uint8(129),
			want: uint8(129),
		},
		{
			name: "atomic zero uint16",
			arg:  uint16(0),
			want: int8(0),
		},
		{
			name: "atomic uint16",
			arg:  uint16(0xfff),
			want: uint16(0xfff),
		},
		{
			name: "atomic zero uint32",
			arg:  uint32(0),
			want: int8(0),
		},
		{
			name: "atomic uint32",
			arg:  uint32(0xfffffff),
			want: uint32(0xfffffff),
		},
		{
			name: "atomic zero uint64",
			arg:  uint64(0),
			want: int8(0),
		},
		{
			name: "atomic uint64",
			arg:  uint64(0xfffffffffff),
			want: uint64(0xfffffffffff),
		},
		{
			name: "atomic zero int8",
			arg:  int8(0),
			want: int8(0),
		},
		{
			name: "atomic negative fixing",
			arg:  int8(-10),
			want: int8(-10),
		},
		{
			name: "atomic int8",
			arg:  int8(-32),
			want: int8(-32),
		},
		{
			name: "atomic zero int16",
			arg:  int16(0),
			want: int8(0),
		},
		{
			name: "atomic int16",
			arg:  int16(-10000),
			want: int16(-10000),
		},
		{
			name: "atomic zero int32",
			arg:  int32(0),
			want: int8(0),
		},
		{
			name: "atomic int32",
			arg:  int32(-10000000),
			want: int32(-10000000),
		},
		{
			name: "atomic zero int64",
			arg:  int64(0),
			want: int8(0),
		},
		{
			name: "atomic int64",
			arg:  int64(-1000000000000),
			want: int64(-1000000000000),
		},
		{
			name: "atomic zero float32",
			arg:  float32(0.0),
			want: int8(0),
		},
		{
			name: "atomic float32",
			arg:  float32(-1.2),
			want: float64(-1.2),
		},
		{
			name: "atomic zero float64",
			arg:  float64(0.0),
			want: int8(0),
		},
		{
			name: "atomic float64",
			arg:  float64(-1000000000000.2),
			want: float64(-1000000000000.2),
		},
		{
			name: "atomic bool false",
			arg:  false,
			want: false,
		},
		{
			name: "atomic bool true",
			arg:  true,
			want: true,
		},
		{
			name: "atomic nil",
			arg:  nil,
			want: nil,
		},
		{
			name: "map empty",
			arg:  map[string]interface{}{},
			want: map[string]interface{}{},
		},
		{
			name: "map less than 16 different entries",
			arg: map[string]interface{}{
				"0": int8(0), "1": nil, "2": false, "3": float64(1.2), "4": "teststring",
			},
			want: map[string]interface{}{
				"0": int8(0), "1": nil, "2": false, "3": float64(1.2), "4": "teststring",
			},
		},
		{
			name: "map less than 2^16 different entries",
			arg: map[string]interface{}{
				"0": int8(0), "1": nil, "2": false, "3": float64(1.2), "4": "teststring", "5": nil, "6": nil, "7": nil, "8": nil, "9": nil, "10": nil, "11": nil, "12": nil, "13": nil, "14": nil, "15": nil, "16": nil, "17": nil, "18": nil, "19": nil,
			},
			want: map[string]interface{}{
				"0": int8(0), "1": nil, "2": false, "3": float64(1.2), "4": "teststring", "5": nil, "6": nil, "7": nil, "8": nil, "9": nil, "10": nil, "11": nil, "12": nil, "13": nil, "14": nil, "15": nil, "16": nil, "17": nil, "18": nil, "19": nil,
			},
		},
		{
			name: "string array less than 16 entries",
			arg: []string{
				"0", "1", "2", "3",
			},
			want: []interface{}{
				"0", "1", "2", "3",
			},
		},
		{
			name: "array less than 16 different entries",
			arg: []interface{}{
				"0", int8(1), float64(2.1), true, false, nil,
			},
			want: []interface{}{
				"0", int8(1), float64(2.1), true, false, nil,
			},
		},
		{
			name: "string array less than 2^16 entries",
			arg: []string{
				"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19",
			},
			want: []interface{}{
				"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19",
			},
		},
		{
			name: "array less than 2^16 different entries",
			arg: []interface{}{
				"0", int8(1), float64(2.1), true, false, nil, "7", int8(8), float64(9.1), true, false, nil, "13", int8(14), float64(15.1), true, false, nil, nil, true,
			},
			want: []interface{}{
				"0", int8(1), float64(2.1), true, false, nil, "7", int8(8), float64(9.1), true, false, nil, "13", int8(14), float64(15.1), true, false, nil, nil, true,
			},
		},
		{
			name: "complex array, less than 2^16 different entries including map",
			arg: []interface{}{
				"0", int8(-1), float64(2.1), true, false, nil, "7", int8(8), float64(9.1), map[string]interface{}{"1": "teststring", "2": int8(0)}, true, false, "13", int8(-14), float64(15.1), true, false, nil, true,
			},
			want: []interface{}{
				"0", int8(-1), float64(2.1), true, false, nil, "7", int8(8), float64(9.1), map[string]interface{}{"1": "teststring", "2": int8(0)}, true, false, "13", int8(-14), float64(15.1), true, false, nil, true,
			},
		},
		{
			name: "complex array, less than 2^16 different entries including another array",
			arg: []interface{}{
				"0", int8(-1), float64(2.1), true, false, nil, "7", int8(8), float64(9.1), []interface{}{
					"0", int8(1), float64(2.1), true, false, nil,
				}, true, false, "13", int8(-14), float64(15.1), []interface{}{
					"0", int8(1), float64(2.1), true, false, nil,
				}, false, nil, true,
			},
			want: []interface{}{
				"0", int8(-1), float64(2.1), true, false, nil, "7", int8(8), float64(9.1), []interface{}{
					"0", int8(1), float64(2.1), true, false, nil,
				}, true, false, "13", int8(-14), float64(15.1), []interface{}{
					"0", int8(1), float64(2.1), true, false, nil,
				}, false, nil, true,
			},
		},
		{
			name: "complex map, less than 2^16 different entries including bigger map",
			arg: map[string]interface{}{
				"0": int8(1), "1": nil, "2": false, "3": float64(1.2), "4": "teststring", "5": map[string]interface{}{
					"0": int8(1), "1": nil, "2": false, "3": float64(1.2), "4": "teststring", "5": nil, "6": nil, "7": nil, "8": nil, "9": nil, "10": nil, "11": nil, "12": nil, "13": nil, "14": nil, "15": nil, "16": nil, "17": nil, "18": nil, "19": nil,
				}, "6": map[string]interface{}{
					"0": int8(1), "1": nil, "2": false, "3": float64(1.2), "4": "teststring", "5": nil, "6": nil, "7": nil, "8": nil, "9": nil, "10": nil, "11": nil, "12": nil, "13": nil, "14": nil, "15": nil, "16": nil, "17": nil, "18": nil, "19": nil,
				}, "7": nil, "8": nil, "9": nil, "10": nil, "11": nil, "12": nil, "13": nil, "14": nil, "15": nil, "16": nil, "17": nil, "18": nil, "19": nil,
			},
			want: map[string]interface{}{
				"0": int8(1), "1": nil, "2": false, "3": float64(1.2), "4": "teststring", "5": map[string]interface{}{
					"0": int8(1), "1": nil, "2": false, "3": float64(1.2), "4": "teststring", "5": nil, "6": nil, "7": nil, "8": nil, "9": nil, "10": nil, "11": nil, "12": nil, "13": nil, "14": nil, "15": nil, "16": nil, "17": nil, "18": nil, "19": nil,
				}, "6": map[string]interface{}{
					"0": int8(1), "1": nil, "2": false, "3": float64(1.2), "4": "teststring", "5": nil, "6": nil, "7": nil, "8": nil, "9": nil, "10": nil, "11": nil, "12": nil, "13": nil, "14": nil, "15": nil, "16": nil, "17": nil, "18": nil, "19": nil,
				}, "7": nil, "8": nil, "9": nil, "10": nil, "11": nil, "12": nil, "13": nil, "14": nil, "15": nil, "16": nil, "17": nil, "18": nil, "19": nil,
			},
		},
		{
			name: "complex map, less than 2^16 different entries including array",
			arg: map[string]interface{}{
				"0": int8(0), "1": nil, "2": false, "3": float64(1.2), "4": "teststring", "5": []interface{}{
					"0", int8(1), float64(2.1), true, false, nil,
				}, "6": nil, "7": nil, "8": nil, "9": nil, "10": nil, "11": nil, "12": nil, "13": nil, "14": nil, "15": nil, "16": nil, "17": nil, "18": nil, "19": nil,
			},
			want: map[string]interface{}{
				"0": int8(0), "1": nil, "2": false, "3": float64(1.2), "4": "teststring", "5": []interface{}{
					"0", int8(1), float64(2.1), true, false, nil,
				}, "6": nil, "7": nil, "8": nil, "9": nil, "10": nil, "11": nil, "12": nil, "13": nil, "14": nil, "15": nil, "16": nil, "17": nil, "18": nil, "19": nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, w := io.Pipe()
			jsonDec := json.NewEncoder(w)
			go func() {
				jsonDec.Encode(tt.arg)
				w.Close()
			}()
			streamer := NewJSON2MsgPackStreamer(r)

			var got interface{}
			msgpackDec := msgpack.NewDecoder(streamer)
			err := msgpackDec.Decode(&got)

			if err != nil {
				t.Errorf("msgpackDec error %v, want nil", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				fmt.Printf("%+v", got)
				t.Errorf("NewJSON2MsgPackStreamer = %+T, want %+T", got, tt.want)
			}
		})
	}
}
