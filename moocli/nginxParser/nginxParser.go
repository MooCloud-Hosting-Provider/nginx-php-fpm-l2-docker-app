package nginxParser

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var fastcgi_cache_profile_disabled string = ``

var fastcgi_cache_profile_default string = `
fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
fastcgi_buffer_size 128k;
fastcgi_buffers 256 4k;
fastcgi_busy_buffers_size 256k;
fastcgi_temp_file_write_size 256k;
fastcgi_cache FASTCGI_CACHE;
fastcgi_cache_valid 200 60m;
fastcgi_cache_valid 404 5s;
`

func FastcgiCacheProfile(option string) error {

	path := os.Getenv("DIRPATH_NGINX")

	if path == "" {
		return errors.New("DIRPATH_NGINX is not set")
	}

	if err := os.Remove(path + "/fastcgi.conf"); err != nil {
		return err
	}

	file, err := os.Create(path + "/fastcgi.conf")
	if err != nil {
		return err
	}
	defer file.Close()

	switch option {
	case "default":
		file.WriteString(fastcgi_cache_profile_default)
	case "disabled":
		file.WriteString(fastcgi_cache_profile_disabled)
	default:
		return errors.New(fmt.Sprintf("Option '%v' is not supported.", option))
	}

	return nil
}

func FastcgiCacheLocation(option string) error {

	pathDrive := os.Getenv("DIRPATH_FASTCGI_CACHE_DRIVE")
	if pathDrive == "" {
		return errors.New("DIRPATH_FASTCGI_CACHE_DRIVE is not set")
	}
	pathRamdisk := os.Getenv("DIRPATH_FASTCGI_CACHE_RAMDISK")
	if pathRamdisk == "" {
		return errors.New("DIRPATH_FASTCGI_CACHE_RAMDISK is not set")
	}

	fastcgiDrive := "fastcgi_cache_path " + pathDrive + " levels=1:2 keys_zone=FASTCGI_CACHE:128m inactive=60m;"
	fastcgiRamdisk := "fastcgi_cache_path " + pathRamdisk + " levels=1:2 keys_zone=FASTCGI_CACHE:128m inactive=60m;"

	path := os.Getenv("DIRPATH_NGINX")
	if path == "" {
		return errors.New("DIRPATH_NGINX is not set")
	}

	filePath := path + "/default.conf"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return err
	}

	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	fileString := string(file)

	fileString = strings.ReplaceAll(fileString, fastcgiRamdisk, "")
	fileString = strings.ReplaceAll(fileString, fastcgiDrive, "")

	switch option {
	case "drive":
		if _, err := os.Stat(pathDrive); os.IsNotExist(err) {
			return err
		}
		fileString = strings.Join([]string{fastcgiDrive, fileString}, "")
	case "ramdisk":
		if _, err := os.Stat(pathRamdisk); os.IsNotExist(err) {
			return err
		}
		fileString = strings.Join([]string{fastcgiRamdisk, fileString}, "")
	default:
		return errors.New(fmt.Sprintf("Option '%v' is not supported.", option))
	}

	if err := ioutil.WriteFile(filePath, []byte(fileString), 0666); err != nil {
		return err
	}

	return nil
}
