package middleware

import (
	"strconv"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

func getMetadata(md metadata.MD, key string) string {
	if vals := md.Get(key); len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func getUUIDMetadata(md metadata.MD, key string) uuid.UUID {
	if vals := md.Get(key); len(vals) > 0 {
		if id, err := uuid.Parse(vals[0]); err == nil {
			return id
		}
	}
	return uuid.Nil
}

func getIntMetadata(md metadata.MD, key string) int {
	if vals := md.Get(key); len(vals) > 0 {
		if n, err := strconv.Atoi(vals[0]); err == nil {
			return n
		}
	}
	return 0
}
