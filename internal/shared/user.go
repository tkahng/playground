package shared

const (
	SuperUserEmail string = "admin@k2dv.io"
)

type UpdateMeInput struct {
	Name  *string `json:"name"`
	Image *string `json:"image"`
}
