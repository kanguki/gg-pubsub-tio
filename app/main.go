package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func main() {
	err := newRootCmd().Execute()
	if err != nil {
		log.Fatalf("newSubscribeCmd.Execute: %v", err)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "root",
		Short: "root command",
	}
	cmd.AddCommand(newPublishCmd())
	cmd.AddCommand(newSubscribeCmd())
	cmd.AddCommand(newSetUpCmd())
	return cmd
}

func newSubscribeCmd() *cobra.Command {
	var (
		googleCloudProject string
		topic              string
		subscription       string
		parallel           int
	)
	cmd := &cobra.Command{
		Use:   "subscribe",
		Short: "listen to messages in google cloud pubsub and print it out",
		Long:  "listen to messages in google cloud pubsub and print it out",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Println("running")
			shutdown := make(chan os.Signal, 1)
			signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
			errorsChan := make(chan error, 1)
			clients := []*pubsub.Client{}
			ctx := context.Background()
			receivedMsgs := make(chan string, 1)
			// spin up N listeners on the same subscription
			for i := 0; i < parallel; i++ {
				go func(listenerID int) {
					client, err := pubsub.NewClient(ctx, googleCloudProject)
					if err != nil {
						errorsChan <- fmt.Errorf("pubsub.NewClient: %w", err)
						return
					}
					clients = append(clients, client)
					sub := client.Subscription(subscription)
					err = sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
						receivedMsgs <- fmt.Sprintf("+++++ %d received\n%s\n+++++\n", listenerID, m.Data)
						m.Ack()
					})
					errorsChan <- fmt.Errorf("sub.Receive: %w", err)
				}(i)
			}
			cnt := 0
			for {
				select {
				case msg := <-receivedMsgs:
					cnt++
					fmt.Printf("%d: %s", cnt, msg)
				case err := <-errorsChan:
					return err
				case sig := <-shutdown:
					log.Printf("receive terminate signal: %v", sig)
					for _, client := range clients {
						err := client.Close()
						if err != nil {
							log.Printf("client.Close: %v", err)
						}
					}
					err := <-errorsChan
					return err
				}
			}
		},
	}
	cmd.Flags().StringVarP(&googleCloudProject, "gg-project", "p", "", "google cloud project")
	cmd.Flags().StringVarP(&topic, "topic", "t", "", "pubsub topic")
	cmd.Flags().StringVarP(&subscription, "subscription", "s", "", "subscription")
	cmd.Flags().IntVar(&parallel, "parallel", 1, "number of parallel listeners")
	cmd.MarkFlagRequired("gg-project")
	cmd.MarkFlagRequired("topic")
	cmd.MarkFlagRequired("subscription")
	return cmd
}

func newPublishCmd() *cobra.Command {
	var (
		googleCloudProject string
		topic              string
		subscription       string
		batch              int
		times              int
	)
	cmd := &cobra.Command{
		Use:   "publish",
		Short: "publish messages to google pubsub",
		Long:  "publish messages to google pubsub",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Println("running")
			ctx := context.Background()
			client, err := pubsub.NewClient(ctx, googleCloudProject)
			if err != nil {
				return fmt.Errorf("pubsub.NewClient: %w", err)
			}

			key := time.Now().Unix()
			totalMsgCount := times * batch
			messages := make(chan string, batch)
			go func() {
				for i := 0; i < totalMsgCount; i++ {
					messages <- fmt.Sprintf("%d-%d", i, key)
				}
			}()

			topic := client.Topic(topic)
			eg, eCtx := errgroup.WithContext(ctx)
			for i := 0; i < totalMsgCount; i++ {
				eg.Go(func() error {
					message := <-messages
					result := topic.Publish(ctx, &pubsub.Message{Data: []byte(message)})
					id, err := result.Get(eCtx)
					if err != nil {
						return fmt.Errorf("result.Get: %v", err)
					}
					log.Printf("published msgid: %s", id)
					return nil
				})
			}
			err = eg.Wait()
			if err != nil {
				log.Fatalf("eg.Wait: %v", err)
			}
			err = client.Close()
			if err != nil {
				log.Fatalf("client.Close: %v", err)
			}
			log.Printf("done sending %d messages key %d", totalMsgCount, key)
			return nil
		},
	}
	cmd.Flags().StringVarP(&googleCloudProject, "gg-project", "p", "", "google cloud project")
	cmd.Flags().StringVarP(&topic, "topic", "t", "", "pubsub topic")
	cmd.Flags().StringVarP(&subscription, "subscription", "s", "", "subscription")
	cmd.Flags().IntVar(&batch, "batch", 10, "number of concurrent messages to send")
	cmd.Flags().IntVar(&times, "times", 5, "number of batches to send")
	cmd.MarkFlagRequired("gg-project")
	cmd.MarkFlagRequired("topic")
	cmd.MarkFlagRequired("subscription")
	return cmd
}

func newSetUpCmd() *cobra.Command {
	var (
		googleCloudProject string
		topic              string
		subscriptions      []string
	)
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "create topic and subscriptions",
		Long:  "create topic and subscriptions",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := pubsub.NewClient(ctx, googleCloudProject)
			if err != nil {
				return fmt.Errorf("pubsub.NewClient: %w", err)
			}
			tp, tErr := client.CreateTopic(ctx, topic)
			if tErr != nil {
				return fmt.Errorf("client.CreateTopic: %w", tErr)
			}
			for _, s := range subscriptions {
				_, sErr := client.CreateSubscription(ctx, s, pubsub.SubscriptionConfig{
					Topic:       tp,
					AckDeadline: 10 * time.Second,
				})
				if sErr != nil {
					return fmt.Errorf("client.CreateSubscription: %w", sErr)
				}
			}
			log.Printf("done creating topic %s , subscriptions %v in project %s", topic, subscriptions, googleCloudProject)
			return nil
		},
	}
	cmd.Flags().StringVarP(&googleCloudProject, "gg-project", "p", "", "google cloud project")
	cmd.Flags().StringVarP(&topic, "topic", "t", "", "pubsub topic")
	cmd.Flags().StringArrayVarP(&subscriptions, "subscription", "s", nil, "subscriptions to create")
	cmd.MarkFlagRequired("gg-project")
	cmd.MarkFlagRequired("topic")
	cmd.MarkFlagRequired("subscription")
	return cmd
}
