package appx_test

import (
	"context"
	"fmt"
	"time"

	"github.com/RussellLuo/appx"
)

func Example() {
	// Typically located in `func init()` of package a.
	appx.MustRegister(appx.New("a").
		InitV2(func(ctx context.Context, lc appx.Lifecycle, apps map[string]*appx.App) (appx.Value, appx.CleanFunc, error) {
			fmt.Printf("Initializing app %q, which requires %d app\n", "a", len(apps))
			lc.Append(appx.Hook{
				OnStart: func(ctx context.Context) error {
					fmt.Println(`Starting app "a"`)
					return nil
				},
				OnStop: func(ctx context.Context) error {
					fmt.Println(`Stopping app "a"`)
					return nil
				},
			})
			return nil, func() error {
				fmt.Println(`Doing cleanup for app "a"`)
				return nil
			}, nil
		}))

	// Typically located in `func init()` of package b.
	appx.MustRegister(appx.New("b").
		Require("a").
		InitV2(func(ctx context.Context, lc appx.Lifecycle, apps map[string]*appx.App) (appx.Value, appx.CleanFunc, error) {
			fmt.Printf("Initializing app %q, which requires app %q\n", "b", apps["a"].Name)
			lc.Append(appx.Hook{
				OnStart: func(ctx context.Context) error {
					fmt.Println(`Starting app "b"`)
					return nil
				},
				OnStop: func(ctx context.Context) error {
					fmt.Println(`Stopping app "b"`)
					return nil
				},
			})
			return nil, func() error {
				fmt.Println(`Doing cleanup for app "b"`)
				return nil
			}, nil
		}))

	// Typically located in `func main()` of package main.
	if err := appx.Install(context.Background(), "b"); err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	defer appx.Uninstall()

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
	// Doing cleanup for app "b"
	// Doing cleanup for app "a"
}
