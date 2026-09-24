// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"strings"
	"testing"
)

func TestJoinTitleAndTextTruncates(t *testing.T) {
	got := joinTitleAndText(openDataHubDetail{Title: "t", BaseText: strings.Repeat("ä", 3000)})
	if n := len([]rune(got)); n != maxCommentLength {
		t.Fatalf("length = %d, want %d", n, maxCommentLength)
	}
}
