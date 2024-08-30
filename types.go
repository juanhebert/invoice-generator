package main

import (
	"time"
)

type Invoice struct {
	Id             string        `json:"id"`
	IssuedAt       time.Time     `json:"issuedAt"`
	DueAt          time.Time     `json:"dueAt"`
	Provider       Company       `json:"provider"`
	Client         Company       `json:"client"`
	Items          []InvoiceItem `json:"items"`
	TotalInCents   int
	PaymentDetails PaymentDetails `json:"paymentDetails"`
}

type Company struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Address1 string `json:"address1"`
	Address2 string `json:"address2"`
	Country  string `json:"country"`
	IdType   string `json:"idType"`
}

type InvoiceItem struct {
	Title            string `json:"title"`
	UnitPriceInCents int    `json:"unitPriceInCents"`
	Quantity         int    `json:"quantity"`
}

type PaymentDetails struct {
	AccountHolder string `json:"accountHolder"`
	Iban          string `json:"iban"`
	Bic           string `json:"bic"`
}
