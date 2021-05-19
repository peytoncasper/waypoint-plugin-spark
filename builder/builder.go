package builder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
)

type BuildConfig struct {
	Type 	   string `hcl:"type"`
	Directory  string `hcl:"directory,optional"`
	OutputPath string `hcl:"output_path,optional"`
}

type Builder struct {
	config BuildConfig
}

// Implement Configurable
func (b *Builder) Config() (interface{}, error) {
	return &b.config, nil
}

// Implement ConfigurableNotify
func (b *Builder) ConfigSet(config interface{}) error {
	c, ok := config.(*BuildConfig)
	if !ok {
		// The Waypoint SDK should ensure this never gets hit
		return fmt.Errorf("Expected *BuildConfig as parameter")
	}

	if c.Type == "" {
		return fmt.Errorf("Spark Job type needs to be set")
	}

	if c.Directory == "" {
		fmt.Printf("Directory not set, defaulting to current directory")
		c.Directory = "."
	}

	return nil
}

// Implement Builder
func (b *Builder) BuildFunc() interface{} {
	// return a function which will be called by Waypoint
	return b.build
}

// A BuildFunc does not have a strict signature, you can define the parameters
// you need based on the Available parameters that the Waypoint SDK provides.
// Waypoint will automatically inject parameters as specified
// in the signature at run time.
//
// Available input parameters:
// - context.Context
// - *component.Source
// - *component.JobInfo
// - *component.DeploymentConfig
// - *datadir.Project
// - *datadir.App
// - *datadir.Component
// - hclog.Logger
// - terminal.UI
// - *component.LabelSet
//
// The output parameters for BuildFunc must be a Struct which can
// be serialzied to Protocol Buffers binary format and an error.
// This Output Value will be made available for other functions
// as an input parameter.
// If an error is returned, Waypoint stops the execution flow and
// returns an error to the user.
func (b *Builder) build(ctx context.Context, ui terminal.UI) (*Binary, error) {
	u := ui.Status()
	defer u.Close()
	u.Update("Building application")

	u.Step(terminal.InfoStyle, os.Getenv("DATABASE_URL"))

	outputArg := fmt.Sprintf("set assemblyOutputPath in assembly := new File(\"%s\")", b.config.OutputPath)

	c := exec.Command(
		"sbt",
		outputArg,
		"assembly",
	)
	c.Dir = b.config.Directory

	_, w := io.Pipe()
	defer w.Close()
	c.Stdout = w

	var b2 bytes.Buffer
	c.Stdout = &b2

	io.Copy(os.Stdout, &b2)


	err := c.Run()
	c.Wait()

	for _, line := range strings.Split(b2.String(), "\n") {
		u.Step(terminal.StatusOK, line)
	}

	if err != nil{
		return nil, err
	}

	return &Binary{ Type: b.config.Type, Path: b.config.Directory + "/" + b.config.OutputPath }, nil
}
