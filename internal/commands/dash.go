/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package commands

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	gocmd "github.com/go-cmd/cmd"
	"github.com/spf13/cobra"
)

var verboseLevel int

func RunOctantWithOptions(kubeConfigPath string, args []string) {
	// cobra command processing has already taken place and has processed and removed all flags and
	// their options from args. Since this is a temporary workaround it is not worth the effort to
	// extract all of these flags back out of cmd and reconstruct the original command line.  But
	// the most important options will be reconstructed and forwarded
	myArgs := args
	for i := 0; i < verboseLevel; i++ {
		myArgs = append(myArgs, "-v")
	}

	if len(kubeConfigPath) > 0 {
		myArgs = append(myArgs, "--kubeconfig", kubeConfigPath)
	}

	// The code in this function should be replaced by the body of the Run function in octant's
	// internal/commands/dash.go in order to launch the octant code directly as a function call. But
	// as of v0.9.1 octant still cannot be called directly due to incorrect leakage of private types
	// into its public interface. This is captured in issue 448
	// (https://github.com/vmware-tanzu/octant/issues/448). Specifically, the logger that is
	// required to be passed to dash.Run is an *internal* type! As a temporary workaround, launch a
	// separate instance of octant exec.Command will find the given command in the PATH.

	// golang does not have any direct support for executing a command and forwarding along its
	// stdout/stderr.  The only safe approach is to use a 3rd-party library, go-cmd, that takes
	// care of race conditions and other anomalies but still requires supplying functions to read
	// the data from a stream and explicitly write it to stderr or stdout  How stupid!
	// The code here is based on the example from go-cmd:
	// https://github.com/go-cmd/cmd/blob/master/examples/blocking-streaming/main.go

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	cmdOptions := gocmd.Options{
		Buffered:  false,
		Streaming: true,
	}

	command := gocmd.NewCmdOptions(cmdOptions, "octant", myArgs...)

	// print stdout and stderr lines streaming from command
	go func() {
		for {
			select {
			case line := <-command.Stdout:
				fmt.Println(line)
			case line := <-command.Stderr:
				fmt.Fprintln(os.Stderr, line)
			}
		}
	}()

	// Start the command to grab its channel but do not block yet
	statusChan := command.Start()

	// If the parent (airshipui) receives a signal to terminate, then kill the child octant process
	go func() {
		<-sigs
		stat := command.Status()
		if stat.PID > 0 {
			proc, err := os.FindProcess(stat.PID)
			if err == nil {
				err = proc.Kill()
				if err != nil {
					fmt.Printf("error trying to kill process: %s\n", err)
				}
			}
		}
	}()

	// run and wait for octant to exit
	status := <-statusChan

	// command has finished but wait for goroutine to print all lines
	for len(command.Stdout) > 0 || len(command.Stderr) > 0 {
		time.Sleep(10 * time.Millisecond)
	}

	if status.Exit != 0 {
		if status.Error != nil {
			fmt.Println(status.Error)
		}
		os.Exit(status.Exit)
	}
}

// some day this may need to get refactored if the options become transportable from external sources
func addDashboardFlags(cmd *cobra.Command) {
	cmd.Flags().SortFlags = true

	// octant specific flags
	cmd.Flags().StringP("context", "", "", "initial context")
	cmd.Flags().BoolP("disable-cluster-overview", "", false, "disable cluster overview")
	cmd.Flags().BoolP("enable-feature-applications", "", false, "enable applications feature")
	cmd.Flags().String("kubeconfig", "", "absolute path to kubeConfig file")
	cmd.Flags().StringP("namespace", "n", "", "initial namespace")
	cmd.Flags().StringP("plugin-path", "", "", "plugin path")
	cmd.Flags().BoolP("verbose", "v", false, "turn on debug logging")
	cmd.Flags().StringP("accepted-hosts", "", "", "accepted hosts list [DEV]")
	cmd.Flags().Float32P("client-qps", "", 200, "maximum QPS for client [DEV]")
	cmd.Flags().IntP("client-burst", "", 400, "maximum burst for client throttle [DEV]")
	cmd.Flags().BoolP("disable-open-browser", "", false, "disable automatic launching of the browser [DEV]")
	cmd.Flags().BoolP("enable-opencensus", "c", false, "enable open census [DEV]")
	cmd.Flags().IntP("klog-verbosity", "", 0, "klog verbosity level [DEV]")
	cmd.Flags().StringP("listener-addr", "", "", "listener address for the octant frontend [DEV]")
	cmd.Flags().StringP("local-content", "", "", "local content path [DEV]")
	cmd.Flags().StringP("proxy-frontend", "", "", "url to send frontend request to [DEV]")
	cmd.Flags().String("ui-url", "", "dashboard url [DEV]")
}
