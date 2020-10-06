package appx_test

import (
	"context"
	"fmt"
	"time"

	"github.com/RussellLuo/appx"
)

func Example() {
	// Typically located in `func init()` of package a.
	appx.MustRegister(
		appx.New("a").
			InitFunc(func(ctx appx.Context) error {
				name := ctx.App.Name
				fmt.Printf("Initializing app %q, which requires %d app\n", name, len(ctx.Required))

				ctx.Lifecycle.Append(appx.Hook{
					OnStart: func(ctx context.Context) error {
						fmt.Printf("Starting app %q\n", name)
						return nil
					},
					OnStop: func(ctx context.Context) error {
						fmt.Printf("Stopping app %q\n", name)
						return nil
					},
				})

				ctx.App.CleanFunc(func() error {
					fmt.Printf("Cleaning up app %q\n", name)
					return nil
				})
				return nil
			}),
	)

	// Typically located in `func init()` of package b.
	appx.MustRegister(
		appx.New("b").
			Require("a").
			InitFunc(func(ctx appx.Context) error {
				name := ctx.App.Name
				fmt.Printf("Initializing app %q, which requires app %q\n", name, ctx.Required["a"].Name)

				ctx.Lifecycle.Append(appx.Hook{
					OnStart: func(ctx context.Context) error {
						fmt.Printf("Starting app %q\n", name)
						return nil
					},
					OnStop: func(ctx context.Context) error {
						fmt.Printf("Stopping app %q\n", name)
						return nil
					},
				})

				ctx.App.CleanFunc(func() error {
					fmt.Printf("Cleaning up app %q\n", name)
					return nil
				})
				return nil
			}),
	)

	// Typically located in `func main()` of package main.
	if err := appx.Install(context.Background(), "b"); err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	defer appx.Uninstall()

	// In a typical scenario, we could just use appx.Run() here. Since we
	// don't want this example to run forever, we'll use the more explicit
	// Start and Stop.
	startCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := appx.Start(startCtx); err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	fmt.Println("Everything is running")

	stopCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	appx.Stop(stopCtx)

	// Output:
	// Initializing app "a", which requires 0 app
	// Initializing app "b", which requires app "a"
	// Starting app "a"
	// Starting app "b"
	// Everything is running
	// Stopping app "b"
	// Stopping app "a"
	// Cleaning up app "b"
	// Cleaning up app "a"
}
