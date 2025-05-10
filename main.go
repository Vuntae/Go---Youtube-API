package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/api/youtube/v3"
)

func main() {
	ctx := context.Background()
	svc := GetClient(ctx)

	// Seleccionar playlist a copiar
	oldID, err := SelectPlaylist(svc)
	if err != nil {
		log.Fatalf("Error al seleccionar playlist: %v", err)
	}

	// Pedir nombre para la nueva playlist
	fmt.Print("Ingresa el nombre de la nueva playlist: ")
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n') // Descarta todo hasta el próximo salto de línea
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error leyendo nombre de playlist: %v", err)
	}

	// Elimina primero el salto de línea y luego los espacios
	name = strings.TrimRight(name, "\r\n")
	name = strings.TrimSpace(name)

	if name == "" {
		log.Fatal("El nombre de la playlist no puede estar vacío.")
	}

	// Crear nueva playlist
	newPl, err := svc.Playlists.Insert([]string{"snippet", "status"}, &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			Title:       name,
			Description: "Playlist duplicada desde Go",
		},
		Status: &youtube.PlaylistStatus{PrivacyStatus: "private"},
	}).Do()
	if err != nil {
		log.Fatalf("Error creando nueva playlist: %v", err)
	}
	fmt.Println("+ Nueva playlist creada con ID:", newPl.Id)

	// Duplicar vídeos
	DuplicatePlaylist(ctx, svc, oldID, newPl.Id)
}
