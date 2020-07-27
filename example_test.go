package appx_test

import (
	"context"
	"fmt"

	"github.com/RussellLuo/appx"
)

func Example() {
	// Typically located in `func init()` of package a.
	appx.MustRegister(appx.New("a").
		Init(func(ctx context.Context, apps map[string]*appx.App) (appx.Value, appx.CleanFunc, error) {
			fmt.Printf("Initializing app %q, which requires %d app\n", "a", len(apps))
			return nil, func() error {
				fmt.Println(`Doing cleanup for app "a"`)
				return nil
			}, nil
		}))

	// Typically located in `func init()` of package b.
	appx.MustRegister(appx.New("b").
		Require("a").
		Init(func(ctx context.Context, apps map[string]*appx.App) (appx.Value, appx.CleanFunc, error) {
			fmt.Printf("Initializing app %q, which requires app %q\n", "b", apps["a"].Name)
			return nil, func() error {
				fmt.Println(`Doing cleanup for app "b"`)
				return nil
			}, nil
		}))

	// Typically located in `func main()` of package main.
	if err := appx.Install(context.Background(), "b"); err != nil {
		fmt.Printf("err: %v\n", err)
	}
	if err := appx.Uninstall("a", "b"); err != nil {
		fmt.Printf("err: %v\n", err)
	}

	// Output:
	// Initializing app "a", which requires 0 app
	// Initializing app "b", which requires app "a"
	// Doing cleanup for app "a"
	// Doing cleanup for app "b"
}
