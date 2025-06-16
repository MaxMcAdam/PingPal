build:
	sudo setcap cap_net_raw+ep pingpal
	go build -o pingpal main.go 