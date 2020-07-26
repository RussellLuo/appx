package appx

import (
	"context"
	"fmt"
)

func Example() {
	MustRegister(New("a").
		Init(func(ctx context.Context, apps map[string]*App) (Value, CleanFunc, error) {
			fmt.Printf("Initializing app %q, which requires %d app.\n", "a", len(apps))
			return nil, func() error {
				fmt.Printf("Doing cleanup for app %q.\n", "a")
				return nil
			}, nil
		}))

	MustRegister(New("b").
		Init(func(ctx context.Context, apps map[string]*App) (Value, CleanFunc, error) {
			fmt.Printf("Initializing app %q, which requires %d app.\n", "b", len(apps))
			return nil, func() error {
				fmt.Printf("Doing cleanup for app %q.\n", "b")
				return nil
			}, nil
		}))

	MustRegister(New("c").
		Require("a").
		Init(func(ctx context.Context, apps map[string]*App) (Value, CleanFunc, error) {
			fmt.Printf("Initializing app %q, which requires app %q.\n", "c", apps["a"].Name)
			return nil, func() error {
				fmt.Printf("Doing cleanup for app %q.\n", "c")
				return nil
			}, nil
		}))

	MustRegister(New("d").
		Require("a", "b").
		Init(func(ctx context.Context, apps map[string]*App) (Value, CleanFunc, error) {
			fmt.Printf("Initializing app %q, which requires app %q and %q.\n", "d", apps["a"].Name, apps["b"].Name)
			return nil, func() error {
				fmt.Printf("Doing cleanup for app %q.\n", "d")
				return nil
			}, nil
		}))

	if err := Install(context.Background()); err != nil {
		fmt.Printf("err: %v\n", err)
	}

	if err := Uninstall(); err != nil {
		fmt.Printf("err: %v\n", err)
	}
}
