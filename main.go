package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		fmt.Println("Por favor, abre esta URL manualmente:", url)
		return
	}
	exec.Command(cmd, args...).Start()
}

// tokenCacheFile devuelve la ruta donde guardaremos el token.
func tokenCacheFile() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Imposible leer home dir: %v", err)
	}
	return dir + "/.youtube_token.json"
}

// saveToken serializa el token a disco.
func saveToken(path string, token *oauth2.Token) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("No pude crear el archivo de token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// tokenFromFile lee y deserializa el token guardado.
func tokenFromFile(path string) (*oauth2.Token, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var token oauth2.Token
	err = json.Unmarshal(b, &token)
	return &token, err
}

// getClient hace el flujo de OAuth en consola y devuelve un *youtube.Service autenticado.
// getClient inicia OAuth2 con cache de token y servidor local de callback.
func getClient(ctx context.Context) *youtube.Service {
	// 1. Carga .env y crea la configuración OAuth2
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error cargando .env:", err)
	}
	conf := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "http://127.0.0.1:8080/oauth2callback",
		Scopes:       []string{"https://www.googleapis.com/auth/youtube"},
		Endpoint:     google.Endpoint,
	}

	cacheFile := tokenCacheFile()

	// 2. Intentar cargar token de archivo
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		// 3. Si no hay token cache, hacer el flujo completo
		codeCh := make(chan string)
		http.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
			code := r.URL.Query().Get("code")
			fmt.Fprintln(w, "¡Listo! Puedes cerrar esta pestaña.")
			go func() { codeCh <- code }()
		})

		srv := &http.Server{Addr: ":8080"}
		go func() {
			if err := srv.ListenAndServe(); err != http.ErrServerClosed {
				log.Fatalf("El servidor de callback falló: %v", err)
			}
		}()

		authURL := conf.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		openBrowser(authURL)
		fmt.Println("Abre tu navegador para autorizar la aplicación...")

		// Espera el código de autorización
		code := <-codeCh

		// Cierra el servidor
		ctxShut, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctxShut)

		// Intercambia el código por el token
		tok, err = conf.Exchange(ctx, code)
		if err != nil {
			log.Fatalf("Exchange de token falló: %v", err)
		}
		// Guarda el token para la próxima ejecución
		saveToken(cacheFile, tok)
	}

	// 4. Crea un cliente HTTP que refresca automáticamente el token
	client := conf.Client(ctx, tok)
	svc, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creando el servicio YouTube: %v", err)
	}
	return svc
}

// selectPlaylist lista todas tus playlists numeradas y devuelve la ID de la escogida.
func selectPlaylist(svc *youtube.Service) (string, error) {
	type entry struct{ Title, ID string }
	var all []entry

	// 0) Agregar “Videos que te gustan” al menú:
	chResp, err := svc.Channels.List([]string{"contentDetails"}).Mine(true).Do()
	if err != nil {
		return "", err
	}
	likeID := chResp.Items[0].ContentDetails.RelatedPlaylists.Likes
	all = append(all, entry{"Videos que te gustan", likeID})

	// 1) Cargar playlists propias:
	pageToken := ""
	for {
		resp, err := svc.Playlists.List([]string{"snippet"}).
			Mine(true).MaxResults(50).PageToken(pageToken).Do()
		if err != nil {
			return "", err
		}
		for _, pl := range resp.Items {
			all = append(all, entry{pl.Snippet.Title, pl.Id})
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	// 2) Mostrar menú numerado:
	fmt.Println("\nSelecciona la playlist que quieres duplicar:")
	for i, e := range all {
		fmt.Printf("%2d) %s\n", i+1, e.Title)
	}

	// 3) Leer elección:
	var choice int
	fmt.Print("Número: ")
	if _, err := fmt.Scan(&choice); err != nil {
		return "", err
	}
	if choice < 1 || choice > len(all) {
		return "", fmt.Errorf("fuera de rango")
	}

	sel := all[choice-1]
	fmt.Printf("Has seleccionado: %q (ID: %s)\n\n", sel.Title, sel.ID)
	return sel.ID, nil
}

func main() {
	ctx := context.Background()
	svc := getClient(ctx)
	// 0) seleccionamos la playlist a copiar
	oldPlaylistID, err := selectPlaylist(svc)
	if err != nil {
		log.Fatalf("Error al seleccionar playlist: %v", err)
	}

	// 1) Pedir el nombre de la nueva playlist
	fmt.Print("Ingresa el nombre de la nueva playlist: ")
	var newPlaylistName string
	if _, err := fmt.Scan(&newPlaylistName); err != nil {
		log.Fatalf("Error leyendo nombre de playlist: %v", err)
	}
	if newPlaylistName == "" {
		log.Fatal("El nombre de la playlist no puede estar vacío.")
	}

	// 2) Crear nueva playlist
	newPl, err := svc.Playlists.Insert([]string{"snippet", "status"}, &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			Title:       newPlaylistName,
			Description: "PLaylist Duplicada desde Go",
		},
		Status: &youtube.PlaylistStatus{PrivacyStatus: "private"},
	}).Do()
	if err != nil {
		log.Fatalf("Error creando nueva playlist: %v", err)
	}
	newPlaylistID := newPl.Id
	fmt.Println("+ Nueva playlist creada con ID:", newPlaylistID)

	// 2) Copiar vídeos de oldPlaylistID a newPlaylistID
	nextPage := ""
	for {
		resp, err := svc.PlaylistItems.List([]string{"snippet"}).
			PlaylistId(oldPlaylistID).
			MaxResults(50).
			PageToken(nextPage).
			Do()
		if err != nil {
			log.Fatalf("Error obteniendo vídeos: %v", err)
		}
		for _, item := range resp.Items {
			_, err := svc.PlaylistItems.Insert([]string{"snippet"}, &youtube.PlaylistItem{
				Snippet: &youtube.PlaylistItemSnippet{
					PlaylistId: newPlaylistID,
					ResourceId: &youtube.ResourceId{
						Kind:    "youtube#video",
						VideoId: item.Snippet.ResourceId.VideoId,
					},
				},
			}).Do()
			if err != nil {
				log.Printf("Error copiando video %s: %v", item.Snippet.ResourceId.VideoId, err)
			}
		}
		if resp.NextPageToken == "" {
			break
		}
		nextPage = resp.NextPageToken
	}

	fmt.Println("+ Todos los vídeos han sido duplicados.")
}
