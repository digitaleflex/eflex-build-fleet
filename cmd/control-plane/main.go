package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"strings"

	"github.com/digitaleflex/eflex-build-fleet/internal/config"
	"github.com/digitaleflex/eflex-build-fleet/internal/controlplane"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg,err:=config.LoadControlPlane()
	if err!=nil { slog.Error("invalid configuration","error",err); os.Exit(1) }
	logger:=slog.New(slog.NewJSONHandler(os.Stdout,&slog.HandlerOptions{Level:cfg.LogLevel}))
	db,err:=sql.Open("pgx",cfg.DatabaseURL)
	if err!=nil { logger.Error("open database","error",err); os.Exit(1) }
	defer db.Close()
	db.SetMaxOpenConns(20); db.SetMaxIdleConns(5); db.SetConnMaxLifetime(30*time.Minute)
	ctx,stop:=signal.NotifyContext(context.Background(),os.Interrupt,syscall.SIGTERM); defer stop()
	pingCtx,cancel:=context.WithTimeout(ctx,5*time.Second)
	err=db.PingContext(pingCtx); cancel()
	if err!=nil { logger.Error("database unavailable","error",err); os.Exit(1) }
	srv:=controlplane.New(db,cfg.APIToken,logger)
	interval:=5*time.Second
	if v:=strings.TrimSpace(os.Getenv("SCHEDULER_INTERVAL")); v!="" { if d,err:=time.ParseDuration(v); err==nil { interval=d } }
	go srv.RunSchedulerLoop(ctx,interval,logger)
	if err:=srv.Run(ctx,cfg.Address); err!=nil { logger.Error("control plane stopped","error",err); os.Exit(1) }
}
