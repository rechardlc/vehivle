package postgres

import (
	"math"
)
type PageResult struct {
	Page int `json:"page"`
	PageSize int `json:"pageSize"`
	Total int `json:"total"`
	TotalPage int `json:"totalPage"`
}

func NewPage(page int, pageSize int, total int) *PageResult {
	return &PageResult{
		Page: page,
		PageSize: pageSize,
		Total: total,
		TotalPage: int(math.Ceil(float64(total) / float64(pageSize))),
	}
}