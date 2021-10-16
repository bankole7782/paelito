build:
	rm -rf bin
	mkdir -p bin

	go build -o bin/paelito_maker ./maker
	go build -o bin ./paelito

	cp paelito.desktop bin/paelito.desktop
	cp paelito.png bin/paelito.png
