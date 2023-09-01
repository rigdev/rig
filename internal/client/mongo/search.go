package mongo

import (
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func StringSearch(search string) bson.M {
	var rs []string
	for _, p := range strings.Split(search, " ") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		rs = append(rs, regexp.QuoteMeta(p))
	}

	if len(rs) == 0 {
		return bson.M{}
	}

	return bson.M{
		"search": bson.M{
			"$regex": primitive.Regex{
				Pattern: strings.Join(rs, "|"),
				Options: "i",
			},
		},
	}
}
