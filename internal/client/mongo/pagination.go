package mongo

import (
	"github.com/rigdev/rig-go-api/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	PaginationDefault = 50
	PaginationLimit   = 1000
)

func SortOptions(p *model.Pagination) *options.FindOptions {
	o := options.Find()
	order := 1
	if p.GetDescending() {
		order = -1
	}
	o.SetSort(bson.D{{Key: "_id", Value: order}})
	var limit uint32 = PaginationDefault
	if p.GetLimit() > 0 && p.GetLimit() <= PaginationLimit {
		limit = p.GetLimit()
	}
	o.SetLimit(int64(limit))
	if p.GetOffset() > 0 {
		o.SetSkip(int64(p.GetOffset()))
	}
	return o
}
