package environment

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	envReg *regexp.Regexp
)

func init() {
	envReg = regexp.MustCompile(`\$\{([a-zA-Z0-9_]*)\}`)
}

func HandleEnv(in []byte) ([]byte, error) {
	content := string(in)
	matches := envReg.FindAllStringSubmatch(content, -1)
	envMap := make(map[string]string)

	for _, match := range matches {
		if len(match) < 1 {
			continue
		}

		key := match[1]
		if value, ok := os.LookupEnv(key); ok {
			envMap[key] = value
		} else {
			return nil, fmt.Errorf("%s not set", key)
		}
	}

	for k, v := range envMap {
		content = strings.Replace(content, fmt.Sprintf("${%s}", k), v, -1)
	}

	return []byte(content), nil
}