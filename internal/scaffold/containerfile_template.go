package scaffold

import (
	"bytes"
	"text/template"
)

const dockerfileTemplate = `# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS build
WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG PROJECT={{.Name}}
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/${PROJECT}

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates

COPY --from=build /out/app /app/app

EXPOSE 8080
ENTRYPOINT ["/app/app"]
CMD ["serve"]
`

func RenderDockerfile(data TemplateData) (string, error) {
	t, err := template.New("dockerfile").Parse(dockerfileTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

