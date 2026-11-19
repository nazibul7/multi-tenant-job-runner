package tenant

type Tenant struct {
	ID             string `json:"id"`
	Budget         int    `json:"budget"`
	Spent          int    `json:"spent"`
	MaxConcurrency int    `json:"max_concurrency"`
	Running        int    `json:"running"`
}