package service

import (
	"regexp"

	"github.com/jackc/pgx/v5/pgxpool"
)

var tagRegex = regexp.MustCompile(`(?:^|\s)#([\p{L}\p{N}_]+)`)

func ExtractTags(content string) []string {
	matches := tagRegex.FindAllStringSubmatch(content, -1)
	seen := make(map[string]struct{}, len(matches))
	tags := make([]string, 0, len(matches))
	for _, m := range matches {
		tag := m[1]
		if _, ok := seen[tag]; !ok {
			seen[tag] = struct{}{}
			tags = append(tags, tag)
		}
	}
	return tags
}

type TagService struct {
	pool *pgxpool.Pool
}

func NewTagService(pool *pgxpool.Pool) *TagService {
	return &TagService{pool: pool}
}
