/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package node

import (
	"github.com/fsnotify/fsnotify"
	"github.com/hopeio/utils/log"

	"os"
	"os/exec"
	"sync"
	"syscall"
)

var (
	currentCmd *exec.Cmd
	cmdMutex   sync.Mutex
)

func runNodeScript(script string) {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	// 如果有正在运行的命令，终止它
	if currentCmd != nil {
		if err := currentCmd.Process.Signal(syscall.SIGINT); err != nil {
			log.Printf("Error stopping previous process: %v", err)
		}
		currentCmd.Wait()
	}

	// 启动新命令
	cmd := exec.Command("node", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatalf("Error starting new script: %v", err)
	}

	currentCmd = cmd
}

func WatchRun(scriptFile string) {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Info("Modified file:", event.Name)
					runNodeScript(scriptFile)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error("Error:", err)
			}
		}
	}()

	err = watcher.Add(scriptFile)
	if err != nil {
		log.Fatal(err)
	}

	// Initial run
	runNodeScript(scriptFile)

	<-done
}
