package store

// MockRepo implements the repository but holds entries only in memory
// this is a defaulf implementation which can be used for testing
type MockRepo struct {
	sites  map[string][]UserSiteEntity
	logins map[string]int64
}

// compile guard for interface
var _ Repository = &MockRepo{}

func NewMock(sites map[string][]UserSiteEntity) Repository {
	return &MockRepo{
		sites:  sites,
		logins: make(map[string]int64),
	}
}

func (m *MockRepo) GetSitesForUser(user string) ([]UserSiteEntity, error) {
	if _, ok := m.sites[user]; ok {
		return m.sites[user], nil
	}
	return make([]UserSiteEntity, 0), nil
}

func (m *MockRepo) GetUsersForSite(site string) ([]string, error) {
	var users []string
	var foundUsers map[string]bool = make(map[string]bool)
	for _, sites := range m.sites {
		for _, s := range sites {
			if s.Name == site {
				if _, ok := foundUsers[s.User]; !ok {
					users = append(users, s.User)
					foundUsers[s.User] = true
				}
			}
		}
	}
	if len(users) == 0 {
		users = make([]string, 0)
	}
	return users, nil
}

func (m *MockRepo) StoreSiteForUser(sites []UserSiteEntity) (err error) {
	user := sites[0].User
	m.sites[user] = sites
	return nil
}
