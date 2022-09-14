all:
	nohup go run client.go &

log:
	tail -f nohup.out

clean:
	rm -rf nohup.out

kill:
	lsof -i:8081
