INSTALL_DIR = /opt/mtcam

build:
	go generate ./version ./cmd/served
	go build ./cmd/scraped
	go build ./cmd/served

install:
	-sudo mkdir $(INSTALL_DIR)
	sudo mv scraped served $(INSTALL_DIR)/
	sudo chown root $(INSTALL_DIR)/scraped $(INSTALL_DIR)/served

service-install:
	sudo cp ./service/* /etc/systemd/system/
	sudo systemctl daemon-reload
	sudo systemctl enable mtcam.target