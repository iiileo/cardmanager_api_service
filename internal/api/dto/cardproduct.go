package dto

type CardProductItemRequest struct {
	Name  string `json:"name" binding:"required"`
	Times int    `json:"times" binding:"required"`
}

type CreateCardProductRequest struct {
	Type        string                    `json:"type" binding:"required"`
	Name        string                    `json:"name" binding:"required"`
	Price       *int                      `json:"price"`
	Times       *int                      `json:"times"`
	ValidMonths *int                      `json:"valid_months"`
	Items       []CardProductItemRequest  `json:"items"`
}

type CardProductItemResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Times int    `json:"times"`
	Sort  int    `json:"sort"`
}

type CardProductResponse struct {
	ID          string                     `json:"id"`
	Type        string                     `json:"type"`
	Name        string                     `json:"name"`
	Price       int                        `json:"price"`
	Times       *int                       `json:"times,omitempty"`
	ValidMonths *int                       `json:"valid_months,omitempty"`
	Status      int8                       `json:"status"`
	Items       []*CardProductItemResponse `json:"items,omitempty"`
}

type CardProductListResponse struct {
	List []*CardProductResponse `json:"list"`
}
