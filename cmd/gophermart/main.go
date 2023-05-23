package main

import (
	"errors"
	"flag"
	"github.com/VladimirMovsesyan/praktikum-gophermart/internal/configuration"
	"github.com/VladimirMovsesyan/praktikum-gophermart/internal/worker"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	flAddress    = flag.String("a", "", "Gophermart's address") // RUN_ADDRESS
	flDsn        = flag.String("d", "", "Database dsn")         // DATABASE_URI
	flAccAddress = flag.String("r", "", "Accrual's address")    // ACCRUAL_SYSTEM_ADDRESS
)

func main() {
	flag.Parse()

	config, err := configuration.NewConfiguration(flAddress, flDsn, flAccAddress)
	if err != nil {
		log.Println(err)
		return
	}

	go func() {
		if err := config.Server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	wp := worker.NewWorkerPool(8, config.AccAddress+"/api/orders/", config.Storage)
	wp.Run()
	defer wp.Stop()

	sig := <-signals

	log.Println("Got signal:", sig.String())
	if err := config.Server.Shutdown(context.Background()); err != nil {
		log.Println("HTTP server Shutdown:", err)
	}
}
