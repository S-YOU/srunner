package main

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/spanner"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/kelseyhightower/envconfig"
	"github.com/sinmetal/gcpmetadata"
	"go.opencensus.io/trace"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

const Service = "srunner"

type EnvConfig struct {
	SpannerDatabase string `required:"true"`
	Goroutine       int    `default:"3"`
}

func main() {
	grpclog.Printf("Start GRPCLOG")

	var env EnvConfig
	if err := envconfig.Process("srunner", &env); err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("ENV_CONFIG %+v\n", env)

	project, err := gcpmetadata.GetProjectID()
	if err != nil {
		panic(err)
	}

	{
		exporter, err := stackdriver.NewExporter(stackdriver.Options{
			ProjectID: project,
		})
		if err != nil {
			panic(err)
		}
		trace.RegisterExporter(exporter)
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	}

	ctx := context.Background()

	// Need to specify scope for the specific service.
	tokenSource, err := DefaultTokenSourceWithProactiveCache(ctx, spanner.Scope)
	if err != nil {
		panic(err)
	}

	sc, err := createClient(ctx, env.SpannerDatabase,
		option.WithGRPCDialOption(
			grpc.WithTransportCredentials(&wrapTransportCredentials{
				TransportCredentials: credentials.NewClientTLSFromCert(nil, ""),
			}),
		),
		option.WithTokenSource(tokenSource),
	)
	if err != nil {
		panic(err)
	}
	ts := NewTweetStore(sc)

	endCh := make(chan error, 10)

	//goInsertTweet(ts, env.Goroutine, endCh)
	//goInsertTweetBenchmark(ts, env.Goroutine, endCh)
	//goUpdateTweet(ts, env.Goroutine, endCh)
	//goGetExitsTweet(ts, env.Goroutine, endCh)
	//goGetNotFoundTweet(ts, env.Goroutine, endCh)
	goGetTweet3Tables(ts, env.Goroutine, endCh)

	err = <-endCh
	fmt.Printf("BOMB %+v", err)
}
