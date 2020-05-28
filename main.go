package main

import (
	"html/template"
	"net/http"

	"strconv"

	"math"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type user struct {
	Name       string
	Email      string
	Password   []byte
	AgreeTerms string
}

var dbUsers = map[string]user{}      // user ID, user
var dbSessions = map[string]string{} // session ID, user ID

var tpl *template.Template

var gTotal int = 0
var gTax float64
var gDue float64

//var t orderTotals

func init() {
	tpl = template.Must(template.ParseGlob("templates/*.html"))
}

func index(w http.ResponseWriter, r *http.Request) {
	u := getUser(w, r)

	if !alreadyLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Add to cart logic
	if r.Method == http.MethodPost {

		// Get values
		pimg := r.FormValue("productImage")
		pname := r.FormValue("productName")
		pprice := r.FormValue("productPrice")

		// append to a slice here
		s := order{OrderImage: pimg,
			OrderName:  pname,
			OrderPrice: pprice,
		}
		dbOrders = append(dbOrders, s)

		m["Cart"] = dbOrders
		// Add to cart logic ends

		//var totalSlice []float64

		total, err := strconv.Atoi(pprice)
		if err == nil {
			/*
				gTotal = gTotal + total
				totalSlice = append(totalSlice, float64(gTotal))
				totalSlice = append(totalSlice, 0.0625*float64(gTotal))
				tot := float64(gTotal) + (0.0625 * float64(gTotal))
				totalSlice = append(totalSlice, tot)
			*/
			gTotal = gTotal + total
			gTax = 0.0625 * float64(gTotal)
			gTax = math.Round(gTax*100) / 100
			gDue = float64(gTotal) + gTax
			gDue = math.Round(gDue*100) / 100
		}

		var t = orderTotals{OrderTot: gTotal,
			OrderTax: gTax,
			OrderDue: gDue,
		}

		m["Totals"] = t

	}

	tpl.ExecuteTemplate(w, "index.html", u)

}

func signUp(w http.ResponseWriter, r *http.Request) {

	if alreadyLoggedIn(r) {
		http.Redirect(w, r, "/index", http.StatusSeeOther)
		return
	}

	// Process form submission
	if r.Method == http.MethodPost {

		// Get form values
		fn := r.FormValue("name")
		em := r.FormValue("email")
		pw := r.FormValue("pass")
		rpw := r.FormValue("re_pass")
		at := r.FormValue("agree-term")

		// Email already registered?
		if _, ok := dbUsers[em]; ok {
			http.Error(w, "Email already registered!", http.StatusForbidden)
			return
		}

		if fn == "" {
			http.Error(w, "Please enter your name!", http.StatusForbidden)
			return
		}

		if pw == "" || rpw == "" {
			http.Error(w, "Please enter password!", http.StatusForbidden)
			return
		}

		// Password did not match
		if pw != rpw {
			http.Error(w, "Password did not match!", http.StatusForbidden)
			return
		}

		// Agree Terms and Conditions
		if at == "" {
			http.Error(w, "Please agree Terms and Conditions to continue!", http.StatusForbidden)
			return
		}

		// No need to create session after sign up. Create only in login!
		/*
			// Create session
			sID, _ := uuid.NewV4()
			c := &http.Cookie{
				Name:  "session",
				Value: sID.String(),
			}
			http.SetCookie(w, c)
			dbSessions[c.Value] = em
		*/

		// Store user in dbUser
		bs, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		u := user{fn, em, bs, at}
		dbUsers[em] = u

		// Redirect
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return

	}

	tpl.ExecuteTemplate(w, "signup.html", nil)
}

func logIn(w http.ResponseWriter, r *http.Request) {
	if alreadyLoggedIn(r) {
		http.Redirect(w, r, "/index", http.StatusSeeOther)
		return
	}

	// Process form submission
	if r.Method == http.MethodPost {
		em := r.FormValue("email")
		pw := r.FormValue("your_pass")

		//  Is there that email?
		u, ok := dbUsers[em]
		if !ok {
			http.Error(w, "Username or/and password do not match!", http.StatusForbidden)
			return
		}

		// Is password correct?
		err := bcrypt.CompareHashAndPassword(u.Password, []byte(pw))
		if err != nil {
			http.Error(w, "Username or/and password do not match!", http.StatusForbidden)
			return
		}

		// Create session
		sID, _ := uuid.NewV4()
		c := &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}
		http.SetCookie(w, c)
		dbSessions[c.Value] = em
		http.Redirect(w, r, "/index", http.StatusSeeOther)
		return
	}

	tpl.ExecuteTemplate(w, "login.html", nil)
}

func logout(w http.ResponseWriter, r *http.Request) {
	if !alreadyLoggedIn(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	c, _ := r.Cookie("session")

	// Delete the session
	delete(dbSessions, c.Value)

	// Remove the cookie
	c = &http.Cookie{
		Name:   "session",
		Value:  "",
		MaxAge: -1,
	}

	http.SetCookie(w, c)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {

	fs := http.FileServer(http.Dir("templates"))
	http.Handle("/images/", fs)

	http.Handle("/favicon.ico", http.NotFoundHandler())

	http.HandleFunc("/", logIn)
	http.HandleFunc("/signup", signUp)
	http.HandleFunc("/index", index)
	http.HandleFunc("/checkout", checkout)
	http.HandleFunc("/orderDetail", orderDetail)
	http.HandleFunc("/logout", logout)

	http.ListenAndServe(":8080", nil)
}
