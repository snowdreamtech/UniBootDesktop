// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package env

import (
	"crypto/rand"
	"fmt"
	"os"
	"strings"
)

// EnvManager provides environment variable operations.
type EnvManager struct{}

// Get returns the value of the environment variable with the given key,
// searching with prefixes in order: UNIBOOTDESKTOP_, UNIGODESKTOP_, MISE_, and then the raw key.
// Note: PATH is retrieved directly to avoid pollution from UNIBOOTDESKTOP_PATH/UNIGODESKTOP_PATH/MISE_PATH.
func Get(key string) string {
	if key == "PATH" {
		return os.Getenv("PATH")
	}

	value := ""

	// 1. UNIBOOTDESKTOP_ prefix (Primary)
	if v := os.Getenv("UNIBOOTDESKTOP_" + key); v != "" {
		value = v
	}
	// 2. UNIGODESKTOP_ prefix (Legacy fallback)
	if value == "" {
		if v := os.Getenv("UNIGODESKTOP_" + key); v != "" {
			value = v
		}
	}
	// 3. MISE_ prefix
	if value == "" {
		if v := os.Getenv("MISE_" + key); v != "" {
			value = v
		}
	}
	// 4. Raw key (Native)
	if value == "" {
		value = os.Getenv(key)
	}

	// 验证特定关键环境变量
	switch key {
	case "GITHUB_PROXY":
		if value != "" && value != "direct" {
			// 简单验证URL格式
			if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
				return "" // 无效的代理URL，返回空
			}
		}
	case "JOBS":
		if value != "" {
			// 验证JOBS是正整数
			var n int
			if _, err := fmt.Sscanf(value, "%d", &n); err != nil || n < 1 || n > 256 {
				return "" // 无效的JOBS值，返回空
			}
		}
	case "HTTP2":
		// 只允许"0"或"1"
		if value != "" && value != "0" && value != "1" {
			return ""
		}
	}

	return value
}

// GithubProxy returns the configured GitHub proxy URL or empty string by default.
func GithubProxy() string {
	return Get("GITHUB_PROXY")
}

var (
	//ProjectName Project Name
	ProjectName string = "unibootdesktop"

	//Author Author
	Author string = "Snowdream Tech <snowdreamtech@qq.com>"

	//BuildTime Build Time
	BuildTime string = "N/A"

	//GitTag Git Tag
	GitTag string = "N/A"

	//CommitHash Commit Hash
	CommitHash string = "N/A"

	//CommitHashFull Commit Hash
	CommitHashFull string = "N/A"

	//COPYRIGHT COPYRIGHT
	COPYRIGHT string = "Copyright (c) 2023-present SnowdreamTech Inc."

	//LICENSE LICENSE
	LICENSE string = "MIT <https://github.com/snowdreamtech/unigodesktop/blob/main/LICENSE>"

	//Config Config File Path
	Config string = "unibootdesktop.toml"

	// Debug indicates whether the application should run in debug mode.
	Debug bool

	// Trace indicates whether the application should run in trace mode.
	Trace bool

	// Quiet indicates whether the application should run in quiet mode.
	Quiet bool

	// Cwd specifies the current working directory for the application.
	Cwd string

	// Silent indicates whether to suppress all output and non-error messages.
	Silent bool

	CryptoRandRead = rand.Read
)

// RandomString returns a random string of the specified length.
func RandomString(n int) (string, error) {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := make([]byte, n)
	if _, err := CryptoRandRead(bytes); err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}
