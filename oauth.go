package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
	// ... imports ...
)

// tokenCacheFile devuelve la ruta donde guardaremos el token.
func tokenCacheFile() string {
	// Obtiene el directorio home del usuario y retorna la ruta del archivo de token.
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Imposible leer home dir: %v", err)
	}
	return dir + "/.youtube_token.json"
}

// saveToken serializa el token a disco.
func saveToken(path string, token *oauth2.Token) {
	// Guarda el token OAuth2 en un archivo local.
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("No pude crear el archivo de token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// tokenFromFile lee y deserializa el token guardado.
func tokenFromFile(path string) (*oauth2.Token, error) {
	// Lee el token desde el archivo y lo deserializa.
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var token oauth2.Token
	err = json.Unmarshal(b, &token)
	return &token, err
}

// GetClient realiza el flujo OAuth y devuelve un servicio de YouTube autenticado.
func GetClient(ctx context.Context) *youtube.Service {
	// Carga las variables de entorno desde .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error cargando .env:", err)
	}

	// Configura OAuth2 con credenciales y scopes de YouTube.
	conf := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "http://127.0.0.1:8080/oauth2callback",
		Scopes:       []string{"https://www.googleapis.com/auth/youtube"},
		Endpoint:     google.Endpoint,
	}

	cacheFile := tokenCacheFile()
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		// Si no hay token, inicia el flujo de autorización OAuth2.
		codeCh := make(chan string)
		http.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
			code := r.URL.Query().Get("code")
			fmt.Fprintln(w, "¡Listo! Puedes cerrar esta pestaña.")
			go func() { codeCh <- code }()
		})

		srv := &http.Server{Addr: ":8080"}
		go srv.ListenAndServe()

		authURL := conf.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		OpenBrowser(authURL)
		fmt.Println("Abre tu navegador para autorizar la aplicación...")

		code := <-codeCh
		srv.Shutdown(context.Background())

		tok, err = conf.Exchange(ctx, code)
		if err != nil {
			log.Fatalf("Exchange de token falló: %v", err)
		}
		saveToken(cacheFile, tok)
	}

	// Crea el cliente autenticado y el servicio de YouTube.
	client := conf.Client(ctx, tok)
	svc, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creando el servicio YouTube: %v", err)
	}
	return svc
}
