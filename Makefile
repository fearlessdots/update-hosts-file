# Name of the binary to be created
BINARY_NAME=update-hosts-file

# Path to the directory where the binary will be installed
INSTALL_PATH=${DESTDIR}/usr/bin

# Name of the systemd service to be installed
SYSMDSERVICE_NAME=updatehostsfile.service

# Path to the directory where the systemd sevice will be installed
SYSMD_PATH=${DESTDIR}/usr/lib/systemd/system/

# Path to the program directory
PROGRAM_DIR=${DESTDIR}/usr/share/update-hosts-file

# Path to the source with default modules
MODULESDIR_SRC=./modules

# Path to the default preferences file
PREFERENCES_SRC=./preferences

# Flags to pass to the go build command
GO_BUILD_FLAGS=-v

.DEFAULT_GOAL := build

.PHONY: deps
deps:
	go get -v

.PHONY: build
build: deps
	go build ${GO_BUILD_FLAGS}
	strip ${BINARY_NAME}

.PHONY: install
install:
	mkdir -p ${INSTALL_PATH}
	cp ${BINARY_NAME} ${INSTALL_PATH}
	mkdir -p ${PROGRAM_DIR} ${PROGRAM_DIR}/config ${PROGRAM_DIR}/modules/local/enabled ${PROGRAM_DIR}/modules/web/enabled
	cp -r ${MODULESDIR_SRC} ${PROGRAM_DIR}/
	cp ${PREFERENCES_SRC} ${PROGRAM_DIR}/config/
	mkdir -p ${SYSMD_PATH}
	cp ${SYSMDSERVICE_NAME} ${SYSMD_PATH}/

.PHONY: uninstall
uninstall:
	rm ${INSTALL_PATH}/${BINARY_NAME}
	rm -rf ${PROGRAM_DIR}
	rm ${SYSMD_PATH}/${SYSMDSERVICE_NAME}
