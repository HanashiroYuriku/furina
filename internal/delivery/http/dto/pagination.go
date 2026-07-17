package dto

type PaginationParam struct {
	Page     int `query:"page" json:"page" validate:"required,min=1"`
	PageSize int `query:"pageSize" json:"pageSize" validate:"required,min=1,max=100"`
}

type PaginationResponse[T any] struct {
	Data         T    `json:"items"`
	TotalRecords int  `json:"totalRecords"`
	TotalPages   int  `json:"totalPages"`
	CurrentPage  int  `json:"currentPage"`
	PageSize     int  `json:"pageSize"`
	HasNext      bool `json:"hasNext"`
	HasPrevious  bool `json:"hasPrevious"`
}

func NewPaginationResponse[T any](data T, totalRecords, currentPage, pageSize int) PaginationResponse[T] {
	totalPages := (totalRecords + pageSize - 1) / pageSize

	if currentPage < 1 {
		currentPage = 1
	}
	
	return PaginationResponse[T]{
		Data:         data,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		CurrentPage:  currentPage,
		PageSize:     pageSize,
		HasNext:      currentPage < totalPages,
		HasPrevious:  currentPage > 1,
	}
}