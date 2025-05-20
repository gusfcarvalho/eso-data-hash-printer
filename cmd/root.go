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
	namespace      string
	labelSelector  = "reconcile.external-secrets.io/managed=true"
	hashAnnotation = "reconcile.external-secrets.io/data-hash"
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

		// List all secrets
		secrets, err := clientset.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: labelSelector})

		for _, secret := range secrets.Items {
			// Only update secrets having the hash annotation set
			currentHash, ok := secret.Annotations[hashAnnotation]
			if !ok {
				fmt.Printf("%s/%s doesn't have the hash annotation, skipping..\n", secret.Namespace, secret.Name)
				continue
			}

			// Calculate the new hash
			newHash := utils.ObjectHash(secret.Data)

			// Skip if hash is up-to-date
			if currentHash == newHash {
				fmt.Printf("%s/%s hash is up-to-date, skipping..\n", secret.Namespace, secret.Name)
				continue
			}

			// Update the hash
			secret.Annotations[hashAnnotation] = newHash
			_, err := clientset.CoreV1().Secrets(secret.Namespace).Update(context.TODO(), &secret, metav1.UpdateOptions{})
			if err != nil {
				fmt.Printf("Error updating secret %s/%s: %v\n", secret.Namespace, secret.Name, err)
				continue
			}

			fmt.Printf("Updated secret %s/%s\n", secret.Namespace, secret.Name)
		}
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
	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace of secrets to update (empty for all namespaces)")
}
