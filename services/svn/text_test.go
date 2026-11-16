// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package svn

import (
	"bytes"
	"golang.org/x/text/encoding/simplifiedchinese"
	"testing"
)

func TestDecodeChineseTextAndMixedEncoding(t *testing.T) {
	text := "中文目录/说明.txt\r\n同步说明与文件内容\r\n"
	legacy, err := simplifiedchinese.GB18030.NewEncoder().Bytes([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	before := bytes.Clone(legacy)
	if got := DecodeText(legacy); got != text {
		t.Fatalf("GBK preview mismatch: %q", got)
	}
	if !bytes.Equal(before, legacy) {
		t.Fatal("decoder modified original bytes")
	}
	mixed := append([]byte("UTF-8中文内容\n"), legacy...)
	if got := DecodeText(mixed); got != "UTF-8中文内容\n"+text {
		t.Fatalf("mixed encoding mismatch: %q", got)
	}
	if got := DecodeText([]byte(text)); got != text {
		t.Fatal("UTF-8 changed")
	}
	if got := DecodeText([]byte{0xff, 0xfe, 0x2d, 0x4e, 0x87, 0x65}); got != "中文" {
		t.Fatalf("UTF16 mismatch: %q", got)
	}
	if IsBinary([]byte{0xff, 0xfe, 0x2d, 0x4e, 0x87, 0x65}) || !IsBinary([]byte{'a', 0, 'b'}) {
		t.Fatal("binary detection regression")
	}
}
