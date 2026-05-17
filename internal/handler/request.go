package handler

type CreateRequest struct {
	ServiceName string  `json:"service_name" binding:"required,min=1,max=255"`
	Price       int     `json:"price" binding:"required,min=0"`
	UserID      string  `json:"user_id" binding:"required,uuid"`
	StartDate   string  `json:"start_date" binding:"required,len=7"`
	EndDate     *string `json:"end_date,omitempty" binding:"omitempty,len=7"`
}

type UpdateRequest struct {
	ServiceName *string `json:"service_name,omitempty" binding:"omitempty,min=1,max=255"`
	Price       *int    `json:"price,omitempty" binding:"omitempty,min=0"`
	StartDate   *string `json:"start_date,omitempty" binding:"omitempty,len=7"`
	EndDate     *string `json:"end_date,omitempty" binding:"omitempty,len=7"`
}
