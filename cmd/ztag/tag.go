package ztag

import (
	"fmt"
	"regexp"
	"strconv"
)

var supportedTagTemplates = []TagTemplate{
	&TagTemplate1{},
	&TagTemplate2{},
}

func GenerateNextTag(oldTag string, level Level, env Env) (string, error) {
	for _, template := range supportedTagTemplates {
		if template.Regex().MatchString(oldTag) {
			c, err := template.Extractor(oldTag)
			if err != nil {
				return "", err
			}
			c = c.Next(level)
			return template.Generator(c, env), nil
		}
	}
	return "", fmt.Errorf("tag does not match any supported template")
}

// TagComponents holds all parts needed to reconstruct a tag.
type TagComponents struct {
	Major int
	Minor int
	Patch int
}

func (c TagComponents) Next(level Level) TagComponents {
	switch level {
	case LevelMajor:
		c.Major++
		c.Minor = 0
		c.Patch = 0
	case LevelMinor:
		c.Minor++
		c.Patch = 0
	case LevelBug:
		c.Patch++
	default:
		c.Patch++
	}
	return c
}

// TagTemplate defines a supported tag format for both matching and generation.
type TagTemplate interface {
	Regex() *regexp.Regexp
	Extractor(tag string) (TagComponents, error)
	Generator(c TagComponents, env Env) string
}

type TagTemplate1 struct{} // qc-v1.0.0, stg-v1.0.0, prod-v1.0.0

func (t *TagTemplate1) Regex() *regexp.Regexp {
	return regexp.MustCompile(`^([a-zA-Z]+)-v(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)$`)
}

func (t *TagTemplate1) Extractor(tag string) (TagComponents, error) {
	match := t.Regex().FindStringSubmatch(tag)
	if len(match) == 0 {
		return TagComponents{}, fmt.Errorf("tag does not match template 1")
	}
	result := map[string]string{}
	names := t.Regex().SubexpNames()
	for i, name := range names {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return TagComponents{
		Major: mustAtoi(result["major"]),
		Minor: mustAtoi(result["minor"]),
		Patch: mustAtoi(result["patch"]),
	}, nil
}

func (t *TagTemplate1) Generator(c TagComponents, env Env) string {
	return fmt.Sprintf("%s-v%d.%d.%d", string(env), c.Major, c.Minor, c.Patch)
}

type TagTemplate2 struct{} // v1.0.0, v1.0.0-beta, v1.0.0-alpha, v1.0.0-rc

func (t *TagTemplate2) Regex() *regexp.Regexp {
	return regexp.MustCompile(`^v(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)(-(\w+))?$`)
}

func (t *TagTemplate2) Extractor(tag string) (TagComponents, error) {
	match := t.Regex().FindStringSubmatch(tag)
	if len(match) == 0 {
		return TagComponents{}, fmt.Errorf("tag does not match template 2")
	}
	result := map[string]string{}
	names := t.Regex().SubexpNames()
	for i, name := range names {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return TagComponents{
		Major: mustAtoi(result["major"]),
		Minor: mustAtoi(result["minor"]),
		Patch: mustAtoi(result["patch"]),
	}, nil
}

func (t *TagTemplate2) Generator(c TagComponents, env Env) string {
	return fmt.Sprintf("v%d.%d.%d-%s", c.Major, c.Minor, c.Patch, string(env))
}

func mustAtoi(s string) int {
	if s == "" {
		return 0
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Sprintf("Regex matched non-integer value: %s", s))
	}
	return i
}
