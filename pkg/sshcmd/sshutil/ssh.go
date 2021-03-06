package sshutil

import (
	"bufio"
	"io"
	"strings"

	"preflight/pkg/logger"
)

//Cmd is in host exec cmd
func (ss *SSH) Cmd(host string, cmd string) []byte {
	//TODO, hack, should enable when use log file
	//logger.Debug("[ssh][%s] %s", host, cmd)
	session, err := ss.Connect(host)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("[ssh][%s]Error create ssh session failed,%s", host, err)
		}
	}()
	if err != nil {
		panic(1)
	}
	defer session.Close()
	b, err := session.CombinedOutput(cmd)
	//TODO, hack, should enable when use log file
	logger.Debug("[ssh][%s]command result is: %s", host, string(b))
	logger.Debug("[ssh][%s]command '[%s]' is %s", host, cmd, err)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("[ssh][%s]Error exec command failed: %s", host, err)
		}
	}()
	if err != nil {
		panic(1)
	}
	return b
}

//CmdNew is alternative for Cmd
func (ss *SSH) CmdNew(host string, cmd string) (info string, err error) {
	session, err := ss.Connect(host)
	defer session.Close()
	if err != nil {
		return "", err
	}

	b, err := session.CombinedOutput(cmd)
	return string(b), err
}

func readPipe(host string, pipe io.Reader, isErr bool) {
	r := bufio.NewReader(pipe)
	for {
		line, _, err := r.ReadLine()
		if line == nil {
			return
		} else if err != nil {
			logger.Debug("[%s] %s", host, line)
			logger.Error("[ssh] [%s] %s", host, err)
			return
		} else {
			if isErr {
				logger.Error("[%s] %s", host, line)
			} else {
				logger.Debug("[%s] %s", host, line)
			}
		}
	}
}

func (ss *SSH) CmdAsync(host string, cmd string) error {
	logger.Debug("[ssh][%s] %s", host, cmd)
	session, err := ss.Connect(host)
	if err != nil {
		logger.Error("[ssh][%s]Error create ssh session failed,%s", host, err)
		return err
	}
	defer session.Close()
	stdout, err := session.StdoutPipe()
	if err != nil {
		logger.Error("[ssh][%s]Unable to request StdoutPipe(): %s", host, err)
		return err
	}
	// [TODO: huizhi]don't know why we can not find error message
	stderr, err := session.StderrPipe()
	if err != nil {
		logger.Error("[ssh][%s]Unable to request StderrPipe(): %s", host, err)
		return err
	}
	// we cannot set env: https://vic.demuzere.be/articles/environment-variables-setenv-ssh-golang/
	// for k, v := range ss.Env {
	// 	if err := session.Setenv(k, v); err != nil {
	// 		logger.Error(err.Error())
	// 	}
	// }
	if err := session.Start(cmd); err != nil {
		logger.Error("[ssh][%s]Unable to execute command: %s", host, err)
		return err
	}

	doneout := make(chan bool, 1)
	doneerr := make(chan bool, 1)
	go func() {
		readPipe(host, stderr, true)
		doneerr <- true
	}()
	go func() {
		readPipe(host, stdout, false)
		doneout <- true
	}()
	<-doneerr
	<-doneout
	return session.Wait()
}

//CmdToString is in host exec cmd and replace to spilt str
func (ss *SSH) CmdToString(host, cmd, spilt string) string {
	data := ss.Cmd(host, cmd)
	if data != nil {
		str := string(data)
		str = strings.ReplaceAll(str, "\r\n", spilt)
		return str
	}
	return ""
}
