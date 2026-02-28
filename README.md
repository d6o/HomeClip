# HomeClip

A self-hosted shared clipboard and file transfer tool for your local network. Open it on any device, paste text or drop files — instantly available everywhere.

No accounts. No cloud. No database. Just a single binary.

## Features

- **Shared Clipboard** — One text buffer shared across all devices. Type on your phone, paste on your laptop.
- **File Sharing** — Upload files up to 100 MB via drag-and-drop or file picker. Download from any device on the network.
- **Auto-Save** — Text saves automatically after 1 second of inactivity. No buttons to click.
- **Auto-Cleanup** — Text and files are automatically deleted after 24 hours. Nothing lingers.
- **Zero Config** — Runs out of the box with sane defaults. Two environment variables if you need them.
- **Single Binary** — Frontend is embedded. No Node.js, no npm, no build step.

## Quick Start

### Binary

```sh
go build -o homeclip ./cmd/homeclip
./homeclip
```

Open `http://<your-ip>:8080` on any device.

### Make

```sh
make run              # build and run on port 8080
make run PORT=9090    # use a different port
make clean            # remove binary and data
```

### Docker

```sh
docker run -p 8080:8080 -v homeclip-data:/data ghcr.io/d6o/homeclip:2
```

Or build from source:

```sh
docker build -t homeclip .
docker run -p 8080:8080 -v homeclip-data:/data homeclip
```

### Kubernetes / K3s

```sh
kubectl apply -f k8s.yaml
```

Deploys a single replica with a 1Gi PersistentVolumeClaim.

## Configuration

| Variable  | Default | Description            |
|-----------|---------|------------------------|
| `PORT`    | `8080`  | HTTP listen port       |
| `DATA_DIR`| `/data` | Path to data directory |

## API

| Method   | Endpoint               | Description                    |
|----------|------------------------|--------------------------------|
| `GET`    | `/api/text`            | Get clipboard content          |
| `PUT`    | `/api/text`            | Update clipboard content       |
| `POST`   | `/api/files`           | Upload a file (multipart form) |
| `GET`    | `/api/files`           | List all files                 |
| `GET`    | `/api/files/{filename}`| Download a file                |
| `DELETE` | `/api/files/{filename}`| Delete a file                  |

## Project Structure

```
cmd/homeclip/          Entry point
internal/
  config/              Environment-based configuration
  clipboard/           Text buffer storage (JSON file)
  filestore/           File upload storage (filesystem)
  cleanup/             Periodic 24h expiry cleanup
  server/              HTTP server, routing, embedded frontend
    static/            Single-page frontend (HTML/CSS/JS)
Dockerfile             Multi-stage build, non-root alpine
k8s.yaml               Deployment + Service + PVC
```

## How It Works

HomeClip stores everything on the filesystem under a single data directory:

- **Text** is persisted as a JSON file with content and timestamp.
- **Files** are stored as-is in a subdirectory. Upload timestamps come from file modification times.
- A **cleanup loop** runs every 10 minutes and removes anything older than 24 hours.

There is no database, no authentication, and no encryption — this is designed for trusted local networks.

## License

MIT
