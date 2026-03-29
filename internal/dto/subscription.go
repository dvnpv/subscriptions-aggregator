package dto

type CreateSubscriptionRequest struct {
	ServiceName string  `json:"service_name" example:"Yandex Plus"`
	Price       int     `json:"price" example:"400"`
	UserID      string  `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string  `json:"start_date" example:"07-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"09-2025"`
}

type UpdateSubscriptionRequest struct {
	ServiceName string  `json:"service_name" example:"Yandex Plus"`
	Price       int     `json:"price" example:"400"`
	UserID      string  `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string  `json:"start_date" example:"07-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"09-2025"`
}

type SubscriptionResponse struct {
	ID          string  `json:"id" example:"f7d0d0e1-7f79-4d1a-b2f4-7bdf6e889111"`
	ServiceName string  `json:"service_name" example:"Yandex Plus"`
	Price       int     `json:"price" example:"400"`
	UserID      string  `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string  `json:"start_date" example:"07-2025"`
	EndDate     *string `json:"end_date,omitempty" example:"09-2025"`
	CreatedAt   string  `json:"created_at" example:"2026-03-25T12:00:00Z"`
	UpdatedAt   string  `json:"updated_at" example:"2026-03-25T12:00:00Z"`
}

type ListSubscriptionsResponse struct {
	Items []SubscriptionResponse `json:"items"`
	Count int                    `json:"count"`
}

type TotalResponse struct {
	Total       int    `json:"total_rub" example:"2400"`
	From        string `json:"from" example:"01-2025"`
	To          string `json:"to" example:"12-2025"`
	UserID      string `json:"user_id,omitempty" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	ServiceName string `json:"service_name,omitempty" example:"Yandex Plus"`
}
