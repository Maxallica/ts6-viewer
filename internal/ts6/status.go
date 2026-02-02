package ts6

// Status represents the standard TS6 API status block
type Status struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
