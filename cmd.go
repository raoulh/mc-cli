package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/howeyc/gopass"
)

const (
	maxFileSize = 524288
)

func processLoginCmd(subCmd, context, login, pwd, desc string, printDesc bool) (err error) {

	m := &MoolticuteMsg{
		Data: MsgData{
			Service: context,
			Login:   login,
		},
	}

	if subCmd == "get" {
		m.Msg = "get_credential"
	} else if subCmd == "set" {
		if pwd == "" {
			fmt.Printf("Password: ")
			p, err := gopass.GetPasswdMasked()
			if err != nil {
				return err
			}
			pwd = string(p)
		}

		m.Msg = "set_credential"
		m.Data.Password = pwd
		m.Data.Description = desc
		m.Data.McCliVersion = "1.0"
	}

	res, err := McSendQuery(m)
	if err != nil {
		return err
	}

	if subCmd == "get" {
		if printDesc {
			fmt.Println(res.Description)
		} else {
			fmt.Println(res.Password)
		}
	} else if subCmd == "set" {
		fmt.Println(green(CharCheck), "Done")
	}

	return
}

func processDataCmd(subCmd, context, filename string, progressFunc ProgressCb) (err error) {

	m := &MoolticuteMsg{
		Data: MsgData{
			Service: context,
		},
	}

	if subCmd == "get" {
		m.Msg = "get_data_node"
	} else if subCmd == "set" {
		m.Msg = "set_data_node"

		//open file and encode to base64
		finfo, err := os.Stat(filename)
		if err != nil {
			return fmt.Errorf("failed to get file info: %v", err)
		}
		if finfo.Size() > maxFileSize {
			return fmt.Errorf("file is too big for beeing saved into the mooltipass :(")
		}

		b, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}

		m.Data.NodeData = base64.StdEncoding.EncodeToString(b)
	}

	res, err := McSendQueryProgress(m, progressFunc)
	if err != nil {
		return err
	}

	if subCmd == "get" {
		//decode the base64
		bdec, err := base64.StdEncoding.DecodeString(res.NodeData)
		if err != nil {
			err = fmt.Errorf("failed to base64 decode data: %v", err)
			return err
		}

		b := bytes.NewBuffer(bdec)
		b.WriteTo(os.Stdout)

	} else if subCmd == "set" {
		fmt.Println(green(CharCheck), "Done")
	}

	return
}

func processParameterCmd(subCmd, parameter string, value string) (err error) {
	fmt.Println(errorRed(CharAbort), "Not implemented yet")

	return (nil)
}
