install:
	$($SHELL which go build)
	chmod +x nvcat
	cp nvcat /usr/local/bin
