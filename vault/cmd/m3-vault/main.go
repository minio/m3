package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/fatih/color"
	vapi "github.com/hashicorp/vault/api"
	"github.com/minio/minio/pkg/env"
)

func main() {

	fmt.Println("Starting vault development service")

	doneCh := make(chan struct{})

	// listen for kill sign to stop all the processes
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Println("Heard someone want us death x_x", sig)
			close(doneCh)
		}
	}()

	vaultServiceCh := startVaultService(color.FgYellow)

	err := <-vaultInitAndUnseal()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Vault is ready to use")

OuterLoop:
	for {
		select {
		case <-vaultServiceCh:
			fmt.Println("Public port forward closed, restarting it after 2 seconds")
			time.Sleep(time.Second * 2)
			vaultServiceCh = startVaultService(color.FgYellow)
		case <-doneCh:
			break OuterLoop
		}
	}
	return
}

func startVaultService(dcolor color.Attribute) chan interface{} {
	doneCh := make(chan interface{})
	go func() {
		defer close(doneCh)
		// command to run
		cmd := exec.Command("vault", "server", "-config", "vault-config.json")
		// prepare to capture the output
		var errStdout, errStderr error
		stdoutIn, _ := cmd.StdoutPipe()
		stderrIn, _ := cmd.StderrPipe()
		err := cmd.Start()
		if err != nil {
			log.Fatalf("cmd.Start() failed with '%s'\n", err)
			return
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

	}()

	return doneCh
}

func vaultInitAndUnseal() chan error {
	doneCh := make(chan error)
	go func() {
		defer close(doneCh)

		rootToken := ""
		address := "http://localhost:8200"
		client, err := isVaultReadyRetry(address)
		if err != nil {
			doneCh <- err
			return
		}
		secretShares := 5
		secretThreshold := 3
		if env.Get("SECRET_SHARES", "") != "" {
			val, err := strconv.Atoi(env.Get("SECRET_SHARES", "5"))
			if err != nil {
				log.Println(err)
			} else {
				secretShares = val
			}
		}
		if env.Get("SECRET_THRESHOLD", "") != "" {
			val, err := strconv.Atoi(env.Get("SECRET_THRESHOLD", "3"))
			if err != nil {
				log.Println(err)
			} else {
				secretThreshold = val
			}
		}
		initConfigs, err := client.Sys().Init(&vapi.InitRequest{
			SecretShares:    secretShares,
			SecretThreshold: secretThreshold,
		})
		if err != nil {
			doneCh <- err
			return
		}
		for _, key := range initConfigs.Keys {
			_, err := client.Sys().Unseal(key)
			if err != nil {
				doneCh <- err
				return
			}
		}
		rootToken = initConfigs.RootToken

		log.Println("Vault root token:", rootToken)
		client.SetToken(rootToken)
		health, err := client.Sys().Health()

		if err != nil {
			doneCh <- err
			return
		}

		if !health.Initialized {
			doneCh <- errors.New("vault is not initialized")
			return
		}

		if health.Sealed {
			doneCh <- errors.New("vault is sealed")
			return
		}

		err = client.Sys().EnableAuthWithOptions("approle", &vapi.EnableAuthOptions{Type: "approle"})
		if err != nil {
			doneCh <- err
			return
		}
		log.Println("Vault enabled auth approle successfully")
		err = client.Sys().Mount("kv", &vapi.MountInput{Type: "kv"})
		if err != nil {
			doneCh <- err
			return
		}
		log.Println("Vault enabled secrets kv successfully")
	}()
	return doneCh
}

func isVaultReadyRetry(address string) (*vapi.Client, error) {
	currentTries := 0
	totalRetries, _ := strconv.Atoi(env.Get("TOTAL_INIT_RETRIES", "5"))
	for {
		config := &vapi.Config{Address: address}
		client, err := vapi.NewClient(config)
		if err != nil {
			return client, err
		}
		healthResponse, err := client.Sys().Health()
		if err != nil {
			// we'll tolerate errors here, probably vault not responding
			log.Println(err)
		}
		if healthResponse != nil {
			log.Println("Vault started successfully")
			return client, nil
		}
		log.Println("Vault not ready, sleeping 2 seconds.")
		time.Sleep(time.Second * 2)
		currentTries++
		if currentTries >= totalRetries {
			return nil, errors.New("vault was never ready. Unable to complete configuration of the KMS")
		}
	}
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
