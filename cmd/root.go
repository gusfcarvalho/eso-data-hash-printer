/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/external-secrets/external-secrets/pkg/utils"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	secretName string
	namespace  string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "eso-data-hash-printer",
	Short: "calculats sha3 hashes for your secret",
	Long:  `calculates sha3 hashes for your secret`,
	Run: func(cmd *cobra.Command, args []string) {
		defaultKubeConfig := os.Getenv("KUBECONFIG")
		if defaultKubeConfig == "" {
			home := homedir.HomeDir()
			defaultKubeConfig = filepath.Join(home, ".kube", "config")
		}
		config, err := clientcmd.BuildConfigFromFlags("", defaultKubeConfig)
		if err != nil {
			fmt.Printf("Error building kubeconfig: %v\n", err)
			os.Exit(1)
		}

		// Create the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			fmt.Printf("Error creating Kubernetes client: %v\n", err)
			os.Exit(1)
		}

		// Get the secret
		secret, err := clientset.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("Error getting secret %s in namespace %s: %v\n", secretName, namespace, err)
			os.Exit(1)
		}

		// Calculate the hash
		hash := utils.ObjectHash(secret.Data)

		fmt.Printf("Secret: %s\n", secretName)
		fmt.Printf("Namespace: %s\n", namespace)
		fmt.Printf("ObjectHash: %s\n", hash)

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&secretName, "secret", "s", "", "Name of the Kubernetes secret to hash (required)")
	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace of the Kubernetes secret")
}
