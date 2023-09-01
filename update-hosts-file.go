package main

// Default preferences file:

//# Full path to default editor (nano, vim, emacs, micro,...)
//DEFAULT_EDITOR /usr/bin/nano
//# Full path to default viewer (less, cat, batcat,...)
//DEFAULT_VIEWER /usr/bin/less
//# Maximum number of files in backup directory
//MAX_BACKUP_FILES 10
//# Tell the program to skip module (and to not restore backup) if any web module source can not be reached
//KEEP_ON_HOST_UNREACHABLE false
//# IP or hostname for remote server to test internet connection
//IP_TEST=8.8.8.8

import (
	// Modules in GOROOT
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"errors"
	"log"
	"net/http"
	"math/rand"
	"encoding/base64"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"path/filepath"
	"runtime"

	// External modules
	cobra "github.com/spf13/cobra"
	color "github.com/gookit/color"
	survey "github.com/AlecAivazis/survey/v2"
	terminal "golang.org/x/crypto/ssh/terminal"

	// Unused modules
	_"runtime/debug"
	_"crypto/rand"
)

//
//// CONFIGURATION VARIABLES
//

var (
	programName = "UpdateHostsFile"
	programNameAscii = `
  _   _           _       _       _   _           _       _____ _ _
 | | | |_ __   __| | __ _| |_ ___| | | | ___  ___| |_ ___|  ___(_) | ___
 | | | | '_ \ / _  |/ _  | __/ _ \ |_| |/ _ \/ __| __/ __| |_  | | |/ _ \
 | |_| | |_) | (_| | (_| | ||  __/  _  | (_) \__ \ |_\__ \  _| | | |  __/
  \___/| .__/ \__,_|\__,_|\__\___|_| |_|\___/|___/\__|___/_|   |_|_|\___|
       |_|
`
	programDir      = "/usr/share/update-hosts-file"
	programVersion	= "0.1.1"
	programShortDescription = "A program to manage and update your /etc/hosts file with custom blocklists, both local and web sourced"
	programLongDescription = `The UpdateHostsFile program is a command-line utility designed to provide users with an efficient and effective method of updating their hosts file. With the ability to leverage a variety of different sources, including local and web-based modules, users can quickly and easily update their hosts file with the most up-to-date information.`
	modulesDir      = programDir + "/modules"
	localModulesDir = modulesDir + "/local"
	webModulesDir   = modulesDir + "/web"
	configDir       = programDir + "/config"
	backupDir       = programDir + "/backup"
	hostsFile       = "/etc/hosts"
)

//
//// DISPLAY FUNCTIONS
//

func showText(msg string) {
	fmt.Println(msg)
}

func showInfo(msg string) {
	grayHex := "#808080"
	gray := color.HEX(grayHex)
	gray.Println(msg)
}

func showAttention(msg string) {
	orangeHex := "#ffa860"
	orange := color.HEX(orangeHex)
	orange.Println(msg)
}

func showInfoSectionTitle(msg string) {
	grayHex := "#c8c4a9"
	gray := color.HEX(grayHex)
	gray.Println(msg)
}

func showSuccess(msg string) {
	blueHex := "#55aaff"
	blue := color.HEX(blueHex)
	blue.Println(msg)
}

func showError(msg string) {
	redHex := "#ff5050"
	red := color.HEX(redHex)
	red.Println(msg)
}

func hr(char string, factor float64) {
	terminalDimensions := getTerminalDimensions()

	horizontalLine := strings.Repeat(string(char), int(float64(terminalDimensions.width)*factor))
	showText(horizontalLine + "\r")
}

func space() {
	fmt.Println("")
}

func displayProgramInfo() {
	lightCopperHex := "#ffaa7f"
	lightCopper := color.HEX(lightCopperHex)
	greenHex := "#55ff7f"
	green := color.HEX(greenHex)

	showText(programNameAscii)

	showText("Version: " + green.Sprintf(programVersion))

	space()

	showText("Running on " + lightCopper.Sprintf(runtime.GOOS + "/" + runtime.GOARCH) + ". Built with " + runtime.Version() + " using " + runtime.Compiler + " as compiler.")
}

//
//// TERMINAL MANAGEMENT
//

type terminalDimensions struct {
	height int
	width int
}

func getTerminalDimensions() (terminalDimensions) {
	// Get the file descriptor for the standard output
	fd := int(os.Stdout.Fd())

	// Check if the file descriptor is associated with a terminal
	if !terminal.IsTerminal(fd) {
		return terminalDimensions{}
	}

	// Retrieve the terminal size
	width, height, err := terminal.GetSize(fd)
	if err != nil {
		return terminalDimensions{}
	}

	return terminalDimensions{height: height, width: width}
}

//
//// FILE EDITING FUNCTIONS
//

func insertLine(filePath string, line string) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		showError("    > Error: failed to open file: " + err.Error())
		return
	}
	defer file.Close()

	_, err = fmt.Fprintln(file, line)
	if err != nil {
		showError("    > Error: failed to write to file: " + err.Error())
		return
	}
}

func insertHost(file string, ip string, hostname string) {
	insertLine(file, fmt.Sprintf("%s %s", ip, hostname))
}

func insertComment(file string, comment string) {
	insertLine(file, fmt.Sprintf("# %s", comment))
}

//
//// COMPLEMENTARY FUNCTIONS
//

func finishProgram(code int) {
	os.Exit(code)
}

func getConfigValue(key string) (string, error) {
	configFilePath := programDir + "/config/preferences"
	configFile, err := os.Open(configFilePath)
	if err != nil {
		return "", errors.New("Failed to open config file")
	}
	defer configFile.Close()

	scanner := bufio.NewScanner(configFile)
  for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") && strings.HasPrefix(line, key) {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				return parts[1], nil
				} else {
					return "", errors.New("Config key found but value is missing")
				}
			}
		}

	return "", errors.New("Config key not found")
}

func getCurrentHostname() string {
	hostname, _ := os.Hostname()

	return hostname
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to download file: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}


//
//// MAIN FUNCTIONS
//

func insertHostname(tmphosts_file string) {
	showInfoSectionTitle("Inserting the hostname")
	insertLine(tmphosts_file,"")
	insertLine(tmphosts_file,"# Hostname")
	insertHost(tmphosts_file,"127.0.0.1", getCurrentHostname())
	insertLine(tmphosts_file,"")
	showSuccess("    > Done")
}

func restoreBackup(backupFile backup_file) {
	fmt.Println("")
	showAttention("An error has occurred. Backup will be restored.")

	err := os.Rename(filepath.Join(backupDir, backupFile.filename), hostsFile)
	if err != nil {
		showError("    > Error: failed to restore backup: " + err.Error())
		finishProgram(1)
	}

	showInfo("    > Hosts file restored.")
	finishProgram(1)
}


func loadLocalModules(tmphosts_file string) error {
	showInfoSectionTitle("Loading local modules")
	time.Sleep(2 * time.Second)

	enabledLocalModules, err := ioutil.ReadDir(localModulesDir + "/enabled")
	if err != nil {
		return errors.New(fmt.Sprintf("    > Error: failed to read local modules directory: " + err.Error()))
	}

	if len(enabledLocalModules) == 0 {
		showAttention("    > No module enabled")
	}

	i := 0
	for _, module := range enabledLocalModules {
	if i >= 1 {
		fmt.Println("")
	}

	insertLine(tmphosts_file,"")
	insertComment(tmphosts_file,fmt.Sprintf("Hosts from local module '%s'",module.Name()))
	showInfo(fmt.Sprintf("    > Loading local modules from '%s' ",module.Name()))

	file, err := os.Open(localModulesDir + "/enabled/" + module.Name())
	if err != nil {
		showError("        > Error: failed to open local module file: " + err.Error())
		continue
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") && line != "" {
			parts := strings.Fields(line)
			ipAddress := parts[0]
			hostname := parts[1]
			insertHost(tmphosts_file,ipAddress, hostname)
		}
	}

	file.Close()
	insertLine(tmphosts_file,"")

	showSuccess("        > Done")
	i++
	}

	return nil
}

func readWebModuleFile(filePath string) (string, error) {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func loadWebModules(tmphosts_file string, tmpDir string) error {
	showInfoSectionTitle("Loading hosts from selected web sources")
	time.Sleep(2 * time.Second)

	enabledWebModules, err := ioutil.ReadDir(filepath.Join(programDir, "modules", "web", "enabled"))
	if err != nil {
		return errors.New(fmt.Sprintf("    > Error reading enabled web modules: " + err.Error()))
	}

	if len(enabledWebModules) == 0 {
		showAttention("    > No module enabled")
	}

	i := 0
	for _, module := range enabledWebModules {
		if i >= 1 {
			fmt.Println("")
		}

		showInfo(fmt.Sprintf("    > Loading module %s",module.Name()))

		moduleSource, err := readWebModuleFile(filepath.Join(programDir, "modules", "web", "enabled",module.Name()))
		moduleSource = strings.TrimRight(moduleSource, "\n")
		showInfo(fmt.Sprintf("        > Source: %s", moduleSource))
		if err != nil {
			showAttention("        > Error getting module source for "+module.Name()+": "+err.Error())
			continue
		}

		moduleTempFile := filepath.Join(tmpDir, module.Name())
		err = downloadFile(moduleTempFile, moduleSource)
		if err != nil {
			showError(fmt.Sprintf("        > Source for "+module.Name()+" could not be reached: %s", err.Error()))
			keepOnHostUnreachable_config, _ := getConfigValue("KEEP_ON_HOST_UNREACHABLE")
			keepOnHostUnreachable, err := strconv.ParseBool(keepOnHostUnreachable_config)
			if err != nil {
				showAttention("            > Invalid option in preferences file for 'KEEP_ON_HOST_UNREACHABLE'.")
				showInfo("        > Skipping module...")
				continue
			}
			if keepOnHostUnreachable == false {
				return errors.New(fmt.Sprintf("        > Error: failed to get module %s and KEEP_ON_HOST_UNREACHABLE is set to 'false'",module.Name()))
			} else {
				showInfo("        > Skipping module...")
			}
		}

		insertLine(tmphosts_file,"")
		insertComment(tmphosts_file,fmt.Sprintf("Hosts from web module '%s'",module.Name()))

		moduleFile, err := os.Open(moduleTempFile)
		if err != nil {
			showAttention("        > Error opening module file "+moduleTempFile+": "+err.Error())
			continue
		}
		defer moduleFile.Close()
		scanner := bufio.NewScanner(moduleFile)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "#") && strings.TrimSpace(line) != "" {
				insertLine(tmphosts_file,line)
			}
		}
		insertLine(tmphosts_file,"")

		showSuccess("        > Done")
		i++
	}

	return nil
}


func verifyInternetConnection() {
	ip_test, err_test := getConfigValue("IP_TEST")

	showInfoSectionTitle(fmt.Sprintf("Internet connection verification (IP/Hostname: %s)",ip_test))

	if err_test != nil {
		showAttention("    > Error getting the IP_TEST preference in config file")
		finishProgram(1)
	}

	cmd := exec.Command("ping", "-c", "1", "-W", "5", ip_test)

	err := cmd.Run()
	if err != nil {
		showError(fmt.Sprintf("    > Error: %s",err.Error()))
		time.Sleep(2 * time.Second)
		finishProgram(1)
	}
	showSuccess("    > Passed")

	time.Sleep(2 * time.Second)
}

func verifyIntegrity() {
	showInfoSectionTitle("Program directories integrity verification")

	if _, err := os.Stat(localModulesDir); os.IsNotExist(err) {
		showError(fmt.Sprintf("    > Error: local hosts directory not found at %s.", localModulesDir))
		finishProgram(1)
	}
	if _, err := os.Stat(webModulesDir); os.IsNotExist(err) {
		showError(fmt.Sprintf("    > Error: Web hosts directory not found at %s.", webModulesDir))
		finishProgram(1)
	}
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		showError(fmt.Sprintf("    > Error: configuration directory not found at %s.", configDir))
		finishProgram(1)
	}
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		showInfo(fmt.Sprintf("    > Error: backup directory not found at %s. Creating one...", backupDir))
		os.Mkdir(backupDir, 0755)
	}
	showSuccess("    > Passed")
}

type backup_file struct {
	filename string
	path string
}

func backupHostfile(tempDir string) (backup_file, error) {
	showInfoSectionTitle("Backing up current /etc/hosts file")

	showInfo("    > Cleaning up backup directory if needed")

	// Clean up backup directory if above limit set by user
	backupDirMaxFilesStr, _ := getConfigValue("MAX_BACKUP_FILES")
	backupDirMaxFiles, err := strconv.Atoi(backupDirMaxFilesStr)
	if err != nil {
		return backup_file{filename:"",path:""}, errors.New(fmt.Sprintf("        > Error: failed to convert MAX_BACKUP_FILES to integer: " + err.Error()))
	}

	currentBackupFiles, err := ioutil.ReadDir(backupDir)
	if err != nil {
		return backup_file{filename:"",path:""}, errors.New(fmt.Sprintf("        > Error: failed to read backup directory: " + err.Error()))
	}

	if len(currentBackupFiles) > backupDirMaxFiles {
		numFilesToRemove := len(currentBackupFiles) - backupDirMaxFiles

		fileNames := make([]string, len(currentBackupFiles))
		for i, file := range currentBackupFiles {
			fileNames[i] = filepath.Join(backupDir, file.Name())
		}

		// Sort backup files by modification time (oldest first)
		sort.Slice(fileNames, func(i, j int) bool {
			fileInfoI, _ := os.Stat(fileNames[i])
			fileInfoJ, _ := os.Stat(fileNames[j])
			return fileInfoI.ModTime().Before(fileInfoJ.ModTime())
		})

		// Remove the oldest backup files
		for i := 0; i < numFilesToRemove; i++ {
			err := os.Remove(fileNames[i])
			if err != nil {
				return backup_file{filename:"",path:""}, errors.New((fmt.Sprintf("        > Error: failed to remove backup file %s: %s",fileNames[i],err)))
			} else {
				showInfo("        > Removed backup file " + fileNames[i])
			}
		}
	}

	showInfo("    > Backing up /etc/hosts file")

	// Generate backup filename and full path
	backupFile := fmt.Sprintf("%v.BACKUP", time.Now().Format("2006-01-02-15-04-05"))
	dst, err := os.Create(filepath.Join(backupDir, backupFile))
	if err != nil {
		return backup_file{filename:"",path:""}, errors.New(("        > Error: failed to create backup file: " + err.Error()))
	}

	// Backup current hosts file
	if _, err := os.Stat(hostsFile); err == nil {
		src, err := os.Open(hostsFile)
		if err != nil {
			return backup_file{filename:"",path:""}, errors.New(("        > Error: failed to open current /etc/hosts file: " + err.Error()))
		}
		_, err = io.Copy(dst, src)
		if err != nil {
			return backup_file{filename:"",path:""}, errors.New(("        > Error: failed to copy hosts file to backup file: " + err.Error()))
		}
		src.Close()
		dst.Close()

		showInfo("        > Current /etc/hosts file backed up as " + backupFile)
	} else {
			return backup_file{filename:"",path:""}, errors.New(("        > Error: /etc/hosts file not found!"))
	}

	showSuccess("    > Done")

	return backup_file{
		filename: backupFile,
		path:	  dst.Name(),
	},nil
}

func createTempDir() (string, error) {
	showInfoSectionTitle("Creating temporary directory")
	// Get the current time as a Unix timestamp
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Create a random string of 32 characters
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", errors.New(fmt.Sprintf("    > Error: %s", err.Error()))
	}
	randomString := base64.URLEncoding.EncodeToString(randomBytes)

	// Create the temporary directory path
	tempDir := "/tmp/" + timestamp + "-" + randomString

	// Create the temporary directory
	err = os.Mkdir(tempDir, 0755)
	if err != nil {
		return "", errors.New(fmt.Sprintf("    > Error: %s",err.Error()))
	}

	showSuccess(fmt.Sprintf("    > Created at %s",tempDir))

	return tempDir, nil
}

func removeTmpDir(tmpDir string) {
	showInfoSectionTitle("Removing temporary directory")
	err := os.RemoveAll(tmpDir)
	if err != nil {
		showError("    > Error: failed to remove temporary directory: " + err.Error())
		finishProgram(1)
	}
	showSuccess("    > Removed")
}

func createTempHostsFile(tmpDir string) (string, error) {
	showInfoSectionTitle("Creating temporary hosts file")
	tmpHostsFile := tmpDir + "/hosts"

	err := ioutil.WriteFile(tmpHostsFile, []byte(""), 0644)
	if err != nil {
		return "", errors.New(fmt.Sprintf("    > Error: failed to create file at %s", tmpHostsFile))
	}
	showSuccess("    > Created")
	return tmpHostsFile, nil
}

func writeHeader(tmphosts_file string) {
	showInfoSectionTitle("Writing header to temporary hosts file")
	insertComment(tmphosts_file, fmt.Sprintf(" This file was edited by update-hosts-file (v%s)", programVersion))
	insertComment(tmphosts_file, " Date: " + time.Now().String())
	insertComment(tmphosts_file," update-hosts-file is a program that automatically updates this file.")
	insertComment(tmphosts_file," It can be configured to pull host information from various sources,")
	insertComment(tmphosts_file," such as web-based and local blocklists files. It also automatically")
	insertComment(tmphosts_file," adds this machine's hostname to make sure any changes to it will be")
	insertComment(tmphosts_file," reflected here.")
	insertLine(tmphosts_file, "")
	showSuccess("    > Done")
}

func overwriteHostsFileWithTempFile(tempFilePath string) error {
	hostsFilePath := "/etc/hosts"

	showInfoSectionTitle("Overwriting current hosts file with the temporary one")
	emptyString := ""
	err := ioutil.WriteFile(hostsFilePath, []byte(emptyString), 0644)
	if err != nil {
		return errors.New(fmt.Sprintf("    > Error: failed to clean contents of hosts file at %s",hostsFilePath))
	}
	err = exec.Command("cp", tempFilePath, hostsFilePath).Run()
	if err != nil {
		return errors.New(fmt.Sprintf("    > Error: failed to copy temporary hosts file to %s",hostsFilePath))
	}

	showSuccess("    > Done")
	return nil
}

func showHostsFileUpdateMessage() error {
	hostsFilePath := "/etc/hosts"

	showInfoSectionTitle("Finished updating the /etc/hosts file")

	output, err := exec.Command("wc", hostsFilePath).Output()
	if err != nil {
		return err
	}

	linesWritten := strings.Fields(string(output))[0]
	showSuccess(fmt.Sprintf("    > %s lines were written.", linesWritten))
	fmt.Println()
	return nil
}

func openHostsFileWithViewer() error {
	hostsFilePath := "/etc/hosts"

	viewer, err := getConfigValue("DEFAULT_VIEWER")
	if err != nil {
		return errors.New("    > Failed to get DEFAULT_VIEWER config variable")
	}

	cmd := exec.Command(viewer, hostsFilePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return errors.New(fmt.Sprintf("    > Failed to open /etc/hosts file: %s",err.Error()))
	}

	return nil
}

func openHostsFileWithEditor() error {
	hostsFilePath := "/etc/hosts"

	editor, err := getConfigValue("DEFAULT_EDITOR")
	if err != nil {
		return errors.New("    > Failed to get DEFAULT_EDITOR config variable")
	}

	cmd := exec.Command(editor, hostsFilePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return errors.New(fmt.Sprintf("    > Failed to open /etc/hosts file: %s",err.Error()))
	}

	return nil
}

func finishProgramMenu() {
	options := []string{
		"Finish",
		"Edit /etc/hosts",
		"View /etc/hosts",
	}

	fmt.Println("")

	prompt := &survey.Select{
		Message: "What do you want to do?",
		Options: options,
	}

	var option string
	err := survey.AskOne(prompt, &option)
	if err != nil {
		showAttention("Error displaying menu: " + err.Error())
		finishProgram(1)
	}

	switch option {
	case options[0]:
		finishProgram(0)
	case options[1]:
		err := openHostsFileWithEditor()
		if err != nil {
			showError(fmt.Sprintf(err.Error()))
			finishProgram(1)
		}
		finishProgram(0)
	case options[2]:
		err := openHostsFileWithViewer()
		if err != nil {
			showError(fmt.Sprintf(err.Error()))
			finishProgram(1)
		}
		finishProgram(0)
	}
}

func viewModule(moduleName string, webModule bool, localModule bool) error {
	showInfo(fmt.Sprintf("Viewing module '%s'",moduleName))
	var moduleDir string
	if webModule {
		moduleDir = webModulesDir
	} else if localModule {
		moduleDir = localModulesDir
	}

	availableModulePath := filepath.Join(moduleDir,"available",moduleName)

	viewer, err := getConfigValue("DEFAULT_VIEWER")
	if err != nil {
		return errors.New("    > Failed to get DEFAULT_VIEWER config variable from preferences file")
	}

	_, err = os.Stat(availableModulePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("    > Not found")
	} else if err != nil {
		return fmt.Errorf("    > Error when trying to verify if module exists: %s", err.Error())
	}

	cmd := exec.Command(viewer, availableModulePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return errors.New(fmt.Sprintf("    > Failed to open /etc/hosts file: %s",err.Error()))
	}

	showSuccess("    > Done")

	return nil
}

func editModule(moduleName string, webModule bool, localModule bool) error {
	showInfo(fmt.Sprintf("Editing module '%s'",moduleName))
	var moduleDir string
	if webModule {
		moduleDir = webModulesDir
	} else if localModule {
		moduleDir = localModulesDir
	}

	availableModulePath := filepath.Join(moduleDir,"available",moduleName)

	editor, err := getConfigValue("DEFAULT_EDITOR")
	if err != nil {
		return errors.New("    > Failed to get DEFAULT_EDITOR config variable from preferences file")
	}

	_, err = os.Stat(availableModulePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("    > Not found")
	} else if err != nil {
		return fmt.Errorf("    > Error when trying to verify if module exists: %s", err.Error())
	}

	cmd := exec.Command(editor, availableModulePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return errors.New(fmt.Sprintf("    > Failed to open /etc/hosts file: %s",err.Error()))
	}

	showSuccess("    > Done")

	return nil
}

func rmModule(moduleName string, webModule bool, localModule bool) error {
	showInfo(fmt.Sprintf("Removing module '%s'",moduleName))
	var moduleDir string
	if webModule {
		moduleDir = webModulesDir
	} else if localModule {
		moduleDir = localModulesDir
	}

	availableModulePath := filepath.Join(moduleDir,"available",moduleName)
	enabledModulePath := filepath.Join(moduleDir,"enabled",moduleName)

	_, err := os.Stat(availableModulePath)
	if err != nil {
		return fmt.Errorf("    > Not found")
	}

	// Disable it (if needed)
	showInfo("    > Verifying if module was disabled")
	_, err = os.Stat(enabledModulePath)
	if os.IsNotExist(err) {
		showAttention("        > Already disabled")
	} else if err != nil {
		return fmt.Errorf("        > Error when trying to verify if module is already disabled: %s", err.Error())
	} else {
		err = os.Remove(enabledModulePath)
		if err != nil {
			return fmt.Errorf("        > Error when trying to remove symlink to module file: %s", err.Error())
		}
		showSuccess("        > Done")
	}

	// And then, remove it
	showInfo("    > Removing module")
	err = os.Remove(availableModulePath)
	if err != nil {
		return fmt.Errorf("        > Error when trying to remove module file: %s", err.Error())
	}
	showSuccess("        > Done")

	return nil
}

func addModule(moduleName string, webModule bool, localModule bool) error {
	showInfo(fmt.Sprintf("Adding module '%s'",moduleName))
	var moduleDir string
	if webModule {
		moduleDir = webModulesDir
	} else if localModule {
		moduleDir = localModulesDir
	}

	availableModulePath := filepath.Join(moduleDir,"available",moduleName)

	editor, err := getConfigValue("DEFAULT_EDITOR")
	if err != nil {
		return errors.New("    > Failed to get DEFAULT_EDITOR config variable")
	}

	_, err = os.Stat(availableModulePath)
	if err == nil {
		return fmt.Errorf("    > Alread exists")
	}

	cmd := exec.Command(editor, availableModulePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	showInfo("        > Opening module file using the default editor")
	err = cmd.Run()
	if err != nil {
		return errors.New(fmt.Sprintf("        > Failed to open /etc/hosts file: %s",err.Error()))
	}

	showSuccess("        > Done")

	return nil
}

func enableModule(moduleName string, webModule bool, localModule bool) error {
	showInfo(fmt.Sprintf("Enabling module '%s'",moduleName))
	var moduleDir string
	if webModule {
		moduleDir = webModulesDir
	} else if localModule {
		moduleDir = localModulesDir
	}

	availableModulePath := filepath.Join(moduleDir,"available",moduleName)
	enabledModulePath := filepath.Join(moduleDir,"enabled",moduleName)

	_, err := os.Stat(availableModulePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("    > Not found")
	} else if err != nil {
		return fmt.Errorf("    > Error when trying to verify if module exists: %s", err.Error())
	}

	_, err = os.Stat(enabledModulePath)
	if os.IsNotExist(err) {
		err = os.Symlink(availableModulePath, enabledModulePath)
		if err != nil {
			return fmt.Errorf("    > Error when trying to symlink module file: %s", err.Error())
		}
		showSuccess("    > Done")
	} else if err == nil {
			showAttention("    > Already enabled")
	} else {
		return fmt.Errorf("    > Error when trying to verify if module is already enabled: %s", err.Error())
	}
	return nil
}

func disableModule(moduleName string, webModule bool, localModule bool) error {
	showInfo(fmt.Sprintf("Disabling module '%s'",moduleName))
	var moduleDir string
	if webModule {
		moduleDir = webModulesDir
	} else if localModule {
		moduleDir = localModulesDir
	}

	availableModulePath := filepath.Join(moduleDir,"available",moduleName)
	enabledModulePath := filepath.Join(moduleDir,"enabled",moduleName)

	_, err := os.Stat(availableModulePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("    > Not found")
	} else if err != nil {
		return fmt.Errorf("    > Error when trying to verify if module exists: %s", err.Error())
	}

	_, err = os.Stat(enabledModulePath)
	if os.IsNotExist(err) {
		showAttention("    > Already disabled")
	} else if err != nil {
		return fmt.Errorf("    > Error when trying to verify if module is already disabled: %s", err.Error())
	} else {
		err = os.Remove(enabledModulePath)
		if err != nil {
			return fmt.Errorf("    > Error when trying to remove symlink to module file: %s", err.Error())
		}
		showSuccess("    > Disabled")
	}

	return nil
}

func listLocalModules() error {
	showInfoSectionTitle("Listing local modules")
	availableLocalModules, err := ioutil.ReadDir(localModulesDir + "/available")
	if err != nil {
		return errors.New(fmt.Sprintf("    > Error: failed to read local modules directory: " + err.Error()))
	}

	for _, module := range availableLocalModules {
		_, err = os.Stat(localModulesDir + "/enabled/" + module.Name())
		if os.IsNotExist(err) {
			redHex := "#ff5050"
			red := color.HEX(redHex)

			fmt.Println(fmt.Sprintf("%s %s",module.Name(),red.Sprintf("(disabled)")))
		} else if err != nil {
			return fmt.Errorf("    > Error when trying to verify if module %s is enabled: %s", module.Name(), err.Error())
		} else {
			blueHex := "#55aaff"
			blue := color.HEX(blueHex)

			fmt.Println(fmt.Sprintf("%s %s",module.Name(),blue.Sprintf("(enabled)")))
		}
	}
	return nil
}

func listWebModules() error {
	showInfoSectionTitle("Listing web modules")
	availableWebModules, err := ioutil.ReadDir(webModulesDir + "/available")
	if err != nil {
		return errors.New(fmt.Sprintf("    > Error: failed to read web modules directory: " + err.Error()))
	}

	for _, module := range availableWebModules {

		_, err = os.Stat(webModulesDir + "/enabled/" + module.Name())
		if os.IsNotExist(err) {
			redHex := "#ff5050"
			red := color.HEX(redHex)

			fmt.Println(fmt.Sprintf("%s %s",module.Name(),red.Sprintf("(disabled)")))
		} else if err != nil {
			return fmt.Errorf("    > Error when trying to verify if module %s is enabled: %s", module.Name(), err.Error())
		} else {
			blueHex := "#55aaff"
			blue := color.HEX(blueHex)

			fmt.Println(fmt.Sprintf("%s %s",module.Name(),blue.Sprintf("(enabled)")))
		}
	}
	return nil
}

func listModules(webModules bool, localModules bool, allModules bool) error {
	if localModules {
		err := listLocalModules()
		if err != nil {
			return errors.New(fmt.Sprintf(err.Error()))
		}
	} else if webModules {
		err := listWebModules()
		if err != nil {
			return errors.New(fmt.Sprintf(err.Error()))
		}
	} else if allModules {
		// List local modules
		err := listLocalModules()
		if err != nil {
			return errors.New(fmt.Sprintf(err.Error()))
		}

		fmt.Println("")

		// List web modules
		err = listWebModules()
		if err != nil {
			return errors.New(fmt.Sprintf(err.Error()))
		}
	}
	return nil
}

//
////
//

func main() {
	var rootCmd = &cobra.Command{
		Use:   "update-hosts-file [command]",
		Short: programShortDescription,
		Long:  programLongDescription,
		Run: func(cmd *cobra.Command, args []string) {
			blueHex := "#55aaff"
			blue := color.HEX(blueHex)

			displayProgramInfo()

			space()

			showText(fmt.Sprintf("Run %v to get started. \n\nTo know more about the program, run %v.", blue.Sprintf("update-hosts-file --help/-h"), blue.Sprintf("update-hosts-file --about")))
			finishProgram(0)
		},
	}

	var showAboutCmd = &cobra.Command{
		Use:	"about",
		Short:	"Shows program's information",
		Run: func(cmd *cobra.Command, args []string) {
			displayProgramInfo()

			space()
			hr("=", 0.8)
			space()

			showText(programLongDescription)
		},
	}

	var enableServiceCmd = &cobra.Command{
		Use:   "enable",
		Short: "Enables the UpdateHostsFile systemd service",
		Run: func(cmd *cobra.Command, args []string) {
			if err := exec.Command("systemctl", "enable", "updatehostsfile.service").Run(); err != nil {
				log.Fatalf("Error enabling systemd service: %v", err)
			}
			fmt.Println("UpdateHostsFile systemd service has been enabled")
		},
	}

	var disableServiceCmd = &cobra.Command{
		Use:   "disable",
		Short: "Disables the UpdateHostsFile systemd service",
		Run: func(cmd *cobra.Command, args []string) {
			if err := exec.Command("systemctl", "disable", "updatehostsfile.service").Run(); err != nil {
				log.Fatalf("Error disabling systemd service: %v", err)
			}
			fmt.Println("UpdateHostsFile systemd service has been disabling")
		},
	}

	var showVersionCmd = &cobra.Command{
		Use:	"version",
		Short:	"Shows program's version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("UpdateHostsFile")
			showInfoSectionTitle(fmt.Sprintf("Version: %s", programVersion))
		},
	}

	var webModule bool
	var localModule bool
	var allModule bool
	var moduleName string

	var modulesCmd = &cobra.Command{
		Use:   "modules",
		Short: "Manages the modules used to update the /etc/hosts file",
	}

	var enableModuleCmd = &cobra.Command{
		Use:   "enable",
		Short: "Enables a module",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if !webModule && !localModule {
				return errors.New("You need to insert at least an option: --web or --local")
				} else if webModule && localModule {
					return errors.New("Options --web and --local are conflicting")
				}
				if moduleName == "" {
					return errors.New("Module name not provided")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := enableModule(moduleName, webModule, localModule)
			if err != nil {
				showError(fmt.Sprintf(err.Error()))
			}
		},
	}

	enableModuleCmd.Flags().BoolVarP(&webModule, "web", "w", false, "Enable web module")
	enableModuleCmd.Flags().BoolVarP(&localModule, "local", "l", false, "Enable local module")
	enableModuleCmd.Flags().StringVarP(&moduleName, "module", "m", "", "Module name")
	enableModuleCmd.MarkFlagRequired("module")
	enableModuleCmd.Flags().SetInterspersed(false)

	var disableModuleCmd = &cobra.Command{
		Use:   "disable [module]",
		Short: "Disables a module",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if !webModule && !localModule {
				return errors.New("You need to insert at least an option: --web or --local")
				} else if webModule && localModule {
					return errors.New("Options --web and --local are conflicting")
				}
				if moduleName == "" {
					return errors.New("Module name not provided")
				}
				return nil
			},
		Run: func(cmd *cobra.Command, args []string) {
			err := disableModule(moduleName, webModule, localModule)
			if err != nil {
				showError(fmt.Sprintf(err.Error()))
			}
		},
	}

	disableModuleCmd.Flags().BoolVarP(&webModule, "web", "w", false, "Disable web module")
	disableModuleCmd.Flags().BoolVarP(&localModule, "local", "l", false, "Disable local module")
	disableModuleCmd.Flags().StringVarP(&moduleName, "module", "m", "", "Module name")
	disableModuleCmd.MarkFlagRequired("module")
	disableModuleCmd.Flags().SetInterspersed(false)

	var addModuleCmd = &cobra.Command{
		Use:   "add [module-file]",
		Short: "Adds a module from a file",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if !webModule && !localModule {
				return errors.New("You need to insert at least an option: --web or --local")
			} else if webModule && localModule {
					return errors.New("Options --web and --local are conflicting")
			}
			if moduleName == "" {
				return errors.New("Module name not provided")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := addModule(moduleName, webModule, localModule)
			if err != nil {
				showError(fmt.Sprintf(err.Error()))
			}
		},
	}

	addModuleCmd.Flags().BoolVarP(&webModule, "web", "w", false, "Add web module")
	addModuleCmd.Flags().BoolVarP(&localModule, "local", "l", false, "Add local module")
	addModuleCmd.Flags().StringVarP(&moduleName, "module", "m", "", "Module name")
	addModuleCmd.MarkFlagRequired("module")
	addModuleCmd.Flags().SetInterspersed(false)

	var rmModuleCmd = &cobra.Command{
		Use:   "rm [module]",
		Short: "Removes a module",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if !webModule && !localModule {
				return errors.New("You need to insert at least an option: --web or --local")
			} else if webModule && localModule {
				return errors.New("Options --web and --local are conflicting")
			}
			if moduleName == "" {
				return errors.New("Module name not provided")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := rmModule(moduleName, webModule, localModule)
			if err != nil {
				showError(fmt.Sprintf(err.Error()))
			}
		},
	}

	rmModuleCmd.Flags().BoolVarP(&webModule, "web", "w", false, "Remove web module")
	rmModuleCmd.Flags().BoolVarP(&localModule, "local", "l", false, "Remove local module")
	rmModuleCmd.Flags().StringVarP(&moduleName, "module", "m", "", "Module name")
	rmModuleCmd.MarkFlagRequired("module")
	rmModuleCmd.Flags().SetInterspersed(false)

	var editModuleCmd = &cobra.Command{
		Use:   "edit [module]",
		Short: "Edits a module",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if !webModule && !localModule {
				return errors.New("You need to insert at least an option: --web or --local")
			} else if webModule && localModule {
				return errors.New("Options --web and --local are conflicting")
			}
			if moduleName == "" {
				return errors.New("Module name not provided")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := editModule(moduleName, webModule, localModule)
			if err != nil {
				showError(fmt.Sprintf(err.Error()))
			}
		},
	}

	editModuleCmd.Flags().BoolVarP(&webModule, "web", "w", false, "Edit web module")
	editModuleCmd.Flags().BoolVarP(&localModule, "local", "l", false, "Edit local module")
	editModuleCmd.Flags().StringVarP(&moduleName, "module", "m", "", "Module name")
	editModuleCmd.MarkFlagRequired("module")
	editModuleCmd.Flags().SetInterspersed(false)

	var viewModuleCmd = &cobra.Command{
		Use:   "view [module]",
		Short: "Views a module",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if !webModule && !localModule {
				return errors.New("You need to insert at least an option: --web or --local")
			} else if webModule && localModule {
				return errors.New("Options --web and --local are conflicting")
			}
			if moduleName == "" {
				return errors.New("Module name not provided")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := viewModule(moduleName, webModule, localModule)
			if err != nil {
				showError(fmt.Sprintf(err.Error()))
			}
		},
	}

	viewModuleCmd.Flags().BoolVarP(&webModule, "web", "w", false, "View web module")
	viewModuleCmd.Flags().BoolVarP(&localModule, "local", "l", false, "View local module")
	viewModuleCmd.Flags().StringVarP(&moduleName, "module", "m", "", "Module name")
	viewModuleCmd.MarkFlagRequired("module")
	viewModuleCmd.Flags().SetInterspersed(false)

	var listModulesCmd = &cobra.Command{
		Use:   "list [module]",
		Short: "Lists modules",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if !webModule && !localModule && !allModule {
				return errors.New("You need to insert at least an option: --web, --local or --all")
			} else if webModule && localModule && allModule {
				return errors.New("Options --web and --local and --all are conflicting")
			} else if webModule && localModule {
				return errors.New("Options --web and --local are conflicting")
			} else if webModule && allModule {
				return errors.New("Options --web and --all are conflicting")
			} else if localModule && allModule {
				return errors.New("Options --local and --all are conflicting")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := listModules(webModule, localModule, allModule)
			if err != nil {
				showError(fmt.Sprintf(err.Error()))
			}
		},
	}

	listModulesCmd.Flags().BoolVarP(&webModule, "web", "w", false, "List web modules")
	listModulesCmd.Flags().BoolVarP(&localModule, "local", "l", false, "List local modules")
	listModulesCmd.Flags().BoolVarP(&allModule, "all", "a", false, "List all modules")
	listModulesCmd.Flags().SetInterspersed(false)

	var noInteractive bool
	var updateHostsFileCmd = &cobra.Command{
		Use:   "update",
		Short: "Updates the /etc/hosts file according to enabled modules" ,
		Run: func(cmd *cobra.Command, args []string) {
			verifyInternetConnection()

			fmt.Println("")

			temp_dir, err := createTempDir()
			if err != nil {
				showError(fmt.Sprintf(err.Error()))
				finishProgram(1)
			}

			fmt.Println("")

			verifyIntegrity()

			fmt.Println("")

			backup_file, err := backupHostfile(temp_dir)
			if err != nil {
				showError(fmt.Sprintf(err.Error()))
				removeTmpDir(temp_dir)
				finishProgram(1)
			}

			fmt.Println("")

			tmphosts_file, err := createTempHostsFile(temp_dir)
			if err != nil {
				showError(fmt.Sprintf(err.Error()))
				removeTmpDir(temp_dir)
				finishProgram(1)
			}

			fmt.Println("")

			writeHeader(tmphosts_file)

			fmt.Println("")

			insertHostname(tmphosts_file)

			fmt.Println("")

			err = loadLocalModules(tmphosts_file)
			if err != nil {
				showError(fmt.Sprintf(err.Error()))
				restoreBackup(backup_file)
				removeTmpDir(temp_dir)
				finishProgram(1)
			}

			fmt.Println("")

			err = loadWebModules(tmphosts_file, temp_dir)
			if err != nil {
				showError(fmt.Sprintf(err.Error()))
				restoreBackup(backup_file)
				removeTmpDir(temp_dir)
				finishProgram(1)
			}

			fmt.Println("")

			err = overwriteHostsFileWithTempFile(tmphosts_file)
			if err != nil {
				showError(fmt.Sprintf(err.Error()))
				restoreBackup(backup_file)
				removeTmpDir(temp_dir)
				finishProgram(1)
			}

			fmt.Println("")

			showHostsFileUpdateMessage()

			removeTmpDir(temp_dir)

			if !noInteractive {
				finishProgramMenu()
			} else {
				fmt.Println("")
				showInfo("Program finished")
			}

			finishProgram(0)
		},
	}
	updateHostsFileCmd.Flags().BoolVar(&noInteractive, "no-interactive", false, "Skip the interactive finish program menu")

	// Add Cobra commands
	rootCmd.AddCommand(enableServiceCmd)
	rootCmd.AddCommand(disableServiceCmd)
	rootCmd.AddCommand(updateHostsFileCmd)
	rootCmd.AddCommand(showVersionCmd)
	rootCmd.AddCommand(showAboutCmd)
	modulesCmd.AddCommand(enableModuleCmd)
	modulesCmd.AddCommand(disableModuleCmd)
	modulesCmd.AddCommand(addModuleCmd)
	modulesCmd.AddCommand(rmModuleCmd)
	modulesCmd.AddCommand(editModuleCmd)
	modulesCmd.AddCommand(viewModuleCmd)
	modulesCmd.AddCommand(listModulesCmd)
	rootCmd.AddCommand(modulesCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
