# YouTube Playlist Duplicator

## English

### Description

This Go application connects to the YouTube Data API v3 to duplicate all videos from one playlist into a new playlist. It guides the user through an OAuth2 authentication flow using a local loopback server, lists available playlists, prompts for selecting the source playlist, allows renaming the duplicated playlist, and copies each video into the newly created playlist while preserving order.

### Objective

* Automate the process of cloning a YouTube playlist, preserving video order and metadata.

### Technologies

* Go (1.18+)
* YouTube Data API v3
* OAuth2 (`golang.org/x/oauth2`)
* Environment variables management (`github.com/joho/godotenv`)

### Prerequisites

1. A Google Cloud project with the YouTube Data API v3 enabled.
2. OAuth 2.0 Client ID of type **Desktop app** (Client ID & Client Secret).
3. Go installed (version 1.18 or higher).
4. Terminal with internet access.

### Setup

```bash
# Clone repository
git clone https://github.com/yourusername/yt-playlist-duplicator.git
cd yt-playlist-duplicator

# Create .env file
cat <<EOF > .env
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret
EOF

# Install dependencies
go mod tidy
```

### Usage

```bash
go run .
```

1. A browser window will open for Google authorization. Grant access.
2. Return to the terminal and select the source playlist by number.
3. Enter a name for the duplicated playlist.
4. The application will create the new playlist and copy all videos.

---

## Español

### Descripción

Esta aplicación en Go se conecta a la YouTube Data API v3 para duplicar todos los vídeos de una lista de reproducción en una nueva lista. Guía al usuario mediante un flujo de autenticación OAuth2 usando un servidor local de loopback, lista las playlists disponibles, solicita la selección de la playlist origen, permite renombrar la lista duplicada y copia cada vídeo en la nueva lista manteniendo el orden.

### Objetivo

* Automatizar el proceso de clonar una playlist de YouTube, conservando el orden de los vídeos y sus metadatos.

### Tecnologías

* Go (1.18+)
* YouTube Data API v3
* OAuth2 (`golang.org/x/oauth2`)
* Gestión de variables de entorno (`github.com/joho/godotenv`)

### Requisitos previos

1. Un proyecto en Google Cloud con la YouTube Data API v3 habilitada.
2. OAuth 2.0 Client ID de tipo **Desktop app** (Client ID y Client Secret).
3. Go instalado (versión 1.18 o superior).
4. Terminal con acceso a internet.

### Configuración

```bash
# Clonar repositorio
git clone https://github.com/Vuntae/Go---Youtube-API.git
cd Go---Youtube-API

# Crear archivo .env
cat <<EOF > .env
GOOGLE_CLIENT_ID=tu-client-id
GOOGLE_CLIENT_SECRET=tu-client-secret
EOF

# Instalar dependencias
go mod tidy
```

### Uso

```bash
go run .
```

1. Se abrirá una ventana de navegador para autorizar la aplicación en Google. Concede el acceso.
2. Regresa a la terminal y selecciona la playlist origen por número.
3. Ingresa un nombre para la playlist duplicada.
4. La aplicación creará la nueva lista y copiará todos los vídeos.

---