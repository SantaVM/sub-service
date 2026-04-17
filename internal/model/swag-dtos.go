package model

type SubscriptionPageResponse struct {
	Content          []*Subscription `json:"content"`
	Size             int             `json:"size"`
	Page             int             `json:"page"`
	NumberOfElements int             `json:"numberOfElements"`
	TotalElements    int64           `json:"totalElements"`
	HasNext          bool            `json:"hasNext"`
	HasPrevious      bool            `json:"hasPrevious"`
}

type Page[T any] struct {
	Content          []T   `json:"content"`
	Size             int   `json:"size"`
	Page             int   `json:"page"`
	NumberOfElements int   `json:"numberOfElements"`
	TotalElements    int64 `json:"totalElements"`
	HasNext          bool  `json:"hasNext"`
	HasPrevious      bool  `json:"hasPrevious"`
}

func NewPage[T any](content []T, size, page int, totalElements int64) *Page[T] {
	numberOfElements := len(content)
	hasNext := int64(page*size) < totalElements
	hasPrevious := page > 1

	return &Page[T]{
		Content:          content,
		Size:             size,
		Page:             page,
		NumberOfElements: numberOfElements,
		TotalElements:    totalElements,
		HasNext:          hasNext,
		HasPrevious:      hasPrevious,
	}
}
