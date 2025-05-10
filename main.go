package main

import (
	"context"
	"fmt"
	"log"

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
	var name string
	if _, err := fmt.Scan(&name); err != nil {
		log.Fatalf("Error leyendo nombre de playlist: %v", err)
	}
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
