# Name of the binary to be created
BINARY_NAME=update-hosts-file

# Path to the directory where the binary will be installed
INSTALL_PATH=${DESTDIR}/usr/bin

# Name of the systemd service to be installed
SYSMDSERVICE_NAME=updatehostsfile.service

# Path to the directory where the systemd sevice will be installed
SYSMD_PATH=${DESTDIR}/usr/lib/systemd/system

# Path to the program directory
PROGRAM_DIR=${DESTDIR}/usr/share/update-hosts-file

# Path to the source with default modules
MODULESDIR_SRC=./modules

# Path to the default preferences file
PREFERENCES_SRC=./preferences

# Autocompletion files
AUTOCOMPLETION_OUT=./autocompletion

BASH_AUTOCOMPLETION_FILE=update-hosts-file
BASH_AUTOCOMPLETION_INSTALL=${DESTDIR}/usr/share/bash-completion/completions

ZSH_AUTOCOMPLETION_FILE=_update-hosts-file
ZSH_AUTOCOMPLETION_INSTALL=${DESTDIR}/usr/share/zsh/site-functions

FISH_AUTOCOMPLETION_FILE=update-hosts-file.fish
FISH_AUTOCOMPLETION_INSTALL=${DESTDIR}/usr/share/fish/vendor_completions.d

# Flags to pass to the go build command
GO_BUILD_FLAGS=-v

.DEFAULT_GOAL := build

.PHONY: clean
clean:
	@echo "====> Removing binary..."
	rm ${BINARY_NAME}
	@echo "====> Removing autocompletion files..."
	rm -rf ${AUTOCOMPLETION_OUT}

.PHONY: deps
deps:
	@echo "====> Installing dependencies..."
	go get -v

.PHONY: build
build: deps
	@echo "====> Building binary..."
	go build ${GO_BUILD_FLAGS}
	strip ${BINARY_NAME}

	mkdir -p ${AUTOCOMPLETION_OUT}
	@echo "====> Building autocompletion file for Bash..."
	./${BINARY_NAME} completion bash > ${AUTOCOMPLETION_OUT}/${BASH_AUTOCOMPLETION_FILE}
	@echo "====> Building autocompletion file for Zsh..."
	./${BINARY_NAME} completion zsh > ${AUTOCOMPLETION_OUT}/${ZSH_AUTOCOMPLETION_FILE}
	@echo "====> Building autocompletion file for Fish..."
	./${BINARY_NAME} completion fish > ${AUTOCOMPLETION_OUT}/${FISH_AUTOCOMPLETION_FILE}

.PHONY: install
install:
	# Binary
	@echo "====> Installing binary"
	mkdir -p ${INSTALL_PATH}
	cp ${BINARY_NAME} ${INSTALL_PATH}
	# Program files
	@echo "====> Installing program files"
	mkdir -p ${PROGRAM_DIR} ${PROGRAM_DIR}/config ${PROGRAM_DIR}/modules/local/enabled ${PROGRAM_DIR}/modules/web/enabled
	cp -r ${MODULESDIR_SRC} ${PROGRAM_DIR}/
	cp ${PREFERENCES_SRC} ${PROGRAM_DIR}/config/
	# Systemd service
	@echo "====> Installing systemd service"
	mkdir -p ${SYSMD_PATH}
	cp ${SYSMDSERVICE_NAME} ${SYSMD_PATH}/
	# Autocompletion
	@echo "====> Installing autocompletion files"
	if [ -d "${BASH_AUTOCOMPLETION_INSTALL}" ]; then \
		cp ${AUTOCOMPLETION_OUT}/${BASH_AUTOCOMPLETION_FILE} ${BASH_AUTOCOMPLETION_INSTALL}/; \
	fi
	if [ -d "${ZSH_AUTOCOMPLETION_INSTALL}" ]; then \
		cp ${AUTOCOMPLETION_OUT}/${ZSH_AUTOCOMPLETION_FILE} ${ZSH_AUTOCOMPLETION_INSTALL}/; \
	fi
	if [ -d "${FISH_AUTOCOMPLETION_INSTALL}" ]; then \
		cp ${AUTOCOMPLETION_OUT}/${FISH_AUTOCOMPLETION_FILE} ${FISH_AUTOCOMPLETION_INSTALL}/; \
	fi

.PHONY: uninstall
uninstall:
	# Binary
	@echo "====> Uninstalling binary"
	rm ${INSTALL_PATH}/${BINARY_NAME}
	# Program directory
	@echo "====> Removing program directory"
	rm -rf ${PROGRAM_DIR}
	# Systemd service
	@echo "====> Uninstalling systemd service"
	rm ${SYSMD_PATH}/${SYSMDSERVICE_NAME}
	# Autocompletion
	@echo "====> Uninstalling autocompletion files"
	rm ${BASH_AUTOCOMPLETION_INSTALL}/${BASH_AUTOCOMPLETION_FILE}
	rm ${ZSH_AUTOCOMPLETION_INSTALL}/${ZSH_AUTOCOMPLETION_FILE}
	rm ${FISH_AUTOCOMPLETION_INSTALL}/${FISH_AUTOCOMPLETION_FILE}
