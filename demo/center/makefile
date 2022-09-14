all:
	nohup go run center.go &

log:
	tail -f nohup.out

clean:
	rm -rf nohup.out

kill:
	lsof -i:8081
