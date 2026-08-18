// Package configx 加载 yaml 配置：文件内容先做环境变量展开（${VAR} / ${VAR:-default}），
// 满足 docs/04 §1.8 的"默认值 < config.yaml < 环境变量"优先级。
package configx

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var envPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(:-([^}]*))?\}`)

// Load 读取 path 的 yaml，展开环境变量后反序列化进 out。
func Load(path string, out any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config %s: %w", path, err)
	}
	expanded := envPattern.ReplaceAllStringFunc(string(raw), func(m string) string {
		groups := envPattern.FindStringSubmatch(m)
		name, def := groups[1], groups[3]
		if v, ok := os.LookupEnv(name); ok && strings.TrimSpace(v) != "" {
			return v
		}
		return def
	})
	if err := yaml.Unmarshal([]byte(expanded), out); err != nil {
		return fmt.Errorf("parse config %s: %w", path, err)
	}
	return nil
}
