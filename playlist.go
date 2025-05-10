package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/api/youtube/v3"
)

// SelectPlaylist muestra un menú con tus playlists y devuelve la ID seleccionada.
func SelectPlaylist(svc *youtube.Service) (string, error) {
	type entry struct{ Title, ID string }
	var all []entry

	// Obtiene la playlist de "Videos que te gustan"
	chResp, err := svc.Channels.List([]string{"contentDetails"}).Mine(true).Do()
	if err != nil {
		return "", err
	}
	likeID := chResp.Items[0].ContentDetails.RelatedPlaylists.Likes
	all = append(all, entry{"Videos que te gustan", likeID})

	// Carga todas las playlists del usuario (paginado)
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

	// Muestra el menú y lee la selección del usuario
	fmt.Println("\nSelecciona la playlist que quieres duplicar:")
	for i, e := range all {
		fmt.Printf("%2d) %s\n", i+1, e.Title)
	}
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

// DuplicatePlaylist copia todos los vídeos de una playlist a otra.
func DuplicatePlaylist(ctx context.Context, svc *youtube.Service, oldID, newID string) {
	nextPage := ""
	for {
		// Obtiene los videos de la playlist original (paginado)
		resp, err := svc.PlaylistItems.List([]string{"snippet"}).
			PlaylistId(oldID).MaxResults(50).PageToken(nextPage).Do()
		if err != nil {
			log.Fatalf("Error obteniendo vídeos: %v", err)
		}
		for _, item := range resp.Items {
			// Inserta cada video en la nueva playlist
			_, err := svc.PlaylistItems.Insert([]string{"snippet"}, &youtube.PlaylistItem{
				Snippet: &youtube.PlaylistItemSnippet{
					PlaylistId: newID,
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
