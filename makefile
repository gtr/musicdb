frontend:
	go build -o frontend frontend.go album.go parse.go message.go

backend:
	go build -o backend backend.go album.go parse.go message.go

log: 
	go build -o log cmdlog.go album.go

clean:
	go clean
