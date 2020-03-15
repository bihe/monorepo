package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"golang.binggl.net/commons/security"
)

type args struct {
	Secret      string
	UserName    string
	UserID      string
	Email       string
	DisplayName string
	Issuer      string
}

type site struct {
	Name     string
	URL      string
	PermList string
}

const ExpiryDays = 100

func main() {
	fmt.Println("dev-helper will set cookies for local development!")
	token := createJWT(parseFlags())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		cookie := http.Cookie{
			Name:     "login_token",
			Value:    token,
			Domain:   "localhost",
			Path:     "/",
			MaxAge:   ExpiryDays * 8 * 3600, /* exp in seconds */
			Secure:   false,
			HttpOnly: true, // only let the api access those cookies
		}
		http.SetCookie(w, &cookie)
		fmt.Fprint(w, token)
	})
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func parseFlags() *args {
	c := new(args)
	flag.StringVar(&c.Secret, "secret", "secret", "the JWT secret to use")
	flag.StringVar(&c.UserName, "username", "UserName", "the UserName to use")
	flag.StringVar(&c.UserID, "userid", "UserID", "the UserID to use")
	flag.StringVar(&c.Email, "email", "Email", "the Email to use")
	flag.StringVar(&c.DisplayName, "displayname", "DisplayName", "the DisplayName to use")
	flag.StringVar(&c.Issuer, "issuer", "Issure", "the JWT issuer to use")
	flag.Parse()

	fmt.Println("------------------------------------------------------------------------------")
	fmt.Printf("UserName:\t%s\n", c.UserName)
	fmt.Printf("UserID:\t\t%s\n", c.UserID)
	fmt.Printf("Email:\t\t%s\n", c.Email)
	fmt.Printf("DisplayName:\t%s\n", c.DisplayName)
	fmt.Println("------------------------------------------------------------------------------")
	fmt.Printf("Issuer:\t\t%s\n", c.Issuer)
	fmt.Printf("Secret:\t\t%s\n", c.Secret)
	fmt.Println("------------------------------------------------------------------------------")

	return c
}

func createJWT(a *args) string {
	sites := []site{
		site{
			Name:     "bookmarks",
			URL:      "http://localhost:3003",
			PermList: "User;Admin",
		},
		site{
			Name:     "mydms",
			URL:      "http://localhost:3002",
			PermList: "User;Admin",
		},
		site{
			Name:     "login",
			URL:      "http://localhost:3001",
			PermList: "User;Admin",
		},
		site{
			Name:     "onefrontend",
			URL:      "http://localhost:3000",
			PermList: "User;Admin",
		},

		site{
			Name:     "bookmarks",
			URL:      "https://bookmarks.binggl.net",
			PermList: "User;Admin",
		},
		site{
			Name:     "mydms",
			URL:      "https://mydms.binggl.net",
			PermList: "User;Admin",
		},
		site{
			Name:     "login",
			URL:      "https://login.binggl.net",
			PermList: "User;Admin",
		},
		site{
			Name:     "onefrontend",
			URL:      "https://one.binggl.net",
			PermList: "User;Admin",
		},
	}

	// create the token using the claims of the database
	var siteClaims []string
	for _, s := range sites {
		siteClaims = append(siteClaims, fmt.Sprintf("%s|%s|%s", s.Name, s.URL, s.PermList))
	}
	claims := security.Claims{
		Type:        "login.User",
		DisplayName: a.DisplayName,
		Email:       a.Email,
		UserID:      a.UserID,
		UserName:    a.UserName,
		GivenName:   "Givenname",
		Surname:     "Surname",
		Claims:      siteClaims,
	}
	token, err := security.CreateToken(a.Issuer, []byte(a.Secret), ExpiryDays, claims)
	if err != nil {
		panic(fmt.Sprintf("cannot create token: %v", err))
	}
	return token
}
