package generics

type (
	ApiResponse struct {
		Meta    ApiMeta     `json:"meta,omitempty"`
		Objects interface{} `json:"objects"`
	}
	ApiMeta struct {
		Limit      int    `json:"limit,omitempty"`
		Next       string `json:"next,omitempty"`
		Previous   string `json:"previous,omitempty"`
		Offset     int    `json:"offset,omitempty"`
		TotalCount int    `json:"totalCount,omitempty"`
	}
)

func NewApiResponse(objects interface{}, totalCount int) ApiResponse {
	response := ApiResponse{
		Meta:    ApiMeta{TotalCount: totalCount},
		Objects: objects,
	}
	return response
}
