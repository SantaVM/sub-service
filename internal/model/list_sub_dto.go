package model

type ListSubscriptionsQuery struct {
	UserID      *string `json:"user_id,omitempty" validate:"omitempty,uuid"`
	ServiceName *string `json:"service_name,omitempty"`
	Size        int     `json:"size" validate:"required,number,min=1,max=100"`
	Page        int     `json:"page" validate:"required,number,min=1"`
}
