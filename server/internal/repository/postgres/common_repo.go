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
// 在go由于没有new关键字，所以需要自己实现一个函数来创建一个PageResult对象
// 默认使用New*函数来创建对象来模拟new关键字
func NewPage(page int, pageSize int, total int) *PageResult {
	return &PageResult{
		Page: page,
		PageSize: pageSize,
		Total: total,
		TotalPage: int(math.Ceil(float64(total) / float64(pageSize))),
	}
}