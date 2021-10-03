package phpParser

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func getFile() (string, string, error) {
	pathPhp := os.Getenv("DIRPATH_PHP")

	if pathPhp == "" {
		return "", "", errors.New("DIRPATH_PHP is not set")
	}

	filePath := pathPhp + "/custom.conf"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", "", err
	}

	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", "", err
	}

	fileString := string(file)

	return filePath, fileString, nil
}

func tryUpdateValue(lower, upper int, command, option string) error {

	value, err := strconv.Atoi(option)
	if err != nil {
		return err
	}

	if value < lower || value > upper {
		return errors.New(fmt.Sprintf("Option '%v' is not supported.", option))
	}

	filePath, fileString, err := getFile()
	if err != nil {
		return err
	}

	commandStr := "php_admin_value[" + command + "] = "
	index := strings.Index(fileString, commandStr)

	if index == -1 {
		fileString = strings.Join([]string{fileString, commandStr + strconv.Itoa(value) + "M"}, "\n")
	} else {
		index2 := strings.Index(fileString[index+len(commandStr):], "M")
		if index2 == -1 {
			return errors.New(fmt.Sprintf("Not able to parse %v correctly, check for bugs", filePath))
		}
		fileString = strings.Join([]string{
			fileString[:index+len(commandStr)],
			strconv.Itoa(value),
			fileString[index+len(commandStr)+index2:],
		}, "")
	}

	// 0666 == read & write permission
	if err := ioutil.WriteFile(filePath, []byte(fileString), 0666); err != nil {
		return err
	}

	return nil
}

func UploadMaxFilesize(option string) error {
	lower, upper := 0, 64
	command := "upload_max_filesize"
	return tryUpdateValue(lower, upper, command, option)
}

func PostMaxSize(option string) error {
	lower, upper := 0, 64
	command := "post_max_size"
	return tryUpdateValue(lower, upper, command, option)
}

func MemoryLimit(option string) error {
	lower, upper := 0, 512
	command := "memory_limit"
	return tryUpdateValue(lower, upper, command, option)
}

func AllowUrlFopen(option string) error {

	if option != "on" && option != "off" {
		return errors.New(fmt.Sprintf("Option '%v' is not supported.", option))
	}

	filePath, fileString, err := getFile()
	if err != nil {
		return err
	}

	command := "php_admin_flag[allow_url_fopen] = "
	index := strings.Index(fileString, command)

	if index == -1 {
		fileString = strings.Join([]string{fileString, command + option}, "\n")
	} else {
		index2 := strings.Index(fileString[index+len(command):], "\n")
		if index2 == -1 {
			return errors.New(fmt.Sprintf("Not able to parse %v correctly, check for bugs", filePath))
		}
		fileString = strings.Join([]string{
			fileString[:index+len(command)],
			option,
			fileString[index+len(command)+index2:],
		}, "")
	}

	if err := ioutil.WriteFile(filePath, []byte(fileString), 0666); err != nil {
		return err
	}

	return nil
}
