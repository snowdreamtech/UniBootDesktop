// Copyright (c) 2026 SnowdreamTech. All rights reserved.
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package env

import (
	"crypto/rand"
	"os"
)

// Get returns the value of the environment variable with the given key,
// searching with prefix UNIBOOTDESKTOP_ first, and then the raw key.
// Note: PATH is retrieved directly to avoid pollution from UNIBOOTDESKTOP_PATH.
func Get(key string) string {
	if key == "PATH" {
		return os.Getenv("PATH")
	}
	// 1. UNIBOOTDESKTOP_ prefix (Primary)
	if v := os.Getenv("UNIBOOTDESKTOP_" + key); v != "" {
		return v
	}
	// 2. Raw key (Native)
	return os.Getenv(key)
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
	LICENSE string = "MIT <https://github.com/snowdreamtech/unibootdesktop/blob/main/LICENSE>"

	//Config Config File Path
	Config string = "unibootdesktop.toml"

	// Debug indicates whether the application should run in debug mode.
	Debug bool

	// Quiet indicates whether the application should run in quiet mode.
	Quiet bool

	// Silent indicates whether to suppress all output and non-error messages.
	Silent bool

	CryptoRandRead = rand.Read
)

// RandomString returns a cryptographically secure random alphanumeric string of the specified length without modulo bias.
func RandomString(n int) (string, error) {
	if n <= 0 {
		return "", nil
	}
	const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const maxValidByte = 256 - (256 % len(letters)) // 248, evenly divisible by 62

	result := make([]byte, n)
	buf := make([]byte, n+(n/4)+4)
	idx := 0

	for idx < n {
		if _, err := CryptoRandRead(buf); err != nil {
			return "", err
		}
		for _, b := range buf {
			if int(b) < maxValidByte {
				result[idx] = letters[int(b)%len(letters)]
				idx++
				if idx == n {
					break
				}
			}
		}
	}
	return string(result), nil
}
