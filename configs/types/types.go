package types

// RouterArr maps routes to required permissions
type RouterArr map[string][]string

// User represents an authenticated user
type User struct {
	ID          string                 `json:"id"`
	Email       string                 `json:"email"`
	Permissions []string               `json:"permissions"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Pagination represents pagination parameters
type Pagination struct {
	Page     int64 `json:"page" form:"page"`
	PageSize int64 `json:"page_size" form:"page_size"`
}

// GetOffset returns the offset for database queries
func (p *Pagination) GetOffset() int64 {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	return (p.Page - 1) * p.PageSize
}

// Meta represents response metadata
type Meta struct {
	Page       int64 `json:"page"`
	PageSize   int64 `json:"page_size"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int64 `json:"total_pages,omitempty"`
}

// Claims represents JWT claims
type Claims struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Exp   int64  `json:"exp"`
}
