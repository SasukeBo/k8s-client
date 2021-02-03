package environment

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

var (
	envReg      *regexp.Regexp
	envMatchReg *regexp.Regexp
)

func init() {
	envReg = regexp.MustCompile(`\$\{([a-zA-Z0-9_]*)\}`)
	envMatchReg = regexp.MustCompile(`\$([a-zA-Z0-9_]*)`)
}

func HandleEnv(in []byte, mergePath string) ([]byte, error) {
	content := string(in)
	matches := envReg.FindAllStringSubmatch(content, -1)
	envMap := make(map[string]string)
	envMerge := make(map[string]interface{})
	var totalEnv []string

	if len(mergePath) != 0 {
		content, err := ioutil.ReadFile(mergePath)
		if err != nil {
			return nil, fmt.Errorf("read env merge file failed: %v", err)
		}

		if err := yaml.Unmarshal(content, &envMerge); err != nil {
			return nil, fmt.Errorf("handle env merge file failed: %v", err)
		}

		for k, v := range envMerge {
			envMap[k] = fmt.Sprint(v)
		}
	}

	for _, match := range matches {
		if len(match) < 1 {
			continue
		}

		key := match[1]
		totalEnv = append(totalEnv, key)
		if value, ok := os.LookupEnv(key); ok {
			envMap[key] = value
		}
	}

	// 处理覆盖环境变量
	for k, v := range envMerge {
		value := fmt.Sprint(v)
		if envMatchReg.MatchString(value) {
			if matches := envMatchReg.FindAllStringSubmatch(value, -1); len(matches) > 0 {
				if len(matches[0]) > 1 {
					key := matches[0][1]
					if replace, ok := envMap[key]; ok {
						envMap[k] = replace
					} else {
						if value, ok := os.LookupEnv(key); ok {
							envMap[k] = value
						} else {
							return nil, fmt.Errorf("missing environment %s", key)
						}
					}
				}
			}
		} else {
			envMap[k] = value
		}
	}

	// 校验环境变量是否都存在
	for _, v := range totalEnv {
		_, ok := envMap[v]
		if !ok {
			return nil, fmt.Errorf("missing environment %s", v)
		}
	}

	for k, v := range envMap {
		content = strings.Replace(content, fmt.Sprintf("${%s}", k), v, -1)
	}

	return []byte(content), nil
}
