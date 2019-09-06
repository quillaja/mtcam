INSTALL_DIR = /opt/mtcam
OUTPUT_DIR = .out

build:
	go generate ./version ./cmd/served
	go build ./cmd/scraped
	go build ./cmd/served
	-mkdir $(OUTPUT_DIR)
	mv scraped served $(OUTPUT_DIR)/
	git restore version/version.go

clean:
	rm $(OUTPUT_DIR)/served $(OUTPUT_DIR)/scraped
	rmdir $(OUTPUT_DIR)

install:
	-sudo mkdir $(INSTALL_DIR)
	sudo mv $(OUTPUT_DIR)/served $(OUTPUT_DIR)/scraped $(INSTALL_DIR)/

service-install:
	sudo cp service/* /etc/systemd/system/
	sudo systemctl daemon-reload
	sudo systemctl enable mtcam.target

upgrade:
	git pull
	-git describe
	make build
	sudo systemctl stop mtcam.target
	make install
	sudo systemctl start mtcam.target

uninstall:
	sudo systemctl stop mtcam.target
	sudo systemctl disable mtcam.target
	sudo rm /etc/systemd/system/mtcam.target
	sudo rm /etc/systemd/system/scraped.service
	sudo rm /etc/systemd/system/served.service
	sudo systemctl daemon-reload
	sudo rm $(INSTALL_DIR)/scraped $(INSTALL_DIR)/served