package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	absP := `tmp`
	var xapkP string
	fmt.Println("请输入xapks的存放路径")
	fmt.Scanln(&xapkP)
	xapkfs := os.DirFS(xapkP)
	dirXAPK, err := fs.ReadDir(xapkfs, ".")
	if err != nil {
		panic(err)
	}
	counter := 0
	for _, file := range dirXAPK {
		err := RemoveContents("tmp")
		if err != nil {
			panic(err)
		}
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) != ".xapk" {
			continue
		}
		err = DeCompress(filepath.Join(xapkP, file.Name()), "tmp")
		if err != nil {
			panic(err)
		}
		dir := os.DirFS(absP)
		Install(dir, absP)
		_ = os.Remove(filepath.Join(xapkP, file.Name()))
		counter++
	}
	if counter != 0 {
		fmt.Printf("好耶，一共安装了%v个xapk文件。呼~全部安装完啦！", counter)
	} else {
		fmt.Println("555，一个xapk文件都没发现呢")
	}
}

func Install(f fs.FS, absP string) {
	manifest, err := fs.ReadFile(f, "manifest.json")
	if err != nil {
		panic(err)
	}
	apksJ := new(apks)
	err = json.Unmarshal([]byte(manifest), apksJ)
	if err != nil {
		panic(err)
	}
	var apksS = make([]string, 1)
	apksS = append(apksS, "install-multiple")
	if err != nil {
		panic(err)
	}
	for i := range apksJ.SplitApks {
		f, _ := f.Open(apksJ.SplitApks[i].File)
		//nolint
		defer f.Close()
		fi, _ := f.Stat()
		apksS = append(apksS, filepath.Join(absP, fi.Name()))
	}
	adbC := &exec.Cmd{
		Path:   "adb",
		Args:   apksS,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	fmt.Println(apksS)
	fmt.Println("Press enter to install...")
	fmt.Scanln()
	err = adbC.Run()
	if err != nil {
		panic(err)
	}
}

func DeCompress(zipFile, dest string) error {
	r, err := zip.OpenReader(zipFile)
	//nolint
	defer r.Close()
	if err != nil {
		return err
	}
	if dest != "" {
		if err := os.MkdirAll(dest, 0755); err != nil {
			return err
		}
	}
	for _, file := range r.File {
		p := filepath.Join(dest, file.Name)
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(p, file.Mode()); err != nil {
				return err
			}
			continue
		}
		fr, err := file.Open()
		defer fr.Close()
		if err != nil {
			return err
		}
		fw, err := os.OpenFile(p, os.O_CREATE|os.O_RDWR|os.O_TRUNC, file.Mode())
		defer fw.Close()
		if err != nil {
			return err
		}
		_, err = io.Copy(fw, fr)
		if err != nil {
			return err
		}
	}
	return nil
}

type apks struct {
	SplitApks []struct {
		File string `json:"file"`
	} `json:"split_apks"`
}

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(0)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}
