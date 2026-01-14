package services

import "regexp"

func ShortViewPars(tmpl string, vals map[string]string) (string, error) {
	re := regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)
	out := re.ReplaceAllStringFunc(tmpl, func(m string) string {
		sub := re.FindStringSubmatch(m)
		if len(sub) < 2 {
			return m
		}
		if v, ok := vals[sub[1]]; ok {
			return v
		}
		return m
	})
	return out, nil
}
