package web

// PageModel combines common model data for pages
type PageModel struct {
	Development   bool
	Integration   bool
	Authenticated bool
	User          UserModel
	CurrPage      string
}

// UserModel provides user information for pages
type UserModel struct {
	Username    string
	Email       string
	UserID      string
	DisplayName string
	ProfileURL  string
	Token       string
}

func (p PageModel) IsActive(site string) string {
	if p.CurrPage == site {
		return "active"
	}
	return ""
}
