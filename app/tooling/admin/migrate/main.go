package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jinchi2013/service/busniess/data/schema"
	"github.com/jinchi2013/service/busniess/sys/database"
)

func main() {
	err := migrate()

	if err != nil {
		fmt.Println(err)
	}
}

func migrate() error {
	cfg := database.Config{
		User:         "postgres",
		Password:     "postgres",
		Host:         "localhost",
		Name:         "postgres",
		MaxIdleConns: 0,
		MaxOpenConns: 0,
		DisableTLS:   true,
	}

	// open db by cfg
	db, err := database.Open(cfg)

	if err != nil {
		return fmt.Errorf("connnect databse: %w", err)
	}

	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := schema.Migrate(ctx, db); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	fmt.Println("migrations complete")

	return seed()
}

func seed() error {
	cfg := database.Config{
		User:         "postgres",
		Password:     "postgres",
		Host:         "localhost",
		Name:         "postgres",
		MaxIdleConns: 0,
		MaxOpenConns: 0,
		DisableTLS:   true,
	}

	// open db by cfg
	db, err := database.Open(cfg)

	if err != nil {
		return fmt.Errorf("connnect databse: %w", err)
	}

	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := schema.Seed(ctx, db); err != nil {
		return fmt.Errorf("seed database: %w", err)
	}

	fmt.Println("========= seed data complete =========")

	return nil
}
