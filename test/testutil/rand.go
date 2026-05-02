/**
 * 功能：测试用唯一前缀与随机工具
 * 创建时间：2026-04-28
 * 创建人：GPT-5.2
 */

package testutil

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

func UniquePrefix() string {
	return fmt.Sprintf("E2E_%s_%s", time.Now().Format("20060102_150405"), randHex(3))
}

func randHex(nBytes int) string {
	b := make([]byte, nBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

