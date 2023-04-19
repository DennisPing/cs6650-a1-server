# Twinder Server

## Getting Started

### Build
```
make
```

### Run
```
./server
```

## API Endpoints

### POST swipe
```
/swipe/{leftorright}/

{
    "swipe": string
    "swipee": string
    "comment": string
}
```

## Build and Test Docker image
```
docker build -t server:a1 .
```

```
docker network create my_network
docker run --name server --network my_network -p 8080:8080 -e PORT=8080 -e LOG_LEVEL=Warn server:a1
```

## Tag and push Docker image to Google Container Registry
```
docker tag server:a1 gcr.io/cs6650-dping/server:a1
docker push gcr.io/cs6650-dping/server:a1
```

## Deploy new Cloud Run service
```
gcloud run deploy server \
    --image gcr.io/cs6650-dping/server:a1 \
    --platform managed \
    --region us-central1 \
    --allow-unauthenticated \
    --set-env-vars="LOG_LEVEL=Info"
```

Record the URL of the Cloud Run service for the server!
Eg. https://hello-server-[random]-uc.a.run.app
