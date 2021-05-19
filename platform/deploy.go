package platform

import (
	dataproc "cloud.google.com/go/dataproc/apiv1"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/peytoncasper/waypoint-plugin-spark/registry"
	"google.golang.org/api/option"
	dataprocpb "google.golang.org/genproto/googleapis/cloud/dataproc/v1"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

type DeployConfig struct {
	Region string "hcl:directory,optional"
	ProjectId string "hcl:project_id"
	ClusterName string "hcl:cluster_name"

	ClusterConfigPath string "hcl:cluster_config_path"
}

type Platform struct {
	config DeployConfig
}

// Implement Configurable
func (p *Platform) Config() (interface{}, error) {
	return &p.config, nil
}

// Implement ConfigurableNotify
func (p *Platform) ConfigSet(config interface{}) error {
	//c, ok := config.(*DeployConfig)
	//if !ok {
	//	// The Waypoint SDK should ensure this never gets hit
	//	return fmt.Errorf("Expected *DeployConfig as parameter")
	//}

	// validate the config
	//if c.Region == "" {
	//	return fmt.Errorf("Region must be set to a valid directory")
	//}

	return nil
}

// Implement Builder
func (p *Platform) DeployFunc() interface{} {
	// return a function which will be called by Waypoint
	return p.deploy
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

// In addition to default input parameters the registry.Artifact from the Build step
// can also be injected.
//
// The output parameters for BuildFunc must be a Struct which can
// be serialzied to Protocol Buffers binary format and an error.
// This Output Value will be made available for other functions
// as an input parameter.
// If an error is returned, Waypoint stops the execution flow and
// returns an error to the user.
func (b *Platform) deploy(ctx context.Context, ui terminal.UI, artifact *registry.Artifact) (*Deployment, error) {
	u := ui.Status()
	defer u.Close()
	u.Update("Deploy application")

	file, err := os.Open("cluster_config.json")
	if err != nil {
		u.Step(terminal.StatusError, "Error opening cluster config")
		return nil, err
	}

	byteValue, _ := ioutil.ReadAll(file)
	config := make(map[string]interface{})

	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		u.Step(terminal.StatusError, "Error parsing cluster config")
		return nil, err
	}

	projectId := config["project_id"].(string)
	region := config["region"].(string)
	clusterName := config["name"].(string)

	u.Step(terminal.InfoStyle, "Project Id: " + projectId)
	u.Step(terminal.InfoStyle, "Region: " + region)
	u.Step(terminal.InfoStyle, "Cluster Name: " + clusterName)

	endpoint := fmt.Sprintf("%s-dataproc.googleapis.com:443", region)
	jobClient, err := dataproc.NewJobControllerClient(ctx, option.WithEndpoint(endpoint))
	if err != nil {
		log.Fatalf("error creating the job client: %s\n", err)
	}


	//// Create the job config.
	submitJobReq := &dataprocpb.SubmitJobRequest{
		ProjectId: projectId,
		Region:    region,
		Job: &dataprocpb.Job{
			Placement: &dataprocpb.JobPlacement{
				ClusterName: clusterName,
			},
			TypeJob: &dataprocpb.Job_SparkJob{
				SparkJob: &dataprocpb.SparkJob{
					Driver: &dataprocpb.SparkJob_MainClass{
						MainClass: "org.apache.spark.examples.SparkPi",
					},
					JarFileUris: []string{"file:///usr/lib/spark/examples/jars/spark-examples.jar"},
					Args:        []string{"1000"},
				},
			},
		},
	}
	submitJobOp, err := jobClient.SubmitJobAsOperation(ctx, submitJobReq)
	if err != nil {
		return nil, err
	}

	submitJobResp, err := submitJobOp.Wait(ctx)
	if err != nil {
		return nil, err
	}
	//
	re := regexp.MustCompile("gs://(.+?)/(.+)")
	matches := re.FindStringSubmatch(submitJobResp.DriverOutputResourceUri)
	//
	if len(matches) < 3 {
		return nil, err
	}
	//
	//// Dataproc job output gets saved to a GCS bucket allocated to it.
	//storageClient, err := storage.NewClient(ctx)
	//if err != nil {
	//	return fmt.Errorf("error creating storage client: %v", err)
	//}
	//
	//obj := fmt.Sprintf("%s.000000000", matches[2])
	//reader, err := storageClient.Bucket(matches[1]).Object(obj).NewReader(ctx)
	//if err != nil {
	//	return fmt.Errorf("error reading job output: %v", err)
	//}
	//
	//defer reader.Close()
	//
	//body, err := ioutil.ReadAll(reader)
	//if err != nil {
	//	return fmt.Errorf("could not read output from Dataproc Job: %v", err)
	//}

	//fmt.Fprintf(w, "Job finished successfully: %s", body)


	return &Deployment{}, nil
}
