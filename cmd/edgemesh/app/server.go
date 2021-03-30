package app

import (
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/kubeedge/beehive/pkg/core"
	beehiveContext "github.com/kubeedge/beehive/pkg/core/context"
	"github.com/kubeedge/edgemesh/cmd/edgemesh/app/options"
	"github.com/kubeedge/edgemesh/pkg/apis/componentconfig/edgemesh/v1alpha1"
	"github.com/kubeedge/edgemesh/pkg/common/client"
	"github.com/kubeedge/edgemesh/pkg/common/dao"
	"github.com/kubeedge/edgemesh/pkg/common/informers"
	"github.com/kubeedge/edgemesh/pkg/controller"
	"github.com/kubeedge/edgemesh/pkg/networking"
)

func NewEdgeMeshCommand() *cobra.Command {
	opts := options.NewEdgeMeshOptions()
	cmd := &cobra.Command{
		Use: "edgemesh",
		Long: `EdgeMesh is an edge listener mesh that supports intra-LAN listener discovery. In the future, edge gateways,
cross-edge-cloud listener discovery, and cross-LAN listener discovery will be added. EdgeMesh will interconnect with Istio.
In the future, EdgeMesh will bring the advanced traffic governance and monitoring capabilities of Istio to Kubeedge.`,
		Run: func(cmd *cobra.Command, args []string) {
			config, err := opts.Config()
			if err != nil {
				klog.Fatal(err)
			}

			client.InitEdgeMeshClient(config.KubeAPIConfig)
			dao.InitDBConfig(config.DataBase.DriverName, config.DataBase.AliasName, config.DataBase.DataSource)
			gis := informers.GetInformersManager()
			registerModules(config)

			// start all modules
			core.StartModules()
			gis.Start(beehiveContext.Done())
			core.GracefulShutdown()
		},
	}

	return cmd
}

// registerModules register all the modules started in edgemesh
func registerModules(c *v1alpha1.EdgeMeshConfig) {
	controller.Register(c.Modules.Controller)
	networking.Register(c.Modules.Networking)
}
