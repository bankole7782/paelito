build:
	go build -o bin\paelito_maker.exe .\maker
	go build -ldflags="-H windowsgui" -o bin\paelito.exe .\paelito
