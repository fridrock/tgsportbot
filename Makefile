migrations:
	goose -dir db/migrations postgres "postgresql://lenchik:cgovno@127.0.0.1:5432/sport_bot?sslmode=disable" up
dropmigrations:
	goose -dir db/migrations postgres "postgresql://lenchik:cgovno@127.0.0.1:5432/sport_bot?sslmode=disable" down
run:
	go build -o tgsportbot && ./tgsportbot