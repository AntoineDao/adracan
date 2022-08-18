/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

var (
	Namespace          string
	From               string
	To                 string
	Resource           string
	Name               string
	defaultConfigFlags = genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag().WithDiscoveryBurst(300).WithDiscoveryQPS(50.0)
)

func fetchNamespace() (string, error) {

	if Namespace != "" {
		return Namespace, nil
	}

	mainClient := getClientFactory(&From)
	Namespace, _, err := mainClient.ToRawKubeConfigLoader().Namespace()
	if err != nil {
		return Namespace, fmt.Errorf("Failed to fetch namespace from kubeconfig: %s", err)
	}
	return Namespace, nil
}

// contextCmd represents the context command
var contextCmd = &cobra.Command{
	Use:   "context --from CTX_1 --to CTX_2 TYPE NAME",
	Short: "Diff kubernetes resources accross kubectl contexts",
	Long:  ``,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		Resource = args[0]
		Name = args[1]

		objects := make([]runtime.Object, 2)

		namespace, err := fetchNamespace()

		if err != nil {
			fmt.Print(err)
			return
		}

		for i, ctx := range []string{From, To} {
			fmt.Printf("Fetching object for context: %s", ctx)
			fmt.Println()
			c := getClientFactory(&ctx)
			o, err := getObject(c, namespace, args...)
			if err != nil {
				fmt.Printf("Failed to fetch in \"%s\": %s", ctx, err)
				return
			}
			objects[i] = o
		}

		fromYaml, err := toRawYaml(objects[0])
		if err != nil {
			panic(err)
		}

		toYaml, err := toRawYaml(objects[1])
		if err != nil {
			panic(err)
		}

		fmt.Print(cmp.Diff(fromYaml, toYaml))
	},
}

func toRawYaml(object runtime.Object) (string, error) {
	d, err := yaml.Marshal(object)
	return fmt.Sprintf("%s", d), err
}

func getClientFactory(context *string) cmdutil.Factory {
	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag().WithDiscoveryBurst(300).WithDiscoveryQPS(50.0)
	kubeConfigFlags.Context = context
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	return cmdutil.NewFactory(matchVersionKubeConfigFlags)
}

func getObject(clientFrom cmdutil.Factory, namespace string, args ...string) (runtime.Object, error) {
	fn := resource.FilenameOptions{}

	r := clientFrom.NewBuilder().
		Unstructured().
		NamespaceParam(namespace).DefaultNamespace().AllNamespaces(false).
		FilenameParam(false, &fn).
		LabelSelectorParam("").
		FieldSelectorParam("").
		Subresource("").
		RequestChunksOf(500).
		ResourceTypeOrNameArgs(true, args...).
		ContinueOnError().
		Latest().
		Flatten().
		Do()

	if err := r.Err(); err != nil {
		return nil, fmt.Errorf("Failed to fetch resource %s, %s", args, err)
	}

	return r.Object()
}

func init() {
	rootCmd.AddCommand(contextCmd)

	contextCmd.Flags().StringVarP(&Namespace, "namespace", "n", "", "The kubernetes namespace to search for the resource in. Defaults to context 1's namespace.")
	contextCmd.Flags().StringVarP(&From, "from", "f", "", "[required] The context to compare from")
	contextCmd.Flags().StringVarP(&To, "to", "t", "", "[required] The context to compare to")

	contextCmd.MarkFlagRequired("from")
	contextCmd.MarkFlagRequired("to")
}
