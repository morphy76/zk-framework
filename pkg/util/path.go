package util

import "strings"

func ConcatPaths(paths ...string) string {
	if len(paths) == 0 {
		return "/"
	}

	for path := range paths {
		paths[path] = strings.Trim(paths[path], "/")
	}

	usePaths := []string{}
	for _, path := range paths {
		if path != "" {
			usePaths = append(usePaths, path)
		}
	}
	return "/" + strings.Join(usePaths, "/")
}
