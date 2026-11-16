// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package svn

import (
	"bytes"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
)

// DecodeText affects web previews only. Downloads and repositories retain their bytes.
func DecodeText(data []byte) string {
	if bytes.HasPrefix(data, []byte{0xff, 0xfe}) || bytes.HasPrefix(data, []byte{0xfe, 0xff}) {
		decoded, err := unicode.UTF16(unicode.LittleEndian, unicode.ExpectBOM).NewDecoder().Bytes(data)
		if err == nil {
			return string(decoded)
		}
	}
	data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	if utf8.Valid(data) {
		return string(data)
	}
	// Legacy text files can contain lines saved with different encodings.
	lines := bytes.SplitAfter(data, []byte{'\n'})
	var out bytes.Buffer
	for _, line := range lines {
		if utf8.Valid(line) {
			out.Write(line)
			continue
		}
		decoded, err := simplifiedchinese.GB18030.NewDecoder().Bytes(line)
		if err != nil {
			out.WriteString(string(bytes.ToValidUTF8(line, []byte("�"))))
		} else {
			out.Write(decoded)
		}
	}
	return out.String()
}
func IsBinary(data []byte) bool {
	if bytes.HasPrefix(data, []byte{0xff, 0xfe}) || bytes.HasPrefix(data, []byte{0xfe, 0xff}) {
		return false
	}
	return bytes.IndexByte(data, 0) >= 0
}
