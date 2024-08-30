package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/text/message"
)

var formatter = message.NewPrinter(message.MatchLanguage("en"))

func main() {
	godotenv.Load()
	port, def := os.LookupEnv("PORT")
	if !def {
		port = "8080"
	}
	gotenbergBaseUrl, def := os.LookupEnv("GOTENBERG_BASE_URL")
	if !def {
		log.Panic("GOTENBERG_BASE_URL not defined.")
	}

	funcs := template.FuncMap{
		"formatCurrency": func(cents int) string {
			return formatter.Sprintf("€%.2f", float64(cents)/100)
		},
		"formatDate": func(t time.Time) string {
			return t.Format("January 2, 2006")
		},
	}
	invoiceTemplate := template.Must(template.New("invoice.html").Funcs(funcs).ParseFiles("invoice.html"))

	router := http.NewServeMux()
	router.HandleFunc("POST /invoice", func(w http.ResponseWriter, r *http.Request) {
		// Parse request body and compute total price field
		var invoice Invoice
		err := json.NewDecoder(r.Body).Decode(&invoice)
		if err != nil {
			log.Print("POST /invoice: could not parse request body.")
			return
		}
		invoice.TotalInCents = 0
		for _, v := range invoice.Items {
			invoice.TotalInCents += v.UnitPriceInCents * v.Quantity
		}

		// Generate Gotenbert FormData
		reqBody := bytes.Buffer{}
		writer := multipart.NewWriter(&reqBody)
		part, err := writer.CreateFormFile("file", "index.html")
		if err != nil {
			log.Print("POST /invoice: could not create form data field.")
			log.Print(err)
			return
		}

		// Generate HTML file
		if err := invoiceTemplate.Execute(part, invoice); err != nil {
			log.Print("POST /invoice: could not render template.")
			log.Print(err)
			return
		}
		// Necessary so the writer.FormDataContentType() returns the correct boundary
		writer.Close()

		// Generate PDF with Gotenberg
		endpoint, err := url.JoinPath(gotenbergBaseUrl, "/forms/chromium/convert/html")
		if err != nil {
			log.Fatal("Could not build Gotenberg endpoint URL.")
		}
		resp, err := http.Post(
			endpoint,
			writer.FormDataContentType(),
			&reqBody,
		)
		if err != nil {
			log.Print("POST /invoice: could not generate PDF.")
			log.Print(err)
			return
		}

		// Return PDF
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Print("POST /invoice: could not read response from Gotenberg.")
			log.Print(err)
			return
		}

		if resp.StatusCode != 200 {
			log.Printf("Call to Gotenberg faild with status: %v.", resp.Status)
			log.Print(string(respBody))
			w.WriteHeader(resp.StatusCode)
			return
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Write(respBody)
		log.Printf("Generated invoice %s successfully.", invoice.Id)
	})

	server := http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	log.Printf("Listening on port %v.", port)
	log.Fatal(server.ListenAndServe())
}
