FROM golang:1.24-alpine AS builder

WORKDIR /src/app

#Curl is used for testing
#RUN apk --no-cache add curl 

COPY ["go.mod","go.sum","./"]

RUN go mod download

COPY . .

RUN go build -o ./bin/app cmd/urlshortner/main.go

FROM alpine AS runner

#Copy bin
COPY --from=builder /src/app/bin/app /

#Copy config
COPY config/local.yaml config/local.yaml

#Copy .env file env vars may be used here
COPY .env /

#Copy db
COPY database/database.db database/database.db

CMD [ "/app" ]