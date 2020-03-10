// This file is part of MinIO Kubernetes Cloud
// Copyright (c) 2019 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/fatih/color"
	"github.com/minio/cli"
)

// Development command, will port-forward the public and private interfaces of m3 and the
var devCmd = cli.Command{
	Name:        "dev",
	Usage:       "dev command",
	Action:      dev,
	Subcommands: []cli.Command{},
}

func dev(ctx *cli.Context) error {
	fmt.Println("Starting development environment")

	m3PFCtx, m3Cancel := context.WithCancel(context.Background())
	nginxCtx, nCancel := context.WithCancel(context.Background())
	doneCh := make(chan struct{})

	// listen for kill sign to stop all the processes
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			close(doneCh)
			m3Cancel()
			nCancel()
			fmt.Println("Heard someone want us death x_x", sig)
		}
	}()
	config := &rest.Config{
		// TODO: switch to using cluster DNS.
		Host:            "http://localhost:8001",
		TLSClientConfig: rest.TLSClientConfig{},
		BearerToken:     "eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRhc2hib2FyZC10b2tlbi1mZ2J4NSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50Lm5hbWUiOiJkYXNoYm9hcmQiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiIyNGE3Mjg1OC00YjE4LTRhZDEtYjM4YS03ZTA2NGM2ODI1ZmEiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6ZGVmYXVsdDpkYXNoYm9hcmQifQ.OTj-gB3OnDA5yDmtRZVF9wxMx-6fT1o3vSmd_lZrCpddTBgSkUb2vnaB8eVDQ_DKN2fHsnWw6JvZoPftJ27gKVZ_dAM_21XwgUJy72_lhI_XLinGcx5TAqObxhLp5-YlCTQPDbVEW56DUs59mvx2KKaYeeS7KE-ORYN4wpH6ecZnhUR7_jhSdJAb9MBp3reUU6Iou2YDfEHtHgrSoF7EpZrQME8zjtTQE0Fkl6YavKA1zjHMg-yKuiFRjLkKcrcXyYa_j4lFXL_ZGEICy94FsjGAPv4iwCqZW9ruTU9EX0B0BbG4xGYEZfgG6B5iqIUdleYzHl86eSpWQMS5H5xguQ",
		BearerTokenFile: "some/file",
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("ERROR WITH INFORMER")
	}

	publicCh := servicePortForwardPort(m3PFCtx, "m3", "50051", color.FgYellow)
	privateCh := servicePortForwardPort(m3PFCtx, "m3", "50052", color.FgGreen)
	nginxCh := servicePortForwardPort(nginxCtx, "nginx-resolver", "9000:80", color.FgCyan)
	portalCh := servicePortForwardPort(nginxCtx, "portal-proxy", "9080:80", color.FgBlue)
	portalBackendCh := servicePortForwardPort(m3PFCtx, "m3-portal-backend", "5050", color.FgMagenta)
	initialized := false
	nginxInitialized := false

	// informer factory
	factory := informers.NewSharedInformerFactory(clientset, 0)

	// monitor m3 with pod informer
	podInformer := factory.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			if val, ok := pod.Labels["app"]; ok && val == "m3" {
				//if strings.HasPrefix(pod.ObjectMeta.Name, "m3") {
				fmt.Println("An m3 pod was added:", pod.ObjectMeta.Name)
				// we are going to ignore the first registered pod
				if !initialized {
					initialized = true
					fmt.Println("Initialized Pod Informer")
				} else {
					// close private and public so they get restarted
					m3Cancel()
					// restart the context
					m3PFCtx, m3Cancel = context.WithCancel(context.Background())
				}
			}
			// monitor nginx
			if strings.HasPrefix(pod.ObjectMeta.Name, "nginx") {
				fmt.Println("An nginx pod was added:", pod.ObjectMeta.Name)
				// we are going to ignore the first registered pod
				if !nginxInitialized {
					nginxInitialized = true
					fmt.Println("Initialized Pod Informer")
				} else {
					// close private and public so they get restarted
					nCancel()
					// restart the context
					nginxCtx, nCancel = context.WithCancel(context.Background())
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*v1.Pod)
			if strings.HasPrefix(pod.ObjectMeta.Name, "m3") {
				fmt.Println("An m3 pod was deleted:", pod.ObjectMeta.Name)
			}
		},
	})

	go podInformer.Run(doneCh)

	// monitor if a channel gets closed
	numTries := 0
OuterLoop:
	for {
		select {
		case <-publicCh:
			fmt.Println("Public port forward closed, restarting it after 2 seconds")
			numTries++
			time.Sleep(time.Second * 2)
			publicCh = servicePortForwardPort(m3PFCtx, "m3", "50051", color.FgYellow)
			// if more than 100 tries, probs the container is down, stop trying
			if numTries > 100 {
				break OuterLoop
			}
		case <-privateCh:
			fmt.Println("Private port forward closed, restarting it after 2 seconds")
			numTries++
			time.Sleep(time.Second * 2)
			privateCh = servicePortForwardPort(m3PFCtx, "m3", "50052", color.FgGreen)
			// if more than 100 tries, probs the container is down, stop trying
			if numTries > 100 {
				break OuterLoop
			}
		case <-portalBackendCh:
			fmt.Println("Portal port forward closed, restarting it after 2 seconds")
			numTries++
			time.Sleep(time.Second * 2)
			portalBackendCh = servicePortForwardPort(m3PFCtx, "m3-portal-backend", "5050", color.FgMagenta)
			// if more than 100 tries, probs the container is down, stop trying
			if numTries > 100 {
				break OuterLoop
			}
		case <-nginxCh:
			fmt.Println("Nginx port forward closed, restarting it after 2 seconds")
			numTries++
			time.Sleep(time.Second * 2)
			nginxCh = servicePortForwardPort(m3PFCtx, "nginx-resolver", "9000:80", color.FgCyan)
			// if more than 100 tries, probs the container is down, stop trying
			if numTries > 100 {
				break OuterLoop
			}
		case <-portalCh:
			fmt.Println("Portal Proxy port forward closed, restarting it after 2 seconds")
			numTries++
			time.Sleep(time.Second * 2)
			portalCh = servicePortForwardPort(m3PFCtx, "portal-proxy", "9080:80", color.FgBlue)
			// if more than 100 tries, probs the container is down, stop trying
			if numTries > 100 {
				break OuterLoop
			}
		case <-doneCh:
			break OuterLoop
		}

	}
	// about to exit
	return nil
}

// run the command inside a goroutine, return a channel that closes then the command dies
func servicePortForwardPort(ctx context.Context, service, port string, dcolor color.Attribute) chan interface{} {
	ch := make(chan interface{})
	go func() {
		defer close(ch)
		// service we are going to forward
		serviceName := fmt.Sprintf("service/%s", service)
		// command to run
		cmd := exec.CommandContext(ctx, "kubectl", "port-forward", serviceName, port)
		// prepare to capture the output
		var errStdout, errStderr error
		stdoutIn, _ := cmd.StdoutPipe()
		stderrIn, _ := cmd.StderrPipe()
		err := cmd.Start()
		if err != nil {
			log.Fatalf("cmd.Start() failed with '%s'\n", err)
		}

		// cmd.Wait() should be called only after we finish reading
		// from stdoutIn and stderrIn.
		// wg ensures that we finish
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			errStdout = copyAndCapture(stdoutIn, dcolor)
			wg.Done()
		}()

		errStderr = copyAndCapture(stderrIn, dcolor)

		wg.Wait()

		err = cmd.Wait()
		if err != nil {
			log.Printf("cmd.Run() failed with %s\n", err.Error())
			return
		}
		if errStdout != nil || errStderr != nil {
			log.Printf("failed to capture stdout or stderr\n")
			return
		}
		//outStr, errStr := string(stdout), string(stderr)
		//fmt.Printf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)
	}()
	return ch
}

// capture and print the output of the command
func copyAndCapture(r io.Reader, dcolor color.Attribute) error {
	var out []byte
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			theColor := color.New(dcolor)
			//_, err := w.Write(d)
			_, err := theColor.Print(string(d))

			if err != nil {
				return err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return err
		}
	}
}
