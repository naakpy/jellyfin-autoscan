# jellyfin-autoscan

Jellyfin-autoscan is a simple tool that automatically triggers a scan of your Jellyfin library.

This tool is only necessary if you are using Jellyfin with a mounted NFS share.  
Otherwise, just use Jellyfin's built-in autoscan feature.

## How It Works

This tool triggers a Jellyfin task via the Jellyfin API to refresh the library whenever Radarr or Sonarr moves or renames a file.

## Getting Your Jellyfin API Key

To obtain your API key from Jellyfin:

1. Log in to your Jellyfin server.
2. Go to **Dashboard** → **API Keys**.
3. Click **+ New API Key**.
4. Give it a name (e.g., `jellyfin-autoscan`) and click **Save**.
5. Copy the generated API key and use it in the configuration.

## Usage

You can run jellyfin-autoscan either using Docker/Docker Compose or build it from source.

### Running from Source

1. Clone the repository:

```bash
git clone https://github.com/naakpy/jellyfin-autoscan.git
cd jellyfin-autoscan
```

1. Build the application:

```bash
go build
```

1. Set the environment variables:

```bash
export JELLYFIN_BASE_URL=http://localhost:8096
export JELLYFIN_API_KEY=your_api_key_here
export LOG_LEVEL=INFO
```

1. Run the application:

```bash
./jellyfin-autoscan
```

### Using Docker

Start the service using Docker or Docker Compose.  
Don't forget to set the required environment variables.

### Example `docker run` command

```bash
docker run -d --name jellyfin-autoscan \
  -e JELLYFIN_BASE_URL=http://localhost:8096 \
  -e JELLYFIN_API_KEY=your_api_key_here \
  -e LOG_LEVEL=INFO \
  naakpy/jellyfin-autoscan
```

### Example `docker-compose.yml`

```yaml
services:
  jellyfin-autoscan:
    image: naakpy/jellyfin-autoscan
    environment:
      - JELLYFIN_BASE_URL=http://localhost:8096
      - JELLYFIN_API_KEY=your_api_key_here
      - LOG_LEVEL=INFO
```

Then start the service with:

```bash
docker compose up -d
```

## Radarr / Sonarr Integration

To set up the webhook in Radarr or Sonarr:

1. Go to **Settings** → **Connect** → **Add a webhook**.  
   ![Sidebar](./docs/sidebar.png)

1. Enable the following events:

   - **On file import**
   - **On file rename**
   - **On file upgrade**

1. Set the **Webhook URL** to the address of your `jellyfin-autoscan` service.

1. Set the **Method** to `POST`.  
   ![Radarr webhook](./docs/radarr.png)

This setup ensures that Jellyfin rescans your library automatically when Radarr or Sonarr updates media files.
