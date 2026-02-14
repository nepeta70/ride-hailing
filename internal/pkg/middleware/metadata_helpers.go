package middleware

import (
	"strconv"

	"google.golang.org/grpc/metadata"
)

func getMetadata(md metadata.MD, key string) string {
	if vals := md.Get(key); len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func getIntMetadata(md metadata.MD, key string) int {
	if vals := md.Get(key); len(vals) > 0 {
		if n, err := strconv.Atoi(vals[0]); err == nil {
			return n
		}
	}
	return 0
}
