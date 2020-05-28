package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type order struct {
	OrderImage string
	OrderName  string
	OrderPrice string
}

type orderTotals struct {
	OrderTot int     `json:"subtotal"`
	OrderTax float64 `json:"tax"`
	OrderDue float64 `json:"total"`
}

type shippingAndCard struct {
	Address1   string `json:"address1"`
	Address2   string `json:"address2"`
	City       string `json:"city"`
	State      string `json:"state"`
	Zip        string `json:"zip"`
	Country    string `json:"country"`
	Phone      string `json:"phone"`
	Card       string `json:"card"`
	Expiry     string `json:"expiry"`
	Cvv        string `json:"cvv"`
	BillingZip string `json:"billingZip"`
	NameOnCard string `json:"nameOnCard"`
}

var dbOrders []order // slice of orders
var m = map[string]interface{}{}

func checkout(w http.ResponseWriter, r *http.Request) {
	if !alreadyLoggedIn(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		add1 := r.FormValue("address")
		add2 := r.FormValue("address2")
		ct := r.FormValue("city")
		st := r.FormValue("state")
		zp := r.FormValue("zip")
		ctry := r.FormValue("country")
		ph := r.FormValue("phone")
		crd := r.FormValue("cardNumber")
		exp := r.FormValue("expiration")
		cv := r.FormValue("cvv")
		bzip := r.FormValue("billingZip")
		noc := r.FormValue("nameOnCard")

		var s = shippingAndCard{Address1: add1,
			Address2:   add2,
			City:       ct,
			State:      st,
			Zip:        zp,
			Country:    ctry,
			Phone:      ph,
			Card:       crd,
			Expiry:     exp,
			Cvv:        cv,
			BillingZip: bzip,
			NameOnCard: noc,
		}

		m["Details"] = s

		timein := time.Now().Local().Add(time.Hour * time.Duration(168))
		m["Date"] = timein

	}

	tpl.ExecuteTemplate(w, "checkout.html", m)
}

func orderDetail(w http.ResponseWriter, r *http.Request) {
	if !alreadyLoggedIn(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Creating a file
	file, err := os.Create("newOrder.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	jsonData, err := json.Marshal(m)

	if err != nil {
		fmt.Println(err)
		return
	}

	//Writing string on a file
	file.Write(jsonData)
	defer file.Close()

	tpl.ExecuteTemplate(w, "orderDetail.html", m)
}
