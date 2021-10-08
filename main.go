package main

import (
	"github.com/takescoop/service-connect/cmd"
)

// func run() error {
// 	kc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
// 		clientcmd.NewDefaultClientConfigLoadingRules(),
// 		&clientcmd.ConfigOverrides{},
// 	)

// 	kConfig, err := kc.ClientConfig()
// 	if err != nil {
// 		return err
// 	}

// 	clientset, err := kubernetes.NewForConfig(kConfig)
// 	if err != nil {
// 		return err
// 	}

// 	ctx := context.Background()

// 	svc, err := clientset.CoreV1().Services("release").Get(ctx, "harbormaster-db-proxy", v1.GetOptions{})
// 	if err != nil {
// 		return err
// 	}

// 	annoType, ok := svc.Annotations["aws-con.service.kubernetes.io/type"]
// 	if !ok {
// 		return fmt.Errorf("aws-con.service.kubernetes.io/type not found")
// 	}

// 	config, err := config.LoadDefaultConfig(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	switch annoType {
// 	case "rds-iam":
// 		annoMeta, ok := svc.Annotations["aws-con.service.kubernetes.io/meta"]
// 		if !ok {
// 			return fmt.Errorf("aws-con.service.kubernetes.io/meta not found")
// 		}

// 		var meta rds.Meta
// 		err := json.Unmarshal([]byte(annoMeta), &meta)
// 		if err != nil {
// 			return err
// 		}

// 		user := "iam-read"

// 		c := rds.New(config.Region, config)
// 		token, err := c.GetAuthToken(ctx, meta, user, config)
// 		if err != nil {
// 			return err
// 		}

// 		localPort := 8080

// 		conn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", url.QueryEscape(user), url.QueryEscape(token), "localhost", localPort, meta.DBName)

// 		fmt.Println(conn)
// 	}

// 	return nil
// }

func main() {
	cmd.Execute()
}
